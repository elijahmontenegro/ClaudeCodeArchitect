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

// Size thresholds for warnings
const (
	SizeWarningThreshold = 20 * 1024 // 20KB - warn user
	SizeLargeThreshold   = 50 * 1024 // 50KB - strongly warn
)

// ValidationOptions controls validation behavior
type ValidationOptions struct {
	SkipConfirm bool   // --yes flag: skip size confirmation
	Ultra       bool   // --ultra flag: multi-run validation with synthesis
	JSON        bool   // --json flag: output JSON (for CI)
	Profile     string // from .cca.yaml; semantic validation runs for "architecture" only
}

// TemplateData holds data passed to prompt templates
type TemplateData struct {
	CompiledSpec string
	Run1         string
	Run2         string
	Run3         string
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

// runClaudeValidationQuiet runs validation without spinner (for parallel runs)
func runClaudeValidationQuiet(ctx context.Context, compiledSpec string, output io.Writer) error {
	prompt, err := RenderPrompt("validate", TemplateData{
		CompiledSpec: compiledSpec,
	})
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

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

// RunClaudeValidation shells out to claude CLI for semantic validation
// It streams output directly to the provided writer
func RunClaudeValidation(compiledSpec string, output io.Writer) error {
	// Render the prompt
	prompt, err := RenderPrompt("validate", TemplateData{
		CompiledSpec: compiledSpec,
	})
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	// Check if claude CLI is available
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude CLI not found in PATH - install from https://claude.ai/code")
	}

	// Run claude with prompt via stdin (avoids command line length limits)
	// Using --print for non-interactive mode
	cmd := exec.Command("claude", "--print", "--no-session-persistence")
	cmd.Stdin = strings.NewReader(prompt)

	// Capture stdout to buffer while showing spinner
	var resultBuf bytes.Buffer
	cmd.Stdout = &resultBuf
	cmd.Stderr = os.Stderr

	// Show spinner only if stderr is a terminal
	showSpinner := isTTY()
	if showSpinner {
		fmt.Fprint(os.Stderr, "Running Claude validation ")
	}

	done := make(chan bool)
	if showSpinner {
		go func() {
			spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			i := 0
			for {
				select {
				case <-done:
					return
				default:
					fmt.Fprintf(os.Stderr, "\rRunning Claude validation %s", spinner[i%len(spinner)])
					i++
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}

	err = cmd.Run()
	if showSpinner {
		done <- true
		fmt.Fprint(os.Stderr, "\r                                    \r")
	}

	if err != nil {
		return fmt.Errorf("claude CLI failed: %w", err)
	}

	// Write the captured output
	fmt.Fprint(output, resultBuf.String())

	return nil
}

// RunClaudeValidationToString runs validation and returns result as string
func RunClaudeValidationToString(compiledSpec string) (string, error) {
	var buf bytes.Buffer
	if err := RunClaudeValidation(compiledSpec, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
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

// RunUltraValidation runs validation 3 times in parallel and synthesizes results
func RunUltraValidation(compiledSpec string, output io.Writer) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Run 3 validations concurrently (silent)
	type result struct {
		output string
		err    error
		index  int
	}
	results := make(chan result, 3)

	for i := 0; i < 3; i++ {
		go func(idx int) {
			var buf bytes.Buffer
			err := runClaudeValidationQuiet(ctx, compiledSpec, &buf)
			results <- result{output: buf.String(), err: err, index: idx}
		}(i)
	}

	// Show spinner only if stderr is a terminal
	showSpinner := isTTY()
	done := make(chan bool)

	if showSpinner {
		go func() {
			spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			i := 0
			for {
				select {
				case <-done:
					return
				case <-ctx.Done():
					return
				default:
					fmt.Fprintf(os.Stderr, "\rRunning Claude validation %s", spinner[i%len(spinner)])
					i++
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}

	// Collect results
	runs := make([]string, 3)
	for i := 0; i < 3; i++ {
		r := <-results
		if r.err != nil {
			if showSpinner {
				done <- true
			}
			if ctx.Err() != nil {
				return fmt.Errorf("cancelled")
			}
			return fmt.Errorf("validation failed: %w", r.err)
		}
		runs[r.index] = r.output
	}

	if showSpinner {
		done <- true
		fmt.Fprint(os.Stderr, "\r                                    \r")
	}

	// Synthesize results
	synthesisPrompt, err := RenderPrompt("synthesize", TemplateData{
		Run1: runs[0],
		Run2: runs[1],
		Run3: runs[2],
	})
	if err != nil {
		return fmt.Errorf("failed to render synthesis prompt: %w", err)
	}

	cmd := exec.CommandContext(ctx, "claude", "--print", "--no-session-persistence")
	cmd.Stdin = strings.NewReader(synthesisPrompt)

	var resultBuf bytes.Buffer
	cmd.Stdout = &resultBuf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("cancelled")
		}
		return fmt.Errorf("synthesis failed: %w", err)
	}

	fmt.Fprint(output, resultBuf.String())
	return nil
}

// CheckSpecSize checks spec size and prompts for confirmation if large
// Returns true if should proceed, false if user cancelled
func CheckSpecSize(compiledSpec string, opts ValidationOptions, output io.Writer) (bool, error) {
	size := len(compiledSpec)

	if size < SizeWarningThreshold {
		return true, nil // Small spec, no warning needed
	}

	// Calculate approximate tokens (rough: 1 token ≈ 4 chars)
	approxTokens := size / 4

	if opts.Ultra {
		approxTokens *= 4 // 3 validations + 1 synthesis
	}

	if size >= SizeLargeThreshold {
		fmt.Fprintf(output, "Large spec detected: %dKB (~%d tokens)\n", size/1024, approxTokens)
		if opts.Ultra {
			fmt.Fprintf(output, "  Ultra mode: 3 validation runs + synthesis\n")
		}
		fmt.Fprintf(output, "  This will use significant Claude API capacity.\n\n")
	} else if opts.Ultra {
		fmt.Fprintf(output, "Ultra mode: %dKB (~%d tokens with 3 runs + synthesis)\n\n", size/1024, approxTokens)
	} else {
		fmt.Fprintf(output, "Spec size: %dKB (~%d tokens)\n\n", size/1024, approxTokens)
	}

	if opts.SkipConfirm {
		return true, nil
	}

	// Interactive confirmation
	fmt.Fprint(output, "Proceed with Claude validation? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}
