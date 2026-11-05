package parser

import (
	"path/filepath"
	"testing"
)

func TestExtractSections_SectionHoldsItsSubsections(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"doc.adoc": "== Rules\n" + // 1
			"Intro.\n" + // 2
			"\n" + // 3
			"=== Working\n" + // 4
			"Work.\n" + // 5
			"\n" + // 6
			"=== Names\n" + // 7
			"Names.\n" + // 8
			"\n" + // 9
			"== Next\n" + // 10
			"Done.\n", // 11
	})

	sections, err := ExtractSectionsFromFile(filepath.Join(dir, "doc.adoc"))
	if err != nil {
		t.Fatal(err)
	}

	want := map[string][2]int{
		"Rules":   {1, 9},
		"Working": {4, 6},
		"Names":   {7, 9},
		"Next":    {10, -1},
	}
	if len(sections) != len(want) {
		t.Fatalf("got %d sections, want %d", len(sections), len(want))
	}
	for _, s := range sections {
		w := want[s.Title]
		if s.StartLine != w[0] || s.EndLine != w[1] {
			t.Errorf("%s spans %d-%d, want %d-%d", s.Title, s.StartLine, s.EndLine, w[0], w[1])
		}
	}
}

func TestExtractSections_IgnoresHeadingsInBlocks(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"doc.adoc": "== Real\n" +
			"An example of AsciiDoc:\n" +
			"\n" +
			"----\n" +
			"== Shown, not a section\n" +
			"----\n" +
			"\n" +
			"////\n" +
			"== Commented out\n" +
			"////\n" +
			"\n" +
			"== Also real\n",
	})

	sections, err := ExtractSectionsFromFile(filepath.Join(dir, "doc.adoc"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 2 || sections[0].Title != "Real" || sections[1].Title != "Also real" {
		var titles []string
		for _, s := range sections {
			titles = append(titles, s.Title)
		}
		t.Fatalf("got sections %q, want [Real, Also real]", titles)
	}
	if sections[0].EndLine != 11 {
		t.Errorf("Real ends at line %d, want 11", sections[0].EndLine)
	}
}

func TestBlockTracker(t *testing.T) {
	var b blockTracker
	steps := []struct {
		line     string
		fence    bool
		inBlock  bool
		verbatim bool
	}{
		{"====", true, true, false},     // example block opens
		{"----", true, true, true},      // listing nested inside it
		{"====", false, true, true},     // content inside the listing
		{"----", true, true, false},     // listing closes
		{"====", true, false, false},    // example closes
		{"--", true, true, false},       // open block holds AsciiDoc
		{"--", true, false, false},      // and closes
		{"-----", true, true, true},     // a 5-dash listing
		{"----", false, true, true},     // a 4-dash line doesn't close it
		{"-----", true, false, false},   // the matching fence does
		{"not a fence", false, false, false},
	}
	for i, s := range steps {
		fence := b.step(s.line)
		if fence != s.fence || b.inBlock() != s.inBlock || b.verbatim() != s.verbatim {
			t.Fatalf("step %d %q: fence=%v inBlock=%v verbatim=%v, want %v %v %v",
				i, s.line, fence, b.inBlock(), b.verbatim(), s.fence, s.inBlock, s.verbatim)
		}
	}
}
