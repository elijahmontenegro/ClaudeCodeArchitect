package parser

import (
	"path/filepath"
	"testing"
)

func TestIsPartial(t *testing.T) {
	tests := map[string]bool{
		"":                     false,
		"leveloffset=+1":       false,
		"tag=retention":        true,
		"tags=a;b":             true,
		"lines=1..7":           true,
		"leveloffset=+1,tag=x": true,
		"opts=optional":        false,
	}
	for options, want := range tests {
		if got := isPartial(options); got != want {
			t.Errorf("isPartial(%q) = %v, want %v", options, got, want)
		}
	}
}

// A file included whole and also by tag elsewhere is one part of the document,
// placed where it's included whole.
func TestBuildStructure_FileIncludedWholeAndByTag(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"MANIFEST.adoc": "= Doc\n:days: 30\n\ninclude::overview.adoc[]\n\ninclude::terms.adoc[]\n\ninclude::systems/delivery.adoc[]\n",
		"overview.adoc": "== Overview\n\ninclude::systems/delivery.adoc[tag=retention]\n",
		"terms.adoc":    "== Terms\nT.\n",
		"systems/delivery.adoc": ":only-in-delivery: yes\n\n== Delivery\n\n=== Retention\n" +
			"// tag::retention[]\nKept {days} days.\n// end::retention[]\n",
	})

	s, err := BuildStructure(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}

	var titles []string
	for _, sec := range s.Sections {
		titles = append(titles, sec.Title)
	}
	want := []string{"Doc", "Overview", "Terms", "Delivery", "Retention"}
	if len(titles) != len(want) {
		t.Fatalf("sections = %q, want %q", titles, want)
	}
	for i := range want {
		if titles[i] != want[i] {
			t.Fatalf("sections = %q, want %q", titles, want)
		}
	}

	if len(s.Files) != 3 {
		t.Errorf("files = %q, want each of the 3 included files once", s.Files)
	}
	if _, err := Lookup(s, "Delivery"); err != nil {
		t.Errorf("Delivery should be one section: %v", err)
	}
	if _, ok := s.Attributes["only-in-delivery"]; !ok {
		t.Error("attribute in a file that is also included whole was missed")
	}
}

// A file only ever included by tag contributes neither sections nor the
// attribute entries outside its tagged region.
func TestBuildStructure_PartialOnlyInclude(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"MANIFEST.adoc": "= Doc\n\n== Main\n\ninclude::snippets.adoc[tag=one]\n",
		"snippets.adoc": ":outside: no\n\n== Snippets\n// tag::one[]\nOne.\n// end::one[]\n",
	})

	s, err := BuildStructure(filepath.Join(dir, "MANIFEST.adoc"))
	if err != nil {
		t.Fatal(err)
	}
	for _, sec := range s.Sections {
		if sec.Title == "Snippets" {
			t.Error("a file included only by tag placed its sections in the document")
		}
	}
	if _, ok := s.Attributes["outside"]; ok {
		t.Error("an attribute outside the included region was collected")
	}
	if len(s.Files) != 1 {
		t.Errorf("files = %q, want the snippets file listed once", s.Files)
	}
}
