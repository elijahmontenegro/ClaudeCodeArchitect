package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A document can't include a file from outside its root, so docs an agent is
// asked to read can't pull other files on the machine into the output.
func TestCompile_IncludesConfinedToDocRoot(t *testing.T) {
	requireAsciidoctor(t)
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("DECOY-SECRET-12345\n"), 0644); err != nil {
		t.Fatal(err)
	}
	dir := writeDocs(t, map[string]string{
		"MANIFEST.adoc": "= T\n\n== A\n\ninclude::" + filepath.ToSlash(outside) + "[]\n\ninclude::sub/ok.adoc[]\n",
		"sub/ok.adoc":   "== Ok\nInside the root.\n",
	})

	report, err := CompileWithReport(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(report.Markdown, "DECOY-SECRET") {
		t.Fatal("a file outside the doc root was included")
	}
	if !strings.Contains(report.Markdown, "Inside the root.") {
		t.Error("an include inside the doc root was dropped")
	}
	if !strings.Contains(strings.Join(report.Warnings, "\n"), "outside of jail") {
		t.Errorf("expected a warning about the include outside the root, got %q", report.Warnings)
	}
}

func TestCompileSection_IncludesResolveFromTheSectionsFile(t *testing.T) {
	requireAsciidoctor(t)
	dir := writeDocs(t, map[string]string{
		"MANIFEST.adoc":    "= T\n:v: 7\n\ninclude::systems/sys.adoc[]\n",
		"shared.adoc":      "== Shared\nShared text {v}.\n",
		"systems/sys.adoc": "== Sys\nIntro.\n\n=== Part\n\ninclude::../shared.adoc[leveloffset=+2]\n\n== After\nNot this.\n",
	})

	report, err := CompileSection(filepath.Join(dir, "MANIFEST.adoc"), "Sys")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Warnings) > 0 {
		t.Errorf("unexpected warnings: %q", report.Warnings)
	}
	for _, want := range []string{"## Sys", "### Part", "Shared text 7."} {
		if !strings.Contains(report.Markdown, want) {
			t.Errorf("missing %q:\n%s", want, report.Markdown)
		}
	}
	if strings.Contains(report.Markdown, "Not this.") {
		t.Errorf("section ran into the next one:\n%s", report.Markdown)
	}
}
