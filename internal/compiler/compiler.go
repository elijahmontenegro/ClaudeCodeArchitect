package compiler

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

// Report is a compile's Markdown plus the problems found along the way.
type Report struct {
	Markdown string

	// Warnings are asciidoctor's warning and error lines, such as a missing
	// include, a reference to an undefined attribute, or a section out of
	// sequence. Asciidoctor exits 0 on all of them.
	Warnings []string

	// BrokenRefs are cross-reference targets that match no id in the
	// document. Only a full compile checks them, since a fragment's
	// references can point outside it.
	BrokenRefs []string
}

// Compile compiles the full document to Markdown
func Compile(entryPath string) (string, error) {
	report, err := CompileWithReport(entryPath)
	return report.Markdown, err
}

// CompileWithReport compiles the full document to Markdown and reports what's
// wrong with it.
func CompileWithReport(entryPath string) (Report, error) {
	absPath, err := filepath.Abs(entryPath)
	if err != nil {
		return Report{}, fmt.Errorf("failed to resolve path: %v", err)
	}

	// Embedded output (-s) is the document's content without the page around
	// it: no <head> title to repeat, no "Last updated" footer to churn in diffs.
	// showtitle keeps the document title as the first heading.
	html, warnings, err := runAsciidoctor(nil, "-s", "-a", "showtitle", "-o", "-", absPath)
	if err != nil {
		return Report{}, err
	}

	markdown, err := HTMLToMarkdown(html)
	if err != nil {
		return Report{}, err
	}
	return Report{Markdown: markdown, Warnings: warnings, BrokenRefs: BrokenReferences(html)}, nil
}

// CompileContent compiles AsciiDoc content to Markdown, resolving its
// includes against baseDir, which is also the include jail. Used for
// sections and single files.
func CompileContent(content string, baseDir string) (Report, error) {
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return Report{}, fmt.Errorf("failed to resolve base dir: %v", err)
	}

	// -s gives embedded output, -B sets the base directory for includes,
	// and the final - reads from stdin
	html, warnings, err := runAsciidoctor(strings.NewReader(content), "-s", "-B", absBaseDir, "-o", "-", "-")
	if err != nil {
		return Report{}, err
	}

	markdown, err := HTMLToMarkdown(html)
	if err != nil {
		return Report{}, err
	}
	return Report{Markdown: markdown, Warnings: warnings}, nil
}

// HTMLToMarkdown converts HTML to Markdown: CommonMark, plus pipe tables so
// a table keeps its rows and columns.
func HTMLToMarkdown(html string) (string, error) {
	conv := converter.NewConverter(converter.WithPlugins(
		base.NewBasePlugin(),
		commonmark.NewCommonmarkPlugin(),
		table.NewTablePlugin(),
	))
	markdown, err := conv.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("failed to convert to markdown: %v", err)
	}

	return markdown, nil
}

// IsAsciidoctorAvailable checks if asciidoctor CLI is installed
func IsAsciidoctorAvailable() bool {
	_, err := exec.LookPath("asciidoctor")
	return err == nil
}

// runAsciidoctor converts to HTML5 and returns asciidoctor's warning and error
// lines alongside the output.
//
// It runs in SAFE mode, so includes are confined to the base directory: the
// directory of the document's entry point. The CLI's default UNSAFE mode lets
// a document include any file on disk, so a repo an agent is asked to read
// could pull, say, a private key into the Markdown the agent then sees.
// An include outside the directory becomes a warning instead.
//
// References to undefined attributes are reported unless the document sets
// attribute-missing itself: the trailing @ makes the setting one the
// document can override.
func runAsciidoctor(stdin io.Reader, args ...string) (string, []string, error) {
	if !IsAsciidoctorAvailable() {
		return "", nil, fmt.Errorf("asciidoctor not found in PATH\n\nInstall with: gem install asciidoctor\nOr: brew install asciidoctor\nOr: npm install -g asciidoctor")
	}

	args = append([]string{"-S", "safe", "-b", "html5", "-a", "attribute-missing=warn@"}, args...)
	cmd := exec.Command("asciidoctor", args...)
	if stdin != nil {
		cmd.Stdin = stdin
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", nil, fmt.Errorf("failed to compile: %v\n%s", err, stderr.String())
	}

	return stdout.String(), parseDiagnostics(stderr.String()), nil
}

var diagnosticPattern = regexp.MustCompile(`^(?:asciidoctor: )?(WARNING|ERROR|FATAL): (.+)$`)

// parseDiagnostics picks asciidoctor's warning and error lines out of its
// stderr, as "LEVEL: message".
func parseDiagnostics(stderr string) []string {
	var diagnostics []string
	for _, line := range strings.Split(stderr, "\n") {
		if m := diagnosticPattern.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			diagnostics = append(diagnostics, m[1]+": "+m[2])
		}
	}
	return diagnostics
}

var (
	idPattern       = regexp.MustCompile(`\sid="([^"]+)"`)
	fragmentPattern = regexp.MustCompile(`\shref="#([^"]*)"`)
)

// BrokenReferences returns the targets of in-document links (href="#target")
// that match no element id in the HTML, each once, in order of first use.
// Asciidoctor renders an unresolved <<target>> as such a link.
func BrokenReferences(html string) []string {
	ids := map[string]bool{}
	for _, m := range idPattern.FindAllStringSubmatch(html, -1) {
		ids[m[1]] = true
	}

	seen := map[string]bool{}
	var broken []string
	for _, m := range fragmentPattern.FindAllStringSubmatch(html, -1) {
		target := m[1]
		if target == "" || ids[target] || seen[target] {
			continue
		}
		seen[target] = true
		broken = append(broken, target)
	}
	return broken
}
