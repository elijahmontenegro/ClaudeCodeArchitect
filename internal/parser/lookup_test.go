package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

// lookupStructure builds a structure whose paths sit under a real directory,
// as BuildStructure produces them.
func lookupStructure(root string) *DocStructure {
	p := func(rel string) string { return filepath.Join(root, filepath.FromSlash(rel)) }
	return &DocStructure{
		ManifestPath: p("MANIFEST.adoc"),
		Files:        []string{p("api.adoc"), p("db/schema.adoc"), p("palinome.adoc"), p("systems/service.adoc")},
		Sections: []SectionInfo{
			{Title: "API Errors", FilePath: p("api.adoc"), Level: 1, StartLine: 1},
			{Title: "API", FilePath: p("api.adoc"), Level: 1, StartLine: 9},
			{Title: "User Types", FilePath: p("api.adoc"), Level: 1, StartLine: 20},
			{Title: "Database Schema", FilePath: p("db/schema.adoc"), Level: 1, StartLine: 1},
			{Title: "Boundaries", FilePath: p("palinome.adoc"), Level: 2, StartLine: 42},
			{Title: "Boundaries", FilePath: p("systems/service.adoc"), Level: 2, StartLine: 23},
			{Title: "Performance", FilePath: p("db/schema.adoc"), Level: 2, StartLine: 30},
		},
	}
}

func TestLookup(t *testing.T) {
	root := t.TempDir()
	s := lookupStructure(root)

	tests := []struct {
		name      string
		query     string
		wantTitle string // a section
		wantFile  string // or a whole file, relative to root
		wantErr   string // or an error containing this
	}{
		{name: "exact title", query: "API", wantTitle: "API"},
		{name: "exact title beats an earlier partial match", query: "api", wantTitle: "API"},
		{name: "partial title", query: "User", wantTitle: "User Types"},
		{name: "whitespace", query: "  Performance  ", wantTitle: "Performance"},
		{name: "file by name", query: "schema.adoc", wantFile: "db/schema.adoc"},
		{name: "file by forward-slash path", query: "db/schema.adoc", wantFile: "db/schema.adoc"},
		{name: "file path must match at a directory boundary", query: "hema.adoc", wantErr: "section not found"},
		{name: "path#Title picks one of two", query: "systems/service.adoc#Boundaries", wantTitle: "Boundaries"},
		{name: "duplicate exact titles", query: "Boundaries", wantErr: `"Boundaries" matches 2 sections`},
		{name: "single partial match", query: "Error", wantTitle: "API Errors"},
		{name: "ambiguous partial titles", query: "a", wantErr: "matches"},
		{name: "no match", query: "nonexistent", wantErr: "section not found"},
		{name: "empty", query: "  ", wantErr: "empty section query"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := Lookup(s, tt.query)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want one containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantFile != "" {
				if want := filepath.Join(root, filepath.FromSlash(tt.wantFile)); target.File != want {
					t.Fatalf("file = %q, want %q", target.File, want)
				}
				return
			}
			if target.Section == nil || target.Section.Title != tt.wantTitle {
				t.Fatalf("section = %+v, want %q", target.Section, tt.wantTitle)
			}
		})
	}
}

func TestLookup_PathTitlePicksTheNamedFile(t *testing.T) {
	root := t.TempDir()
	target, err := Lookup(lookupStructure(root), "palinome.adoc#boundaries")
	if err != nil {
		t.Fatal(err)
	}
	if target.Section == nil || target.Section.StartLine != 42 {
		t.Fatalf("got %+v, want the Boundaries section in palinome.adoc", target.Section)
	}
}

func TestLookup_AmbiguityListsCandidates(t *testing.T) {
	root := t.TempDir()
	_, err := Lookup(lookupStructure(root), "Boundaries")
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"palinome.adoc:42", "systems/service.adoc:23", "systems/service.adoc#Boundaries"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error is missing %q:\n%v", want, err)
		}
	}
}
