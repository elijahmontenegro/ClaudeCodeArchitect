package differ

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetCurrentCommit returns the current HEAD commit hash
func GetCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCommitShort returns the short hash for a commit
func GetCommitShort(commit string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", commit)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get short commit: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ResolveCommit resolves a commit reference (HEAD~1, branch name, etc.) to a hash
func ResolveCommit(ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", ref)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to resolve commit '%s': %v", ref, err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetFileAtCommit retrieves file content at a specific commit
func GetFileAtCommit(commit, filePath string) (string, error) {
	// Make path relative to git root
	relPath, err := getRelativeToGitRoot(filePath)
	if err != nil {
		relPath = filePath
	}

	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", commit, relPath))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get file at commit: %v", err)
	}
	return string(output), nil
}

// CreateWorktree creates a temporary git worktree for a specific commit
func CreateWorktree(commit string) (string, error) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "cca-diff-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Create worktree
	cmd := exec.Command("git", "worktree", "add", "--detach", tempDir, commit)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to create worktree: %v", err)
	}

	return tempDir, nil
}

// RemoveWorktree removes a git worktree
func RemoveWorktree(path string) error {
	// Remove from git worktree list
	cmd := exec.Command("git", "worktree", "remove", "--force", path)
	cmd.Run() // Ignore errors, cleanup anyway

	// Remove the directory
	return os.RemoveAll(path)
}

// GetGitRoot returns the root directory of the git repository
func GetGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// IsGitRepository checks if we're in a git repository
func IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// getRelativeToGitRoot converts an absolute path to relative to git root
func getRelativeToGitRoot(absPath string) (string, error) {
	gitRoot, err := GetGitRoot()
	if err != nil {
		return "", err
	}

	relPath, err := filepath.Rel(gitRoot, absPath)
	if err != nil {
		return "", err
	}

	// Convert to forward slashes for git
	return filepath.ToSlash(relPath), nil
}

// GetChangedFiles returns files changed between two commits
func GetChangedFiles(oldCommit, newCommit string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", oldCommit, newCommit)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// GetCommitMessage returns the commit message for a commit
func GetCommitMessage(commit string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%s", commit)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit message: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
