package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// AttributeDefinition represents a single attribute definition
type AttributeDefinition struct {
	Name     string
	Value    string
	FilePath string
	Line     int
}

// AttributeUsage represents a reference to an attribute in content
type AttributeUsage struct {
	Name         string
	FilePath     string
	Line         int
	Context      string // Surrounding text for context
	SectionTitle string // Which section contains this usage
}

// Regex patterns for attribute detection
var (
	// Matches :attribute-name: value
	attrDefPattern = regexp.MustCompile(`^:([a-zA-Z0-9_-]+):\s*(.*)$`)

	// Matches {attribute-name} references
	attrRefPattern = regexp.MustCompile(`\{([a-zA-Z0-9_-]+)\}`)
)

// ExtractAttributes extracts all attribute declarations from content
func ExtractAttributes(content string) map[string]string {
	attrs := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if matches := attrDefPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			value := strings.TrimSpace(matches[2])
			attrs[name] = value
		}
	}

	return attrs
}

// ExtractAttributesFromFile extracts attributes from a file with line numbers
func ExtractAttributesFromFile(filePath string) ([]AttributeDefinition, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var attrs []AttributeDefinition
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if matches := attrDefPattern.FindStringSubmatch(line); matches != nil {
			attrs = append(attrs, AttributeDefinition{
				Name:     matches[1],
				Value:    strings.TrimSpace(matches[2]),
				FilePath: filePath,
				Line:     lineNum,
			})
		}
	}

	return attrs, scanner.Err()
}

// CollectAttributes returns every attribute entry in a document in the order
// asciidoctor reads them: the manifest's lines, with each included file's
// entries at the point where it is included. Entries shown inside verbatim
// blocks are skipped. A later entry for the same name overrides an earlier one.
func CollectAttributes(manifestPath string) ([]AttributeDefinition, error) {
	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(absPath); err != nil {
		return nil, err
	}

	var defs []AttributeDefinition
	collectAttributes(absPath, map[string]bool{}, &defs)
	return defs, nil
}

func collectAttributes(absPath string, including map[string]bool, defs *[]AttributeDefinition) {
	if including[absPath] {
		return // include cycle
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return // a missing include is reported by the compile check
	}
	including[absPath] = true
	defer delete(including, absPath)

	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	var blocks blockTracker
	for i, line := range strings.Split(text, "\n") {
		if blocks.step(line) {
			continue
		}
		// Asciidoctor resolves include directives before parsing, even inside
		// blocks. A partial include (tag=, tags=, lines=) brings in a region,
		// not the file's attribute entries, so only whole includes are followed.
		if m := includePattern.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			if !isPartial(m[2]) {
				collectAttributes(ResolveIncludePath(filepath.Dir(absPath), m[1]), including, defs)
			}
			continue
		}
		if blocks.verbatim() {
			continue
		}
		if m := attrDefPattern.FindStringSubmatch(line); m != nil {
			*defs = append(*defs, AttributeDefinition{
				Name:     m[1],
				Value:    strings.TrimSpace(m[2]),
				FilePath: absPath,
				Line:     i + 1,
			})
		}
	}
}

// FindAttributeUsages finds all references to a specific attribute in content
func FindAttributeUsages(content, filePath, attrName string) []AttributeUsage {
	var usages []AttributeUsage
	pattern := regexp.MustCompile(`\{` + regexp.QuoteMeta(attrName) + `\}`)

	lines := strings.Split(content, "\n")
	currentSection := ""

	for lineNum, line := range lines {
		// Track current section from headings
		if strings.HasPrefix(line, "=") {
			currentSection = strings.TrimLeft(line, "= ")
			currentSection = strings.TrimSpace(currentSection)
		}

		if pattern.MatchString(line) {
			usages = append(usages, AttributeUsage{
				Name:         attrName,
				FilePath:     filePath,
				Line:         lineNum + 1,
				Context:      strings.TrimSpace(line),
				SectionTitle: currentSection,
			})
		}
	}

	return usages
}

// FindAllAttributeUsages finds all attribute references in content
func FindAllAttributeUsages(content, filePath string) []AttributeUsage {
	var usages []AttributeUsage

	lines := strings.Split(content, "\n")
	currentSection := ""

	for lineNum, line := range lines {
		// Track current section from headings
		if strings.HasPrefix(line, "=") {
			currentSection = strings.TrimLeft(line, "= ")
			currentSection = strings.TrimSpace(currentSection)
		}

		matches := attrRefPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			usages = append(usages, AttributeUsage{
				Name:         match[1],
				FilePath:     filePath,
				Line:         lineNum + 1,
				Context:      strings.TrimSpace(line),
				SectionTitle: currentSection,
			})
		}
	}

	return usages
}

// ResolveAttributes substitutes attribute references in content
func ResolveAttributes(content string, attrs map[string]string) string {
	result := content
	for name, value := range attrs {
		pattern := regexp.MustCompile(`\{` + regexp.QuoteMeta(name) + `\}`)
		result = pattern.ReplaceAllString(result, value)
	}
	return result
}

// GetAttributeNames returns just the names of referenced attributes in content
func GetAttributeNames(content string) []string {
	seen := make(map[string]bool)
	var names []string

	matches := attrRefPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		name := match[1]
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}

	return names
}
