package compiler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/emontenegr/ClaudeCodeArchitect/internal/parser"
)

// CompileSection compiles what a section query names, with the document's
// attributes resolved: one section with its subsections, or a whole file.
// See parser.Lookup for how queries match.
func CompileSection(manifestPath, sectionQuery string) (string, error) {
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse spec structure: %v", err)
	}

	target, err := parser.Lookup(structure, sectionQuery)
	if err != nil {
		if strings.HasPrefix(err.Error(), "section not found") {
			return "", fmt.Errorf("%v\nTop-level sections:\n%s", err, formatAvailableSections(structure.Sections))
		}
		return "", err
	}
	if target.File != "" {
		return compileFile(structure, target.File)
	}
	section := target.Section

	content, err := parser.GetSectionContent(section)
	if err != nil {
		return "", fmt.Errorf("failed to read section content: %v", err)
	}

	// Prepend attributes for resolution, and shift the section to the top
	// level so asciidoctor doesn't see a subsection without its parent.
	attrBlock := buildAttributeBlock(structure.AttributeOrder)
	fullContent := attrBlock + "\n" + levelOffset(section.Level) + "\n" + content

	// Compile with attributes resolved
	baseDir := filepath.Dir(section.FilePath)
	return CompileContent(fullContent, baseDir)
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
func CompileFile(manifestPath, filePath string) (string, error) {
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse spec structure: %v", err)
	}
	return compileFile(structure, filePath)
}

func compileFile(structure *parser.SpecStructure, filePath string) (string, error) {
	content, err := parser.GetFileContent(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Prepend the document's attributes, shifting the file's first heading
	// to the top level
	firstLevel := 1
	for _, s := range structure.Sections {
		if s.FilePath == filePath {
			firstLevel = s.Level
			break
		}
	}
	attrBlock := buildAttributeBlock(structure.AttributeOrder)
	fullContent := attrBlock + "\n" + levelOffset(firstLevel) + "\n" + content

	// Compile with attributes resolved
	baseDir := filepath.Dir(filePath)
	return CompileContent(fullContent, baseDir)
}

// ListSections returns all sections in the spec for navigation
func ListSections(manifestPath string) ([]SectionSummary, error) {
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec structure: %v", err)
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
