package validator

// ValidationResult is the outcome of the structural checks
type ValidationResult struct {
	StructuralChecks []StructuralCheck `json:"structural_checks"`
	StructuralPassed bool              `json:"structural_passed"`
}

// Validate runs the structural checks: deterministic checks for docs that
// are broken, such as a missing include or a cross-reference with no target.
// Judging what the docs say is left to the agent reading them.
func Validate(entryPath string) (*ValidationResult, error) {
	checks, err := RunStructuralChecks(entryPath)
	if err != nil {
		return nil, err
	}

	return &ValidationResult{
		StructuralChecks: checks,
		StructuralPassed: AllStructuralChecksPassed(checks),
	}, nil
}
