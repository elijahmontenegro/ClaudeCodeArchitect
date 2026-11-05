package validator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/compiler"
	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/config"
)

// ValidationResult represents the complete validation result
type ValidationResult struct {
	StructuralChecks []StructuralCheck `json:"structural_checks"`
	StructuralPassed bool              `json:"structural_passed"`
	SemanticRun      bool              `json:"semantic_run"`
	Cancelled        bool              `json:"cancelled"`
}

// Validate runs the structural checks and, for a profile that has one, the
// semantic checklist with Claude. General docs (no profile) get the
// structural checks only.
func Validate(manifestPath string, output io.Writer, opts ValidationOptions) (*ValidationResult, error) {
	if opts.Profile != config.ProfileArchitecture {
		result, err := ValidateQuick(manifestPath)
		if err != nil {
			return nil, err
		}
		fmt.Fprint(output, FormatStructuralChecks(result.StructuralChecks))
		fmt.Fprintf(output, "\nNo semantic checklist for these docs. Set `profile: %s` in .cca.yaml to check an architecture spec for completeness.\n", config.ProfileArchitecture)
		return result, nil
	}

	result := &ValidationResult{}

	// Phase 1: Fast structural checks
	fmt.Fprint(output, "=== Phase 1: Structural Checks ===\n\n")

	checks, err := RunStructuralChecks(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("structural checks failed: %w", err)
	}
	result.StructuralChecks = checks
	result.StructuralPassed = AllStructuralChecksPassed(checks)

	fmt.Fprint(output, FormatStructuralChecks(checks))
	fmt.Fprintln(output)

	// If structural checks failed, stop here
	if !result.StructuralPassed {
		fmt.Fprintln(output, colorize("✗", colorRed)+" Structural checks failed. Fix these before semantic validation.")
		return result, nil
	}

	fmt.Fprint(output, colorize("✓", colorGreen)+" Structural checks passed\n\n")

	// Phase 2: Semantic validation with Claude
	fmt.Fprint(output, "=== Phase 2: Semantic Validation (Claude) ===\n\n")

	if !IsClaudeAvailable() {
		return nil, fmt.Errorf("claude CLI not found - required for semantic validation\n\nInstall from: https://claude.ai/code\n\nOr use 'validate --quick' for structural checks only")
	}

	// Compile the spec
	compiledSpec, err := compiler.Compile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compile spec: %w", err)
	}

	// Check spec size and confirm if large
	proceed, err := CheckSpecSize(compiledSpec, opts, output)
	if err != nil {
		return nil, fmt.Errorf("size check failed: %w", err)
	}
	if !proceed {
		fmt.Fprintln(output, "Validation cancelled by user.")
		result.Cancelled = true
		return result, nil
	}

	// Run Claude validation (ultra or normal)
	result.SemanticRun = true
	if opts.Ultra {
		if err := RunUltraValidation(compiledSpec, output); err != nil {
			return nil, fmt.Errorf("ultra validation failed: %w", err)
		}
	} else {
		if err := RunClaudeValidation(compiledSpec, output); err != nil {
			return nil, fmt.Errorf("semantic validation failed: %w", err)
		}
	}

	fmt.Fprintln(output)

	return result, nil
}

// ValidateQuick runs only structural checks (no Claude)
func ValidateQuick(manifestPath string) (*ValidationResult, error) {
	result := &ValidationResult{}

	checks, err := RunStructuralChecks(manifestPath)
	if err != nil {
		return nil, err
	}

	result.StructuralChecks = checks
	result.StructuralPassed = AllStructuralChecksPassed(checks)

	return result, nil
}

// FormatResult formats the validation result summary
func FormatResult(result *ValidationResult, baseDir string) string {
	if result.StructuralPassed {
		if result.SemanticRun {
			return "Validation complete (structural + semantic)"
		}
		return "Structural validation passed (semantic skipped)"
	}
	return "Validation failed at structural checks"
}

// FormatSummary returns a brief summary
func FormatSummary(result *ValidationResult) string {
	if result.StructuralPassed && result.SemanticRun {
		return "✓ Full validation complete"
	} else if result.StructuralPassed {
		return "✓ Structural checks passed (Claude not available for semantic)"
	}
	return "✗ Structural checks failed"
}

// ListRules returns all validation rules (for documentation)
func ListRules() []string {
	return []string{
		"exact-versions: All dependencies have exact versions",
		"no-or-choices: No unresolved alternatives",
		"no-conditionals: No conditional logic (if needed, optional, TBD)",
		"no-optional: No optional sections",
		"file-tree: Complete file structure provided",
		"types-complete: All types fully defined",
		"example-data: Actual examples, not just schemas",
		"db-schema: Database schema complete",
		"api-routes: API routes fully specified",
		"perf-quantified: Performance has numbers, not adjectives",
		"numeric-derivation: Constants have rationale",
		"data-formats: Format examples with real values",
		"error-handling: Error responses specified",
		"concurrency: Threading model specified if needed",
		"persistence: Storage format concrete",
		"deployment: Deployment config complete",
		"secrets-separated: Config/secrets properly separated",
		"no-weak-language: No should/could/might",
	}
}

// GetCompiledSpec compiles and returns the spec (for external use)
func GetCompiledSpec(manifestPath string) (string, error) {
	return compiler.Compile(manifestPath)
}

// ValidateWithOutput is a convenience function that writes to stdout with default options
func ValidateWithOutput(manifestPath string) (*ValidationResult, error) {
	return Validate(manifestPath, os.Stdout, ValidationOptions{})
}

// ValidateToFile writes validation output to a file with default options
func ValidateToFile(manifestPath, outputPath string) (*ValidationResult, error) {
	f, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Validate(manifestPath, f, ValidationOptions{})
}

// BaseDir returns the base directory for relative path display
func BaseDir(manifestPath string) string {
	return filepath.Dir(manifestPath)
}
