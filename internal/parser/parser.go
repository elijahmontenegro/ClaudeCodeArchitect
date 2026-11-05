package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// SpecStructure represents the complete structure of a specification
type SpecStructure struct {
	ManifestPath   string                         // Path to MANIFEST.adoc
	Attributes     map[string]AttributeDefinition // Each attribute's effective (last) definition
	AttributeOrder []AttributeDefinition          // Every attribute entry, in document order
	Includes       []IncludeInfo                  // All include directives
	Sections       []SectionInfo                  // All sections across all files
	Files          []string                       // All included files
}

// SectionInfo represents a section in the specification
type SectionInfo struct {
	Title     string // Section heading text
	Level     int    // Heading level (1-6)
	FilePath  string // Source file containing section
	StartLine int    // Starting line in source
	EndLine   int    // Ending line (-1 for last section in file)
}

// Regex for section headings
// = Level 0 (document title)
// == Level 1
// === Level 2, etc.
var sectionPattern = regexp.MustCompile(`^(=+)\s+(.+)$`)

// BuildStructure builds the complete spec structure from a manifest
func BuildStructure(manifestPath string) (*SpecStructure, error) {
	structure := &SpecStructure{
		ManifestPath: manifestPath,
		Attributes:   make(map[string]AttributeDefinition),
	}

	// Extract attributes from the manifest and every file it includes
	attrs, err := CollectAttributes(manifestPath)
	if err != nil {
		return nil, err
	}
	structure.AttributeOrder = attrs
	for _, attr := range attrs {
		structure.Attributes[attr.Name] = attr
	}

	// Extract includes from manifest
	includes, err := ExtractIncludesFromFile(manifestPath)
	if err != nil {
		return nil, err
	}
	structure.Includes = includes

	// Get all included files recursively
	files, err := GetIncludedFiles(manifestPath)
	if err != nil {
		return nil, err
	}
	structure.Files = files

	// Extract sections from manifest
	sections, err := ExtractSectionsFromFile(manifestPath)
	if err != nil {
		return nil, err
	}
	structure.Sections = append(structure.Sections, sections...)

	// Extract sections from all included files
	for _, filePath := range files {
		sections, err := ExtractSectionsFromFile(filePath)
		if err != nil {
			continue // Skip files that can't be read
		}
		structure.Sections = append(structure.Sections, sections...)
	}

	return structure, nil
}

// ExtractSectionsFromFile extracts all section headings from a file
func ExtractSectionsFromFile(filePath string) ([]SectionInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sections []SectionInfo
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var blocks blockTracker

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// A heading shown inside a delimited block isn't a section.
		if blocks.step(line) || blocks.inBlock() {
			continue
		}

		if matches := sectionPattern.FindStringSubmatch(line); matches != nil {
			sections = append(sections, SectionInfo{
				Title:     strings.TrimSpace(matches[2]),
				Level:     len(matches[1]) - 1, // = is level 0, == is level 1, etc.
				FilePath:  filePath,
				StartLine: lineNum,
				EndLine:   -1, // runs to the end of the file unless a later heading closes it
			})
		}
	}

	// A section holds its subsections: it runs to the line before the next
	// heading at its own level or above.
	for i := range sections {
		for j := i + 1; j < len(sections); j++ {
			if sections[j].Level <= sections[i].Level {
				sections[i].EndLine = sections[j].StartLine - 1
				break
			}
		}
	}

	return sections, scanner.Err()
}

// FindSection finds a section by title or file path
func FindSection(structure *SpecStructure, query string) *SectionInfo {
	query = strings.ToLower(strings.TrimSpace(query))

	for i := range structure.Sections {
		section := &structure.Sections[i]

		// Match by exact title (case-insensitive)
		if strings.ToLower(section.Title) == query {
			return section
		}

		// Match by file path
		if strings.HasSuffix(strings.ToLower(section.FilePath), query) {
			return section
		}

		// Match by partial title
		if strings.Contains(strings.ToLower(section.Title), query) {
			return section
		}
	}

	return nil
}

// FindSectionByFile finds all sections in a specific file
func FindSectionsByFile(structure *SpecStructure, filePath string) []SectionInfo {
	var sections []SectionInfo
	for _, section := range structure.Sections {
		if section.FilePath == filePath {
			sections = append(sections, section)
		}
	}
	return sections
}

// GetSectionContent extracts the raw content of a section from its file
func GetSectionContent(section *SectionInfo) (string, error) {
	file, err := os.Open(section.FilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if lineNum >= section.StartLine {
			if section.EndLine > 0 && lineNum > section.EndLine {
				break
			}
			lines = append(lines, scanner.Text())
		}
	}

	return strings.Join(lines, "\n"), scanner.Err()
}

// GetFileContent reads the entire content of a file
func GetFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetAttributeMap returns attributes as a simple map[string]string
func (s *SpecStructure) GetAttributeMap() map[string]string {
	result := make(map[string]string)
	for name, attr := range s.Attributes {
		result[name] = attr.Value
	}
	return result
}
