package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstall_WritesEveryFile(t *testing.T) {
	dir := t.TempDir()
	if IsInstalled(dir) || NeedsUpdate(dir) {
		t.Fatal("nothing is installed yet")
	}

	if err := Install(dir); err != nil {
		t.Fatal(err)
	}
	for _, name := range Files {
		data, err := os.ReadFile(filepath.Join(dir, skillDirName, name))
		if err != nil {
			t.Fatalf("%s not installed: %v", name, err)
		}
		if len(data) == 0 || string(data) != EmbeddedContent(name) {
			t.Errorf("%s installed with the wrong content", name)
		}
	}
	if !UpToDate(dir) || NeedsUpdate(dir) {
		t.Error("a fresh install should be up to date")
	}
}

func TestNeedsUpdate_WhenAnyFileDiffers(t *testing.T) {
	dir := t.TempDir()
	if err := Install(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, skillDirName, "writing.md"), []byte("old guide"), 0644); err != nil {
		t.Fatal(err)
	}
	if !NeedsUpdate(dir) {
		t.Error("a changed writing.md should need an update")
	}

	if err := os.Remove(filepath.Join(dir, skillDirName, "writing.md")); err != nil {
		t.Fatal(err)
	}
	if !NeedsUpdate(dir) {
		t.Error("a missing writing.md should need an update")
	}
}

func TestSkillPointsToItsGuide(t *testing.T) {
	if !strings.Contains(EmbeddedContent("SKILL.md"), "writing.md") {
		t.Error("SKILL.md should tell the agent when to read writing.md")
	}
}

func TestRefresh_UpdatesStaleCopiesOnly(t *testing.T) {
	stale, current, absent := t.TempDir(), t.TempDir(), t.TempDir()
	for _, dir := range []string{stale, current} {
		if err := Install(dir); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(stale, skillDirName, "SKILL.md"), []byte("old instructions"), 0644); err != nil {
		t.Fatal(err)
	}

	updated, errs := Refresh([]string{stale, current, absent})
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	if len(updated) != 1 || updated[0] != stale {
		t.Fatalf("updated = %q, want only the stale copy", updated)
	}
	if !UpToDate(stale) {
		t.Error("the stale copy wasn't brought up to date")
	}
	if IsInstalled(absent) {
		t.Error("Refresh installed the skill where it wasn't installed")
	}
}
