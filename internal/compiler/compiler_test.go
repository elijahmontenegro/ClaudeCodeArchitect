package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeDocs creates files under a temp dir and returns the dir.
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

func requireAsciidoctor(t *testing.T) {
	t.Helper()
	if !IsAsciidoctorAvailable() {
		t.Skip("asciidoctor not in PATH")
	}
}

var sampleDocs = map[string]string{
	"MANIFEST.adoc": "= Sample Docs\n:base: alpha\n\n" +
		"include::rules.adoc[]\n",
	"rules.adoc": ":child: from-child\n\n" +
		"== Rules\nIntro {base}.\n\n" +
		"=== Working\nWork {child}.\n\n" +
		"=== Names\nNames.\n",
}

func TestCompile_ContentOnly(t *testing.T) {
	requireAsciidoctor(t)
	dir := writeDocs(t, sampleDocs)

	out, err := Compile(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "# Sample Docs") {
		t.Errorf("output should open with the document title as a heading, got %q", firstLine(out))
	}
	if strings.Count(out, "Sample Docs") != 1 {
		t.Error("document title appears more than once")
	}
	if strings.Contains(out, "Last updated") {
		t.Error("output carries the page footer")
	}
}

func TestCompileSection_HoldsSubsectionsAndResolvesAttributes(t *testing.T) {
	requireAsciidoctor(t)
	dir := writeDocs(t, sampleDocs)

	out, err := CompileSection(filepath.Join(dir, "MANIFEST.adoc"), "Rules")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## Rules", "Intro alpha.", "### Working", "Work from-child.", "### Names"} {
		if !strings.Contains(out, want) {
			t.Errorf("section output is missing %q:\n%s", want, out)
		}
	}
}

func TestCompileSection_SubsectionRendersAtTopLevel(t *testing.T) {
	requireAsciidoctor(t)
	dir := writeDocs(t, sampleDocs)

	out, err := CompileSection(filepath.Join(dir, "MANIFEST.adoc"), "Working")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "## Working") {
		t.Errorf("subsection should render as a top-level section, got %q", firstLine(out))
	}
	if strings.Contains(out, "Names") {
		t.Error("subsection output runs into its sibling")
	}
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
