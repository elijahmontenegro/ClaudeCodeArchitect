package skill

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

// The skill is a folder: SKILL.md, which an agent loads, and writing.md, the
// guide it reads when it writes or restructures docs.
//
//go:embed SKILL.md writing.md
var embedded embed.FS

// Files are the skill's files, installed side by side.
var Files = []string{"SKILL.md", "writing.md"}

const skillDirName = "adoc"

// EmbeddedContent returns the content of one of the skill's files.
func EmbeddedContent(name string) string {
	data, _ := embedded.ReadFile(name)
	return string(data)
}

// IsInstalled checks if the skill exists in the given skills directory
func IsInstalled(skillsDir string) bool {
	_, err := os.Stat(filepath.Join(skillsDir, skillDirName, "SKILL.md"))
	return err == nil
}

// UpToDate reports whether every skill file in the given skills directory
// matches this version of cca.
func UpToDate(skillsDir string) bool {
	for _, name := range Files {
		data, err := os.ReadFile(filepath.Join(skillsDir, skillDirName, name))
		if err != nil || string(data) != EmbeddedContent(name) {
			return false
		}
	}
	return true
}

// Install writes the skill's files to the given skills directory
func Install(skillsDir string) error {
	targetDir := filepath.Join(skillsDir, skillDirName)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	for _, name := range Files {
		if err := os.WriteFile(filepath.Join(targetDir, name), []byte(EmbeddedContent(name)), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", name, err)
		}
	}
	return nil
}

// NeedsUpdate reports whether the skill is installed but differs from this
// version of cca.
func NeedsUpdate(skillsDir string) bool {
	return IsInstalled(skillsDir) && !UpToDate(skillsDir)
}

// Refresh brings every installed copy of the skill among skillsDirs in step
// with this version of cca, and returns the directories it rewrote. The skill
// describes cca's commands, so a copy from another version gives the agent
// reading it wrong instructions. A directory without the skill is left alone:
// installing is `cca skill`'s job.
func Refresh(skillsDirs []string) (updated []string, errs []error) {
	for _, dir := range skillsDirs {
		if !NeedsUpdate(dir) {
			continue
		}
		if err := Install(dir); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", dir, err))
			continue
		}
		updated = append(updated, dir)
	}
	return updated, errs
}

// GetProjectSkillDir returns .claude/skills in current directory
func GetProjectSkillDir() string {
	return filepath.Join(".claude", "skills")
}

// GetGlobalSkillDir returns ~/.claude/skills
func GetGlobalSkillDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "skills"), nil
}
