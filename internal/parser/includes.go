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
	Partial    bool     // Includes only part of the file (tag=, tags= or lines=)
	SourceFile string   // File containing this include
}

// partialPattern matches include options that select part of a file.
var partialPattern = regexp.MustCompile(`(^|,)\s*(tags?|lines)\s*=`)

// isPartial reports whether include options select only part of the file.
func isPartial(options string) bool {
	return partialPattern.MatchString(options)
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
				Path:    path,
				Line:    lineNum,
				Partial: isPartial(options),
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

// GetIncludedFiles returns every file the manifest includes, directly or
// through other files, each once, in the order first included.
func GetIncludedFiles(manifestPath string) ([]string, error) {
	all, _, err := includedFiles(manifestPath)
	return all, err
}

// includedFiles walks the include tree from the manifest. It returns every
// included file, each once in order of first inclusion, and separately the
// files included whole somewhere, each once in order of first whole
// inclusion. Only a whole include places a file's sections in the document;
// a partial one (tag=, tags=, lines=) brings in a region of it, and so does
// anything included from inside that file.
func includedFiles(manifestPath string) (all, whole []string, err error) {
	root, err := filepath.Abs(manifestPath)
	if err != nil {
		return nil, nil, err
	}
	if _, err := os.Stat(root); err != nil {
		return nil, nil, err
	}

	seen := map[string]bool{}
	seenWhole := map[string]bool{}
	visited := map[bool]map[string]bool{false: {}, true: {}} // by "reached through a partial include"

	var walk func(path string, partial bool)
	walk = func(path string, partial bool) {
		if visited[partial][path] {
			return // already walked this way; also stops include cycles
		}
		visited[partial][path] = true

		includes, err := ExtractIncludesFromFile(path)
		if err != nil {
			return // a missing include is reported by the compile check
		}
		for _, inc := range includes {
			if !seen[inc.AbsPath] {
				seen[inc.AbsPath] = true
				all = append(all, inc.AbsPath)
			}
			incPartial := partial || inc.Partial
			if !incPartial && !seenWhole[inc.AbsPath] {
				seenWhole[inc.AbsPath] = true
				whole = append(whole, inc.AbsPath)
			}
			walk(inc.AbsPath, incPartial)
		}
	}
	walk(root, false)

	return all, whole, nil
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
