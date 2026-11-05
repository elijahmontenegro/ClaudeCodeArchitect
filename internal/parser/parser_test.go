package parser

import "testing"

func TestFindSectionsByFile(t *testing.T) {
	structure := &DocStructure{
		Sections: []SectionInfo{
			{Title: "Section 1", FilePath: "a.adoc"},
			{Title: "Section 2", FilePath: "a.adoc"},
			{Title: "Section 3", FilePath: "b.adoc"},
			{Title: "Section 4", FilePath: "a.adoc"},
		},
	}

	result := FindSectionsByFile(structure, "a.adoc")

	if len(result) != 3 {
		t.Errorf("expected 3 sections, got %d", len(result))
	}

	result = FindSectionsByFile(structure, "nonexistent.adoc")
	if len(result) != 0 {
		t.Errorf("expected 0 sections, got %d", len(result))
	}
}

func TestSectionPattern(t *testing.T) {
	tests := []struct {
		line        string
		expectMatch bool
		expectLevel int
		expectTitle string
	}{
		{"= Document Title", true, 0, "Document Title"},
		{"== Section Level 1", true, 1, "Section Level 1"},
		{"=== Subsection", true, 2, "Subsection"},
		{"==== Deep Section", true, 3, "Deep Section"},
		{"Not a section", false, 0, ""},
		{"=No space", false, 0, ""},
		{"  == Indented", false, 0, ""},
		{"== Title with = sign", true, 1, "Title with = sign"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			matches := sectionPattern.FindStringSubmatch(tt.line)

			if tt.expectMatch {
				if matches == nil {
					t.Errorf("expected match, got nil")
					return
				}
				level := len(matches[1]) - 1
				title := matches[2]

				if level != tt.expectLevel {
					t.Errorf("expected level %d, got %d", tt.expectLevel, level)
				}
				if title != tt.expectTitle {
					t.Errorf("expected title %q, got %q", tt.expectTitle, title)
				}
			} else {
				if matches != nil {
					t.Errorf("expected no match, got %v", matches)
				}
			}
		})
	}
}
