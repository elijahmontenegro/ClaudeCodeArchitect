package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadChecklist(t *testing.T) {
	got, err := LoadChecklist("")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"contradiction", "undefined-term", "decision-status", "undated-fact", "unexplained-number"} {
		if !strings.Contains(got, id) {
			t.Errorf("default checklist is missing %q", id)
		}
	}

	dir := t.TempDir()
	custom := filepath.Join(dir, "checklist.md")
	if err := os.WriteFile(custom, []byte("1. **owner-named**: Every system names its owner.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err = LoadChecklist(custom)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "owner-named") || strings.Contains(got, "contradiction") {
		t.Errorf("custom checklist should replace the default, got %q", got)
	}

	if _, err := LoadChecklist(filepath.Join(dir, "missing.md")); err == nil {
		t.Error("expected an error for a missing checklist")
	}
	empty := filepath.Join(dir, "empty.md")
	os.WriteFile(empty, []byte("  \n"), 0644)
	if _, err := LoadChecklist(empty); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected an error for an empty checklist, got %v", err)
	}
}

func TestReviewPrompt(t *testing.T) {
	docs := "# Docs\n\n```go\nfmt.Println(\"a code block\")\n```\n"
	prompt, err := reviewPrompt(docs, "1. **owner-named**: Every system names its owner.")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"owner-named", "<docs>\n" + docs + "\n</docs>", "The docs pass every check."} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt is missing %q:\n%s", want, prompt)
		}
	}
	for _, retired := range []string{"architecture", "19-Point", "exact-versions"} {
		if strings.Contains(prompt, retired) {
			t.Errorf("prompt still carries %q", retired)
		}
	}
}

func TestSynthesisPrompt(t *testing.T) {
	prompt, err := RenderPrompt("synthesize", TemplateData{Run1: "first", Run2: "second", Run3: "third"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<review-1>\nfirst\n</review-1>", "<review-3>\nthird\n</review-3>", "Confidence"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("synthesis prompt is missing %q", want)
		}
	}
	if strings.Contains(prompt, "architecture") {
		t.Error("synthesis prompt still assumes an architecture spec")
	}
}
