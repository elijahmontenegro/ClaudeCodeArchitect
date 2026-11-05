package validator

import (
	"fmt"
	"io"

	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/compiler"
)

// ValidationResult represents the complete validation result
type ValidationResult struct {
	StructuralChecks []StructuralCheck `json:"structural_checks"`
	StructuralPassed bool              `json:"structural_passed"`
	ReviewRun        bool              `json:"review_run"`
	Cancelled        bool              `json:"cancelled"`
}

// Validate runs the structural checks, then, if they pass, has Claude review
// the compiled docs against a checklist: the docs' own, or the default.
func Validate(entryPath string, output io.Writer, opts ValidationOptions) (*ValidationResult, error) {
	checklist, err := LoadChecklist(opts.Checklist)
	if err != nil {
		return nil, err
	}

	result := &ValidationResult{}

	fmt.Fprint(output, "=== Structural checks ===\n\n")

	checks, err := RunStructuralChecks(entryPath)
	if err != nil {
		return nil, fmt.Errorf("structural checks failed: %w", err)
	}
	result.StructuralChecks = checks
	result.StructuralPassed = AllStructuralChecksPassed(checks)

	fmt.Fprint(output, FormatStructuralChecks(checks))
	fmt.Fprintln(output)

	if !result.StructuralPassed {
		fmt.Fprintln(output, colorize("✗", colorRed)+" Structural checks failed. Fix these before the review.")
		return result, nil
	}

	fmt.Fprint(output, colorize("✓", colorGreen)+" Structural checks passed\n\n")

	source := "default checklist"
	if opts.Checklist != "" {
		source = opts.Checklist
	}
	fmt.Fprintf(output, "=== Review (Claude, %s) ===\n\n", source)

	if !IsClaudeAvailable() {
		return nil, fmt.Errorf("claude CLI not found - required for the review\n\nInstall from: https://claude.ai/code\n\nOr use 'validate --quick' for structural checks only")
	}

	compiledDocs, err := compiler.Compile(entryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compile the docs: %w", err)
	}

	proceed, err := CheckSize(compiledDocs, opts, output)
	if err != nil {
		return nil, fmt.Errorf("size check failed: %w", err)
	}
	if !proceed {
		fmt.Fprintln(output, "Review cancelled.")
		result.Cancelled = true
		return result, nil
	}

	result.ReviewRun = true
	if opts.Ultra {
		err = RunUltraReview(compiledDocs, checklist, output)
	} else {
		err = RunReview(compiledDocs, checklist, output)
	}
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(output)
	return result, nil
}

// ValidateQuick runs only structural checks (no Claude)
func ValidateQuick(entryPath string) (*ValidationResult, error) {
	result := &ValidationResult{}

	checks, err := RunStructuralChecks(entryPath)
	if err != nil {
		return nil, err
	}

	result.StructuralChecks = checks
	result.StructuralPassed = AllStructuralChecksPassed(checks)

	return result, nil
}
