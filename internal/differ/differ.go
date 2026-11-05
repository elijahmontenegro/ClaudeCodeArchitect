package differ

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/compiler"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// DiffResult represents the result of comparing compiled specs
type DiffResult struct {
	OldCommit      string
	NewCommit      string
	OldCommitShort string
	NewCommitShort string
	ChangedFiles   []string        // Source files that changed
	UnifiedDiff    string          // Unified diff of compiled output
	SectionChanges []SectionChange // Per-section breakdown
	HasChanges     bool
}

// SectionChange represents changes in a specific section
type SectionChange struct {
	SectionTitle string
	ChangeType   string // "added", "removed", "modified"
	AddedLines   int
	RemovedLines int
}

// DiffCompiled compares compiled output between current and a previous commit
func DiffCompiled(manifestPath, targetCommit string) (*DiffResult, error) {
	if !IsGitRepository() {
		return nil, fmt.Errorf("not in a git repository")
	}

	// Resolve commits
	currentCommit, err := GetCurrentCommit()
	if err != nil {
		return nil, err
	}

	oldCommit, err := ResolveCommit(targetCommit)
	if err != nil {
		return nil, err
	}

	result := &DiffResult{
		OldCommit: oldCommit,
		NewCommit: currentCommit,
	}

	result.OldCommitShort, _ = GetCommitShort(oldCommit)
	result.NewCommitShort, _ = GetCommitShort(currentCommit)

	// Get changed source files
	changedFiles, err := GetChangedFiles(oldCommit, currentCommit)
	if err != nil {
		return nil, err
	}
	result.ChangedFiles = filterAdocFiles(changedFiles)

	// Compile current version
	currentOutput, err := compiler.Compile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compile current spec: %v", err)
	}

	// Create worktree for old version and compile
	worktreePath, err := CreateWorktree(oldCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to create worktree: %v", err)
	}
	defer RemoveWorktree(worktreePath)

	// Find manifest in worktree
	oldManifestPath := filepath.Join(worktreePath, getRelativeManifestPath(manifestPath))
	oldOutput, err := compiler.Compile(oldManifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compile old spec: %v", err)
	}

	// Generate diff
	result.UnifiedDiff = generateUnifiedDiff(oldOutput, currentOutput, result.OldCommitShort, result.NewCommitShort)
	result.HasChanges = oldOutput != currentOutput

	// Analyze section changes
	result.SectionChanges = analyzeSectionChanges(oldOutput, currentOutput)

	return result, nil
}

// generateUnifiedDiff creates a unified diff between two strings
func generateUnifiedDiff(old, new, oldLabel, newLabel string) string {
	dmp := diffmatchpatch.New()

	// Get line-based diff
	a, b, c := dmp.DiffLinesToChars(old, new)
	diffs := dmp.DiffMain(a, b, false)
	diffs = dmp.DiffCharsToLines(diffs, c)

	// Convert to unified format
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- %s\n", oldLabel))
	sb.WriteString(fmt.Sprintf("+++ %s\n", newLabel))

	lineNum := 0
	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			lineNum++
			switch diff.Type {
			case diffmatchpatch.DiffDelete:
				sb.WriteString(fmt.Sprintf("-%s\n", line))
			case diffmatchpatch.DiffInsert:
				sb.WriteString(fmt.Sprintf("+%s\n", line))
			case diffmatchpatch.DiffEqual:
				// Don't include unchanged lines to keep diff readable
			}
		}
	}

	return sb.String()
}

// analyzeSectionChanges determines which sections were modified
func analyzeSectionChanges(old, new string) []SectionChange {
	oldSections := extractSectionBlocks(old)
	newSections := extractSectionBlocks(new)

	var changes []SectionChange

	// Check for modified and removed sections
	for title, oldContent := range oldSections {
		if newContent, exists := newSections[title]; exists {
			if oldContent != newContent {
				added, removed := countChangedLines(oldContent, newContent)
				changes = append(changes, SectionChange{
					SectionTitle: title,
					ChangeType:   "modified",
					AddedLines:   added,
					RemovedLines: removed,
				})
			}
		} else {
			changes = append(changes, SectionChange{
				SectionTitle: title,
				ChangeType:   "removed",
				RemovedLines: strings.Count(oldContent, "\n") + 1,
			})
		}
	}

	// Check for added sections
	for title, newContent := range newSections {
		if _, exists := oldSections[title]; !exists {
			changes = append(changes, SectionChange{
				SectionTitle: title,
				ChangeType:   "added",
				AddedLines:   strings.Count(newContent, "\n") + 1,
			})
		}
	}

	return changes
}

// extractSectionBlocks extracts sections from markdown content
func extractSectionBlocks(content string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(content, "\n")

	var currentTitle string
	var currentContent strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			// Save previous section
			if currentTitle != "" {
				sections[currentTitle] = currentContent.String()
			}
			// Start new section
			currentTitle = strings.TrimLeft(line, "# ")
			currentContent.Reset()
		} else if currentTitle != "" {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	// Save last section
	if currentTitle != "" {
		sections[currentTitle] = currentContent.String()
	}

	return sections
}

// countChangedLines counts added and removed lines between two strings
func countChangedLines(old, new string) (added, removed int) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(old, new, false)

	for _, diff := range diffs {
		lineCount := strings.Count(diff.Text, "\n")
		if diff.Text != "" && !strings.HasSuffix(diff.Text, "\n") {
			lineCount++
		}

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			added += lineCount
		case diffmatchpatch.DiffDelete:
			removed += lineCount
		}
	}

	return added, removed
}

// filterAdocFiles filters to only .adoc files
func filterAdocFiles(files []string) []string {
	var result []string
	for _, f := range files {
		if strings.HasSuffix(f, ".adoc") {
			result = append(result, f)
		}
	}
	return result
}

// getRelativeManifestPath gets the relative path of manifest from git root
func getRelativeManifestPath(manifestPath string) string {
	gitRoot, err := GetGitRoot()
	if err != nil {
		return filepath.Base(manifestPath)
	}

	relPath, err := filepath.Rel(gitRoot, manifestPath)
	if err != nil {
		return filepath.Base(manifestPath)
	}

	return relPath
}

// FormatDiffResult formats the diff result for display
func FormatDiffResult(result *DiffResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Comparing: %s -> %s\n\n", result.OldCommitShort, result.NewCommitShort))

	if !result.HasChanges {
		sb.WriteString("No changes in compiled output.\n")
		return sb.String()
	}

	if len(result.ChangedFiles) > 0 {
		sb.WriteString("Changed source files:\n")
		for _, f := range result.ChangedFiles {
			sb.WriteString(fmt.Sprintf("  - %s\n", f))
		}
		sb.WriteString("\n")
	}

	if len(result.SectionChanges) > 0 {
		sb.WriteString("Section changes:\n")
		for _, sc := range result.SectionChanges {
			switch sc.ChangeType {
			case "added":
				sb.WriteString(fmt.Sprintf("  + %s (+%d lines)\n", sc.SectionTitle, sc.AddedLines))
			case "removed":
				sb.WriteString(fmt.Sprintf("  - %s (-%d lines)\n", sc.SectionTitle, sc.RemovedLines))
			case "modified":
				sb.WriteString(fmt.Sprintf("  ~ %s (+%d/-%d lines)\n", sc.SectionTitle, sc.AddedLines, sc.RemovedLines))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Diff:\n")
	sb.WriteString(result.UnifiedDiff)

	return sb.String()
}
