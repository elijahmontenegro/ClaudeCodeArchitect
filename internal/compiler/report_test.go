package compiler

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseDiagnostics(t *testing.T) {
	stderr := "asciidoctor: ERROR: MANIFEST.adoc: line 6: include file not found: sub/missing.adoc\r\n" +
		"some unrelated line\n" +
		"asciidoctor: WARNING: skipping reference to missing attribute: nope\n" +
		"\n"
	got := parseDiagnostics(stderr)
	want := []string{
		"ERROR: MANIFEST.adoc: line 6: include file not found: sub/missing.adoc",
		"WARNING: skipping reference to missing attribute: nope",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBrokenReferences(t *testing.T) {
	html := `<h2 id="_rules">Rules</h2>
<p>See <a href="#_rules">Rules</a>, <a href="#nowhere">[nowhere]</a>, <a href="#nowhere">[nowhere]</a>.</p>
<p><a id="there"></a><a href="#there">there</a> <a href="#">top</a> <a href="other.html#x">elsewhere</a> <a href="#gone">[gone]</a></p>`

	got := BrokenReferences(html)
	want := []string{"nowhere", "gone"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCompileWithReport_FindsProblems(t *testing.T) {
	requireAsciidoctor(t)
	dir := writeDocs(t, map[string]string{
		"MANIFEST.adoc": "= T\n\n== Top\nSee <<nowhere>> and <<there>>. Never defined: {nope}. Built-in: A{nbsp}B.\n\n" +
			"include::sub/missing.adoc[]\n\n[[there]]\ninclude::sub/there.adoc[]\n",
		"sub/there.adoc": "== There\nok\n",
	})

	report, err := CompileWithReport(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}

	joined := strings.Join(report.Warnings, "\n")
	for _, want := range []string{"include file not found", "missing attribute: nope"} {
		if !strings.Contains(joined, want) {
			t.Errorf("warnings are missing %q: %q", want, report.Warnings)
		}
	}
	if strings.Contains(joined, "nbsp") {
		t.Errorf("a built-in attribute was reported missing: %q", report.Warnings)
	}
	if !reflect.DeepEqual(report.BrokenRefs, []string{"nowhere"}) {
		t.Errorf("broken refs = %q, want [nowhere]", report.BrokenRefs)
	}
}

func TestCompileWithReport_CleanDocs(t *testing.T) {
	requireAsciidoctor(t)
	dir := writeDocs(t, sampleDocs)

	report, err := CompileWithReport(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Warnings) > 0 || len(report.BrokenRefs) > 0 {
		t.Errorf("clean docs reported warnings %q and broken refs %q", report.Warnings, report.BrokenRefs)
	}
}

func TestCompileWithReport_DocumentCanSilenceMissingAttributes(t *testing.T) {
	requireAsciidoctor(t)
	dir := writeDocs(t, map[string]string{
		"MANIFEST.adoc": "= T\n:attribute-missing: skip\n\n== A\n{nope}\n",
	})

	report, err := CompileWithReport(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Warnings) > 0 {
		t.Errorf("document set attribute-missing: skip, but got %q", report.Warnings)
	}
}
