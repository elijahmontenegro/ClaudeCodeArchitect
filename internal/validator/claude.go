package validator

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"text/template"
	"time"

	"golang.org/x/term"
)

//go:embed prompts/*.tmpl
var promptTemplates embed.FS

// defaultChecklist is what the review checks when the docs supply no
// checklist of their own.
//
//go:embed prompts/checklist.md
var defaultChecklist string

// Size thresholds for warnings
const (
	SizeWarningThreshold = 20 * 1024 // 20KB - warn user
	SizeLargeThreshold   = 50 * 1024 // 50KB - strongly warn
)

// ValidationOptions controls validation behavior
type ValidationOptions struct {
	SkipConfirm bool   // --yes: skip the size confirmation
	Ultra       bool   // --ultra: three independent reviews, then a synthesis
	JSON        bool   // --json: output JSON (for CI, with --quick)
	Checklist   string // the docs' own checklist file, from .cca.yaml; empty means the default
}

// TemplateData holds data passed to prompt templates
type TemplateData struct {
	CompiledDocs string
	Checklist    string
	Run1         string
	Run2         string
	Run3         string
}

// DefaultChecklist returns the checklist the review uses when the docs
// supply none.
func DefaultChecklist() string {
	return defaultChecklist
}

// LoadChecklist returns the content of the checklist file at path, or the
// default checklist when path is empty.
func LoadChecklist(path string) (string, error) {
	if path == "" {
		return defaultChecklist, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read checklist: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return "", fmt.Errorf("checklist %s is empty", path)
	}
	return string(data), nil
}

// LoadPromptTemplate loads and parses a prompt template
func LoadPromptTemplate(name string) (*template.Template, error) {
	return template.ParseFS(promptTemplates, "prompts/"+name+".tmpl")
}

// RenderPrompt renders a prompt template with data
func RenderPrompt(templateName string, data TemplateData) (string, error) {
	tmpl, err := LoadPromptTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}

// runClaude sends a prompt to the claude CLI on stdin, which avoids command
// line length limits, and writes its answer to output.
func runClaude(ctx context.Context, prompt string, output io.Writer) error {
	cmd := exec.CommandContext(ctx, "claude", "--print", "--no-session-persistence")
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("claude CLI failed: %w", err)
	}
	return nil
}

// reviewPrompt renders the review prompt for compiled docs and a checklist.
func reviewPrompt(compiledDocs, checklist string) (string, error) {
	prompt, err := RenderPrompt("validate", TemplateData{
		CompiledDocs: compiledDocs,
		Checklist:    checklist,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render prompt: %w", err)
	}
	return prompt, nil
}

// RunReview reviews compiled docs against a checklist with Claude and writes
// the findings to output.
func RunReview(compiledDocs, checklist string, output io.Writer) error {
	prompt, err := reviewPrompt(compiledDocs, checklist)
	if err != nil {
		return err
	}

	if !IsClaudeAvailable() {
		return fmt.Errorf("claude CLI not found in PATH - install from https://claude.ai/code")
	}

	stop := startSpinner(context.Background())
	var resultBuf bytes.Buffer
	err = runClaude(context.Background(), prompt, &resultBuf)
	stop()
	if err != nil {
		return err
	}

	fmt.Fprint(output, resultBuf.String())
	return nil
}

// RunUltraReview runs three independent reviews in parallel, then has Claude
// merge them, rating each finding by how many reviews found it.
func RunUltraReview(compiledDocs, checklist string, output io.Writer) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	prompt, err := reviewPrompt(compiledDocs, checklist)
	if err != nil {
		return err
	}

	type result struct {
		output string
		err    error
		index  int
	}
	results := make(chan result, 3)
	for i := 0; i < 3; i++ {
		go func(idx int) {
			var buf bytes.Buffer
			err := runClaude(ctx, prompt, &buf)
			results <- result{output: buf.String(), err: err, index: idx}
		}(i)
	}

	stop := startSpinner(ctx)
	runs := make([]string, 3)
	for i := 0; i < 3; i++ {
		r := <-results
		if r.err != nil {
			stop()
			if ctx.Err() != nil {
				return fmt.Errorf("cancelled")
			}
			return fmt.Errorf("review failed: %w", r.err)
		}
		runs[r.index] = r.output
	}
	stop()

	synthesisPrompt, err := RenderPrompt("synthesize", TemplateData{
		Run1: runs[0],
		Run2: runs[1],
		Run3: runs[2],
	})
	if err != nil {
		return fmt.Errorf("failed to render synthesis prompt: %w", err)
	}

	var resultBuf bytes.Buffer
	if err := runClaude(ctx, synthesisPrompt, &resultBuf); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("cancelled")
		}
		return fmt.Errorf("synthesis failed: %w", err)
	}

	fmt.Fprint(output, resultBuf.String())
	return nil
}

// startSpinner shows a spinner on stderr while Claude works, if stderr is a
// terminal, and returns the function that stops and clears it.
func startSpinner(ctx context.Context) func() {
	if !isTTY() {
		return func() {}
	}

	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			default:
				fmt.Fprintf(os.Stderr, "\rReviewing with Claude %s", frames[i%len(frames)])
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	return func() {
		close(done)
		<-stopped
		fmt.Fprint(os.Stderr, "\r                                    \r")
	}
}

// IsClaudeAvailable checks if claude CLI is installed
func IsClaudeAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// isTTY checks if stderr is a terminal (for spinner support)
func isTTY() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// CheckSize reports the size of the compiled docs and, when they're large,
// asks before sending them to Claude. It returns false if the user declines.
func CheckSize(compiledDocs string, opts ValidationOptions, output io.Writer) (bool, error) {
	size := len(compiledDocs)

	if size < SizeWarningThreshold {
		return true, nil
	}

	// Approximate tokens (roughly 4 characters each)
	approxTokens := size / 4

	if opts.Ultra {
		approxTokens *= 4 // 3 reviews + 1 synthesis
	}

	if size >= SizeLargeThreshold {
		fmt.Fprintf(output, "Large docs: %dKB (~%d tokens)\n", size/1024, approxTokens)
		if opts.Ultra {
			fmt.Fprintf(output, "  Ultra mode: 3 reviews + synthesis\n")
		}
		fmt.Fprintf(output, "  This will use significant Claude capacity.\n\n")
	} else if opts.Ultra {
		fmt.Fprintf(output, "Ultra mode: %dKB (~%d tokens with 3 reviews + synthesis)\n\n", size/1024, approxTokens)
	} else {
		fmt.Fprintf(output, "Docs size: %dKB (~%d tokens)\n\n", size/1024, approxTokens)
	}

	if opts.SkipConfirm {
		return true, nil
	}

	fmt.Fprint(output, "Proceed with the Claude review? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}
