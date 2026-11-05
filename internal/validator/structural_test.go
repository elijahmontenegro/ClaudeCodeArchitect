package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emontenegr/ClaudeCodeArchitect/internal/compiler"
)

func writeDocs(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func checkByID(checks []StructuralCheck, id string) *StructuralCheck {
	for i := range checks {
		if checks[i].ID == id {
			return &checks[i]
		}
	}
	return nil
}

func TestRunStructuralChecks_FailsOnBrokenDocs(t *testing.T) {
	if !compiler.IsAsciidoctorAvailable() {
		t.Skip("asciidoctor not in PATH")
	}
	dir := writeDocs(t, map[string]string{
		"MANIFEST.adoc": "= T\n\n== Top\nSee <<nowhere>>. {nope}\n\ninclude::missing.adoc[]\n",
	})

	checks, err := RunStructuralChecks(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}
	if AllStructuralChecksPassed(checks) {
		t.Fatal("broken docs passed the structural checks")
	}

	warnings := checkByID(checks, "no-warnings")
	if warnings == nil || warnings.Passed || !strings.Contains(warnings.Message, "include file not found") ||
		!strings.Contains(warnings.Message, "missing attribute: nope") {
		t.Errorf("no-warnings check = %+v", warnings)
	}
	refs := checkByID(checks, "references")
	if refs == nil || refs.Passed || !strings.Contains(refs.Message, "<<nowhere>>") {
		t.Errorf("references check = %+v", refs)
	}
}

func TestRunStructuralChecks_PassesCleanDocs(t *testing.T) {
	if !compiler.IsAsciidoctorAvailable() {
		t.Skip("asciidoctor not in PATH")
	}
	dir := writeDocs(t, map[string]string{
		"MANIFEST.adoc": "= T\n:v: 1\n\n[[top]]\n== Top\nValue {v}. See <<top>>.\n",
	})

	checks, err := RunStructuralChecks(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range checks {
		if !c.Passed {
			t.Errorf("%s failed: %s", c.Name, c.Message)
		}
	}
}

func TestFormatStructuralChecks_NoEscapeCodesWhenPiped(t *testing.T) {
	// Under `go test`, stdout is not a terminal.
	out := FormatStructuralChecks([]StructuralCheck{{Name: "x", Passed: true, Message: "OK"}})
	if strings.Contains(out, "\033[") {
		t.Errorf("output carries escape codes: %q", out)
	}
}
