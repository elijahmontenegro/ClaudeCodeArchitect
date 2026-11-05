package compiler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/emontenegr/ClaudeCodeArchitect/internal/parser"
)

// CompileSection compiles a specific section with attributes resolved
func CompileSection(manifestPath, sectionQuery string) (string, error) {
	// Build the spec structure
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse spec structure: %v", err)
	}

	// Find the matching section
	section := parser.FindSection(structure, sectionQuery)
	if section == nil {
		return "", fmt.Errorf("section not found: %s", sectionQuery)
	}

	// Get the section content
	content, err := parser.GetSectionContent(section)
	if err != nil {
		return "", fmt.Errorf("failed to read section content: %v", err)
	}

	// Prepend attributes for resolution
	attrBlock := buildAttributeBlock(structure.AttributeOrder)
	fullContent := attrBlock + "\n" + content

	// Compile with attributes resolved
	baseDir := filepath.Dir(section.FilePath)
	return CompileContent(fullContent, baseDir)
}

// CompileFile compiles a specific included file with attributes from manifest
func CompileFile(manifestPath, filePath string) (string, error) {
	// Build the spec structure to get attributes
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse spec structure: %v", err)
	}

	// Read the file content
	content, err := parser.GetFileContent(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Prepend manifest attributes
	attrBlock := buildAttributeBlock(structure.AttributeOrder)
	fullContent := attrBlock + "\n" + content

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

	var summaries []SectionSummary
	for _, section := range structure.Sections {
		summaries = append(summaries, SectionSummary{
			Title:    section.Title,
			Level:    section.Level,
			FilePath: section.FilePath,
			Line:     section.StartLine,
		})
	}

	return summaries, nil
}

// SectionSummary is a simplified section info for listing
type SectionSummary struct {
	Title    string
	Level    int
	FilePath string
	Line     int
}

// FormatSectionList formats sections for display
func FormatSectionList(sections []SectionSummary) string {
	var sb strings.Builder

	for _, s := range sections {
		indent := strings.Repeat("  ", s.Level)
		relPath := filepath.Base(s.FilePath)
		sb.WriteString(fmt.Sprintf("%s%s (%s:%d)\n", indent, s.Title, relPath, s.Line))
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

// FindMatchingSection finds a section by various matching strategies
func FindMatchingSection(manifestPath, query string) (*parser.SectionInfo, error) {
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec structure: %v", err)
	}

	section := parser.FindSection(structure, query)
	if section == nil {
		// Try matching as file path
		absPath := query
		if !filepath.IsAbs(query) {
			absPath = filepath.Join(filepath.Dir(manifestPath), query)
		}

		for _, s := range structure.Sections {
			if s.FilePath == absPath {
				return &s, nil
			}
		}

		return nil, fmt.Errorf("section not found: %s\nAvailable sections:\n%s",
			query, formatAvailableSections(structure.Sections))
	}

	return section, nil
}

func formatAvailableSections(sections []parser.SectionInfo) string {
	var sb strings.Builder
	for _, s := range sections {
		if s.Level <= 2 { // Only show top-level sections
			sb.WriteString(fmt.Sprintf("  - %s\n", s.Title))
		}
	}
	return sb.String()
}
