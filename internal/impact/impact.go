package impact

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/parser"
)

// AttributeImpact represents the impact analysis for a single attribute
type AttributeImpact struct {
	AttributeName string
	Definition    *parser.AttributeDefinition
	Usages        []parser.AttributeUsage
}

// AnalyzeAttribute finds all usages of a specific attribute
func AnalyzeAttribute(manifestPath, attrName string) (*AttributeImpact, error) {
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec structure: %v", err)
	}

	impact := &AttributeImpact{
		AttributeName: attrName,
	}

	// Find the definition
	if def, ok := structure.Attributes[attrName]; ok {
		impact.Definition = &def
	}

	// Search in manifest
	manifestContent, err := parser.GetFileContent(manifestPath)
	if err == nil {
		usages := parser.FindAttributeUsages(manifestContent, manifestPath, attrName)
		impact.Usages = append(impact.Usages, usages...)
	}

	// Search in all included files
	for _, filePath := range structure.Files {
		content, err := parser.GetFileContent(filePath)
		if err != nil {
			continue
		}
		usages := parser.FindAttributeUsages(content, filePath, attrName)
		impact.Usages = append(impact.Usages, usages...)
	}

	return impact, nil
}

// AnalyzeAllAttributes returns impact for all defined attributes
func AnalyzeAllAttributes(manifestPath string) (map[string]*AttributeImpact, error) {
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec structure: %v", err)
	}

	impacts := make(map[string]*AttributeImpact)

	for attrName := range structure.Attributes {
		impact, err := AnalyzeAttribute(manifestPath, attrName)
		if err != nil {
			continue
		}
		impacts[attrName] = impact
	}

	return impacts, nil
}

// ListAttributes returns all defined attributes
func ListAttributes(manifestPath string) ([]parser.AttributeDefinition, error) {
	structure, err := parser.BuildStructure(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec structure: %v", err)
	}

	var attrs []parser.AttributeDefinition
	for _, attr := range structure.Attributes {
		attrs = append(attrs, attr)
	}

	return attrs, nil
}

// FormatImpact formats an attribute impact for display
func FormatImpact(impact *AttributeImpact, baseDir string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Attribute: %s\n", impact.AttributeName))

	if impact.Definition != nil {
		relPath, _ := filepath.Rel(baseDir, impact.Definition.FilePath)
		if relPath == "" {
			relPath = filepath.Base(impact.Definition.FilePath)
		}
		sb.WriteString(fmt.Sprintf("Defined in: %s:%d = \"%s\"\n",
			relPath, impact.Definition.Line, impact.Definition.Value))
	} else {
		sb.WriteString("Defined in: (not found)\n")
	}

	sb.WriteString("\nUsed in:\n")

	if len(impact.Usages) == 0 {
		sb.WriteString("  (no usages found)\n")
	} else {
		// Group by file
		byFile := make(map[string][]parser.AttributeUsage)
		for _, u := range impact.Usages {
			byFile[u.FilePath] = append(byFile[u.FilePath], u)
		}

		for filePath, usages := range byFile {
			relPath, _ := filepath.Rel(baseDir, filePath)
			if relPath == "" {
				relPath = filepath.Base(filePath)
			}

			for _, u := range usages {
				section := ""
				if u.SectionTitle != "" {
					section = fmt.Sprintf(" (Section: \"%s\")", u.SectionTitle)
				}
				sb.WriteString(fmt.Sprintf("  - %s:%d%s\n", relPath, u.Line, section))
				sb.WriteString(fmt.Sprintf("    Context: %s\n", truncate(u.Context, 60)))
			}
		}
	}

	return sb.String()
}

// FormatAttributeList formats all attributes for display
func FormatAttributeList(attrs []parser.AttributeDefinition, baseDir string) string {
	var sb strings.Builder

	sb.WriteString("Defined Attributes:\n\n")

	for _, attr := range attrs {
		relPath, _ := filepath.Rel(baseDir, attr.FilePath)
		if relPath == "" {
			relPath = filepath.Base(attr.FilePath)
		}
		sb.WriteString(fmt.Sprintf("  %s = \"%s\" (%s:%d)\n",
			attr.Name, attr.Value, relPath, attr.Line))
	}

	return sb.String()
}

// GetAffectedSections returns unique sections affected by an attribute
func GetAffectedSections(impact *AttributeImpact) []string {
	seen := make(map[string]bool)
	var sections []string

	for _, u := range impact.Usages {
		if u.SectionTitle != "" && !seen[u.SectionTitle] {
			seen[u.SectionTitle] = true
			sections = append(sections, u.SectionTitle)
		}
	}

	return sections
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
