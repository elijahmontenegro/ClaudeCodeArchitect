package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

// Target is what a section query names: one section, or a whole file.
type Target struct {
	Section *SectionInfo
	File    string
}

// Lookup resolves a section query. It tries, in order: an exact title, a
// "path#Title" pair, a file path, and a partial title, all case-insensitive.
// The first kind that matches decides. If it matches more than one section
// or file, Lookup returns an error listing them instead of picking one.
func Lookup(structure *SpecStructure, query string) (Target, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return Target{}, fmt.Errorf("empty section query")
	}
	lq := strings.ToLower(q)
	baseDir := filepath.Dir(structure.ManifestPath)

	sectionPasses := []func(s *SectionInfo) bool{
		func(s *SectionInfo) bool { return strings.ToLower(s.Title) == lq },
	}
	if i := strings.LastIndex(q, "#"); i > 0 {
		path, title := q[:i], strings.ToLower(strings.TrimSpace(q[i+1:]))
		sectionPasses = append(sectionPasses, func(s *SectionInfo) bool {
			return pathMatches(s.FilePath, path) && strings.ToLower(s.Title) == title
		})
	}
	for _, pass := range sectionPasses {
		if target, found, err := lookupSection(structure, q, baseDir, pass); found || err != nil {
			return target, err
		}
	}

	var files []string
	for _, f := range append([]string{structure.ManifestPath}, structure.Files...) {
		if pathMatches(f, q) && !contains(files, f) {
			files = append(files, f)
		}
	}
	switch len(files) {
	case 1:
		return Target{File: files[0]}, nil
	case 0:
	default:
		var sb strings.Builder
		fmt.Fprintf(&sb, "%q matches %d files:\n", q, len(files))
		for _, f := range files {
			fmt.Fprintf(&sb, "  %s\n", relPath(baseDir, f))
		}
		return Target{}, fmt.Errorf("%s", strings.TrimRight(sb.String(), "\n"))
	}

	partial := func(s *SectionInfo) bool { return strings.Contains(strings.ToLower(s.Title), lq) }
	if target, found, err := lookupSection(structure, q, baseDir, partial); found || err != nil {
		return target, err
	}

	return Target{}, fmt.Errorf("section not found: %s", q)
}

// lookupSection applies one matching pass. It reports found when exactly one
// section matches, and an error listing the candidates when several do.
func lookupSection(structure *SpecStructure, query, baseDir string, match func(*SectionInfo) bool) (Target, bool, error) {
	var matches []*SectionInfo
	for i := range structure.Sections {
		if match(&structure.Sections[i]) {
			matches = append(matches, &structure.Sections[i])
		}
	}
	switch len(matches) {
	case 0:
		return Target{}, false, nil
	case 1:
		return Target{Section: matches[0]}, true, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%q matches %d sections:\n", query, len(matches))
	for _, s := range matches {
		fmt.Fprintf(&sb, "  %s:%d  %s\n", relPath(baseDir, s.FilePath), s.StartLine, s.Title)
	}
	last := matches[len(matches)-1]
	fmt.Fprintf(&sb, "Name one by its exact title, or as path#Title (for example %s#%s)",
		relPath(baseDir, last.FilePath), last.Title)
	return Target{}, false, fmt.Errorf("%s", sb.String())
}

// pathMatches reports whether query names path: the same path, or a trailing
// part of it that starts at a directory boundary. Either separator works.
func pathMatches(path, query string) bool {
	p := strings.ToLower(filepath.ToSlash(path))
	q := strings.ToLower(filepath.ToSlash(strings.TrimSpace(query)))
	q = strings.TrimPrefix(q, "./")
	if q == "" {
		return false
	}
	return p == q || strings.HasSuffix(p, "/"+q)
}

// relPath shows path relative to baseDir, with forward slashes.
func relPath(baseDir, path string) string {
	if rel, err := filepath.Rel(baseDir, path); err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(path)
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
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
