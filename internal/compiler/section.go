package compiler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/parser"
)

// CompileSection compiles what a section query names, with the document's
// attributes resolved: one section with its subsections, or a whole file.
// See parser.Lookup for how queries match.
func CompileSection(manifestPath, sectionQuery string) (Report, error) {
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return Report{}, fmt.Errorf("failed to read the document structure: %v", err)
	}

	target, err := parser.Lookup(structure, sectionQuery)
	if err != nil {
		if strings.HasPrefix(err.Error(), "section not found") {
			return Report{}, fmt.Errorf("%v\nTop-level sections:\n%s", err, formatAvailableSections(structure.Sections))
		}
		return Report{}, err
	}
	if target.File != "" {
		return compileFile(structure, target.File)
	}
	section := target.Section

	// EndLine -1 means the end of the file, which is also what lines=N..-1 means.
	lines := fmt.Sprintf("lines=%d..%d", section.StartLine, section.EndLine)
	return compileFragment(structure, section.FilePath, section.Level, lines)
}

// compileFragment compiles part of the document (a line range of one file, or
// all of it) with the whole document's attributes. The part is included from
// the manifest's directory rather than copied, so it compiles inside the same
// include jail as the full document, and its own includes resolve relative to
// its file as they do there. The fragment is shifted so that its first
// heading, at the given level, renders as a top-level section.
func compileFragment(structure *parser.DocStructure, filePath string, level int, includeAttrs string) (Report, error) {
	root := filepath.Dir(structure.ManifestPath)
	rel, err := filepath.Rel(root, filePath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return Report{}, fmt.Errorf("%s is outside the docs' root %s", filePath, root)
	}

	wrapper := buildAttributeBlock(structure.AttributeOrder) + "\n" +
		levelOffset(level) + "\n" +
		fmt.Sprintf("include::%s[%s]\n", filepath.ToSlash(rel), includeAttrs)
	return CompileContent(wrapper, root)
}

// levelOffset returns the attribute entry that renders a fragment whose first
// heading is at the given level as a top-level section (or, for level 0, as
// a document with its title shown).
func levelOffset(level int) string {
	switch {
	case level == 0:
		return ":showtitle:"
	case level > 1:
		return fmt.Sprintf(":leveloffset: -%d", level-1)
	default:
		return ""
	}
}

// CompileFile compiles one file of the document, with the document's
// attributes resolved
func CompileFile(manifestPath, filePath string) (Report, error) {
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return Report{}, fmt.Errorf("failed to read the document structure: %v", err)
	}
	return compileFile(structure, filePath)
}

func compileFile(structure *parser.DocStructure, filePath string) (Report, error) {
	firstLevel := 1
	for _, s := range structure.Sections {
		if s.FilePath == filePath {
			firstLevel = s.Level
			break
		}
	}
	return compileFragment(structure, filePath, firstLevel, "")
}

// ListSections returns every section of the document, for navigation
func ListSections(manifestPath string) ([]SectionSummary, error) {
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the document structure: %v", err)
	}

	baseDir := filepath.Dir(manifestPath)
	var summaries []SectionSummary
	for _, section := range structure.Sections {
		path := section.FilePath
		if rel, err := filepath.Rel(baseDir, path); err == nil {
			path = rel
		}
		summaries = append(summaries, SectionSummary{
			Title:    section.Title,
			Level:    section.Level,
			FilePath: filepath.ToSlash(path),
			Line:     section.StartLine,
		})
	}

	return summaries, nil
}

// SectionSummary is a simplified section info for listing
type SectionSummary struct {
	Title    string
	Level    int
	FilePath string // relative to the manifest, with forward slashes
	Line     int
}

// FormatSectionList formats sections for display
func FormatSectionList(sections []SectionSummary) string {
	var sb strings.Builder

	for _, s := range sections {
		indent := strings.Repeat("  ", s.Level)
		fmt.Fprintf(&sb, "%s%s (%s:%d)\n", indent, s.Title, s.FilePath, s.Line)
	}

	return sb.String()
}

// buildAttributeBlock writes a document's attribute entries back out in
// document order, so redefinitions and values built from other attributes
// resolve as they do in the full document.
func buildAttributeBlock(attrs []parser.AttributeDefinition) string {
	var lines []string
	for _, attr := range attrs {
		lines = append(lines, fmt.Sprintf(":%s: %s", attr.Name, attr.Value))
	}
	return strings.Join(lines, "\n")
}

func formatAvailableSections(sections []parser.SectionInfo) string {
	var sb strings.Builder
	for _, s := range sections {
		if s.Level == 1 {
			fmt.Fprintf(&sb, "  - %s\n", s.Title)
		}
	}
	return sb.String()
}
