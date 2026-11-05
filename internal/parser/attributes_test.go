package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFiles creates files under a temp dir and returns the dir.
func writeFiles(t *testing.T, files map[string]string) string {
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

func TestCollectAttributes_IncludedFilesInDocumentOrder(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"MANIFEST.adoc": "= Doc\n:base: alpha\n\ninclude::sub/child.adoc[]\n\n:after: last\n",
		"sub/child.adoc": ":child: from-child\n:base: redefined\n\n== Child\n{child}\n" +
			"\n----\n:shown: in-a-listing\n----\n",
	})

	defs, err := CollectAttributes(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	for _, d := range defs {
		got = append(got, d.Name+"="+d.Value)
	}
	want := []string{"base=alpha", "child=from-child", "base=redefined", "after=last"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}

	child := defs[1]
	if filepath.Base(child.FilePath) != "child.adoc" || child.Line != 1 {
		t.Errorf("child defined at %s:%d, want child.adoc:1", child.FilePath, child.Line)
	}
}

func TestBuildStructure_LastDefinitionWins(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"MANIFEST.adoc":  ":base: alpha\n\ninclude::child.adoc[]\n",
		"child.adoc":     ":base: redefined\n:child: yes\n",
		"unrelated.adoc": ":never: included\n",
	})

	s, err := BuildStructure(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}
	if v := s.Attributes["base"].Value; v != "redefined" {
		t.Errorf("base = %q, want redefined", v)
	}
	if _, ok := s.Attributes["child"]; !ok {
		t.Error("attribute defined in an included file is missing")
	}
	if _, ok := s.Attributes["never"]; ok {
		t.Error("attribute from a file that isn't included was collected")
	}
	if len(s.AttributeOrder) != 3 {
		t.Errorf("AttributeOrder has %d entries, want 3", len(s.AttributeOrder))
	}
}

func TestCollectAttributes_IncludeCycle(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"MANIFEST.adoc": ":a: 1\ninclude::b.adoc[]\n",
		"b.adoc":        ":b: 2\ninclude::MANIFEST.adoc[]\n",
	})

	defs, err := CollectAttributes(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) != 2 {
		t.Errorf("got %d definitions, want 2", len(defs))
	}
}
