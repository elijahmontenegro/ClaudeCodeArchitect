package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/compiler"
	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/parser"
)

// StructuralCheck represents a fast pre-flight check
type StructuralCheck struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

// RunStructuralChecks performs fast pre-flight validation
// These checks don't require Claude - they're instant Go checks
func RunStructuralChecks(manifestPath string) ([]StructuralCheck, error) {
	var checks []StructuralCheck

	// Check 1: Docs compile
	compileCheck := StructuralCheck{
		ID:   "compiles",
		Name: "Docs compile",
	}
	report, err := compiler.CompileWithReport(manifestPath)
	if err != nil {
		compileCheck.Passed = false
		compileCheck.Message = fmt.Sprintf("Failed to compile: %v", err)
	} else {
		compileCheck.Passed = true
		compileCheck.Message = "OK"
	}
	checks = append(checks, compileCheck)

	// If compile failed, skip other checks
	if !compileCheck.Passed {
		return checks, nil
	}

	// Asciidoctor exits 0 on a missing include or an undefined attribute,
	// so its warnings are checked on their own.
	warningsCheck := StructuralCheck{
		ID:      "no-warnings",
		Name:    "No asciidoctor warnings",
		Passed:  len(report.Warnings) == 0,
		Message: "OK",
	}
	if !warningsCheck.Passed {
		warningsCheck.Message = listMessage(report.Warnings)
	}
	checks = append(checks, warningsCheck)

	refsCheck := StructuralCheck{
		ID:      "references",
		Name:    "Cross-references resolve",
		Passed:  len(report.BrokenRefs) == 0,
		Message: "OK",
	}
	if !refsCheck.Passed {
		var refs []string
		for _, r := range report.BrokenRefs {
			refs = append(refs, "no target for <<"+r+">>")
		}
		refsCheck.Message = listMessage(refs)
	}
	checks = append(checks, refsCheck)

	// Check 2: Can parse structure
	structureCheck := StructuralCheck{
		ID:   "parseable",
		Name: "Structure parseable",
	}
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		structureCheck.Passed = false
		structureCheck.Message = fmt.Sprintf("Failed to parse structure: %v", err)
		checks = append(checks, structureCheck)
		return checks, nil
	}
	structureCheck.Passed = true
	structureCheck.Message = "OK"
	checks = append(checks, structureCheck)

	// Check 3: Has sections
	sectionsCheck := StructuralCheck{
		ID:   "has-sections",
		Name: "Has defined sections",
	}
	if len(structure.Sections) == 0 {
		sectionsCheck.Passed = false
		sectionsCheck.Message = "No sections found - spec appears empty"
	} else {
		sectionsCheck.Passed = true
		sectionsCheck.Message = fmt.Sprintf("Found %d sections", len(structure.Sections))
	}
	checks = append(checks, sectionsCheck)

	// Check 4: Has attributes (optional but good indicator)
	attrsCheck := StructuralCheck{
		ID:   "has-attributes",
		Name: "Has reusable attributes",
	}
	if len(structure.Attributes) == 0 {
		attrsCheck.Passed = true // Not a failure, just a note
		attrsCheck.Message = "No attributes defined (consider using :attr: for reusable values)"
	} else {
		attrsCheck.Passed = true
		attrsCheck.Message = fmt.Sprintf("Found %d attributes", len(structure.Attributes))
	}
	checks = append(checks, attrsCheck)

	return checks, nil
}

// listMessage formats several findings as an indented list under a count.
func listMessage(items []string) string {
	return fmt.Sprintf("%d found\n      %s", len(items), strings.Join(items, "\n      "))
}

// AllStructuralChecksPassed returns true if all checks passed
func AllStructuralChecksPassed(checks []StructuralCheck) bool {
	for _, check := range checks {
		if !check.Passed {
			return false
		}
	}
	return true
}

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

// FormatStructuralChecks formats checks for display, in color when stdout
// is a terminal
func FormatStructuralChecks(checks []StructuralCheck) string {
	return formatStructuralChecks(checks, useColor())
}

// useColor reports whether stdout is a terminal, so that piped output (an
// agent, CI, a file) carries no escape codes.
func useColor() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// colorize wraps s in a color when stdout is a terminal.
func colorize(s, color string) string {
	if !useColor() {
		return s
	}
	return color + s + colorReset
}

// FormatStructuralChecksPlain formats checks without color
func FormatStructuralChecksPlain(checks []StructuralCheck) string {
	return formatStructuralChecks(checks, false)
}

func formatStructuralChecks(checks []StructuralCheck, color bool) string {
	var sb strings.Builder

	sb.WriteString("Structural Checks:\n")
	for _, check := range checks {
		if check.Passed {
			if color {
				sb.WriteString(fmt.Sprintf("  %s✓%s %s: %s\n", colorGreen, colorReset, check.Name, check.Message))
			} else {
				sb.WriteString(fmt.Sprintf("  ✓ %s: %s\n", check.Name, check.Message))
			}
		} else {
			if color {
				sb.WriteString(fmt.Sprintf("  %s✗%s %s: %s\n", colorRed, colorReset, check.Name, check.Message))
			} else {
				sb.WriteString(fmt.Sprintf("  ✗ %s: %s\n", check.Name, check.Message))
			}
		}
	}

	return sb.String()
}

// FormatStructuralChecksJSON formats checks as JSON
func FormatStructuralChecksJSON(checks []StructuralCheck) string {
	data, _ := json.MarshalIndent(checks, "", "  ")
	return string(data)
}
