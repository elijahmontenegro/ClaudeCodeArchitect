package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// IncludeInfo represents an include directive
type IncludeInfo struct {
	Path       string   // Relative path (e.g., "core/types.adoc")
	AbsPath    string   // Absolute resolved path
	Line       int      // Line number in source file
	Tags       []string // Optional tag filters
	SourceFile string   // File containing this include
}

// Regex pattern for include directives
// Matches: include::path/to/file.adoc[] or include::path/to/file.adoc[tag=name]
var includePattern = regexp.MustCompile(`^include::([^\[]+)\[(.*)\]$`)

// ExtractIncludes extracts all include directives from content
func ExtractIncludes(content string) []IncludeInfo {
	var includes []IncludeInfo

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if matches := includePattern.FindStringSubmatch(line); matches != nil {
			path := matches[1]
			options := matches[2]

			info := IncludeInfo{
				Path: path,
				Line: lineNum,
			}

			// Parse tag options
			if strings.Contains(options, "tag=") {
				tagPattern := regexp.MustCompile(`tag=([a-zA-Z0-9_-]+)`)
				tagMatches := tagPattern.FindAllStringSubmatch(options, -1)
				for _, tm := range tagMatches {
					info.Tags = append(info.Tags, tm[1])
				}
			}

			includes = append(includes, info)
		}
	}

	return includes
}

// ExtractIncludesFromFile extracts includes from a file with resolved paths
func ExtractIncludesFromFile(filePath string) ([]IncludeInfo, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(filePath)
	includes := ExtractIncludes(string(content))

	// Resolve absolute paths and set source file
	for i := range includes {
		includes[i].SourceFile = filePath
		includes[i].AbsPath = ResolveIncludePath(baseDir, includes[i].Path)
	}

	return includes, nil
}

// ResolveIncludePath converts a relative include path to absolute
func ResolveIncludePath(baseDir, includePath string) string {
	if filepath.IsAbs(includePath) {
		return includePath
	}
	return filepath.Join(baseDir, includePath)
}

// GetIncludedFiles returns all files included by a manifest (recursively)
func GetIncludedFiles(manifestPath string) ([]string, error) {
	visited := make(map[string]bool)
	var files []string

	if err := collectIncludes(manifestPath, visited, &files); err != nil {
		return nil, err
	}

	return files, nil
}

// collectIncludes recursively collects included files
func collectIncludes(filePath string, visited map[string]bool, files *[]string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	if visited[absPath] {
		return nil // Already processed, avoid cycles
	}
	visited[absPath] = true

	includes, err := ExtractIncludesFromFile(absPath)
	if err != nil {
		return err
	}

	for _, inc := range includes {
		*files = append(*files, inc.AbsPath)
		// Recursively process included files
		if err := collectIncludes(inc.AbsPath, visited, files); err != nil {
			// File might not exist yet, skip
			continue
		}
	}

	return nil
}

// BuildIncludeTree builds a tree of include dependencies
type IncludeNode struct {
	Path     string
	AbsPath  string
	Includes []*IncludeNode
	Tags     []string
}

// BuildIncludeTree builds a tree of include dependencies starting from manifest
func BuildIncludeTree(manifestPath string) (*IncludeNode, error) {
	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return nil, err
	}

	visited := make(map[string]bool)
	return buildNodeRecursive(absPath, nil, visited)
}

func buildNodeRecursive(filePath string, tags []string, visited map[string]bool) (*IncludeNode, error) {
	if visited[filePath] {
		return nil, nil // Cycle detected
	}
	visited[filePath] = true

	node := &IncludeNode{
		Path:    filepath.Base(filePath),
		AbsPath: filePath,
		Tags:    tags,
	}

	includes, err := ExtractIncludesFromFile(filePath)
	if err != nil {
		return node, nil // File might not exist, return partial node
	}

	for _, inc := range includes {
		child, err := buildNodeRecursive(inc.AbsPath, inc.Tags, visited)
		if err != nil {
			continue
		}
		if child != nil {
			node.Includes = append(node.Includes, child)
		}
	}

	return node, nil
}
