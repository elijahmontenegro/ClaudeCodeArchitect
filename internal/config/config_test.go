package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindSpecInDir_Priority(t *testing.T) {
	// .spec.yaml should take priority over convention

	dir := t.TempDir()

	// Create both .spec.yaml and MANIFEST.adoc
	specYaml := filepath.Join(dir, ".spec.yaml")
	manifest := filepath.Join(dir, "MANIFEST.adoc")
	customSpec := filepath.Join(dir, "custom.adoc")

	os.WriteFile(manifest, []byte("= Manifest"), 0644)
	os.WriteFile(customSpec, []byte("= Custom"), 0644)
	os.WriteFile(specYaml, []byte("spec: ./custom.adoc"), 0644)

	result, err := FindSpecInDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filepath.Base(result) != "custom.adoc" {
		t.Errorf("expected custom.adoc (from .spec.yaml), got %s", filepath.Base(result))
	}
}

func TestFindSpecInDir_ConventionOrder(t *testing.T) {
	// MANIFEST.adoc should be found before spec/MANIFEST.adoc

	dir := t.TempDir()

	// Create spec/MANIFEST.adoc first
	specDir := filepath.Join(dir, "spec")
	os.MkdirAll(specDir, 0755)
	os.WriteFile(filepath.Join(specDir, "MANIFEST.adoc"), []byte("= Spec"), 0644)

	// Should find spec/MANIFEST.adoc
	result, err := FindSpecInDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got %s", result)
	}

	// Now create root MANIFEST.adoc - should take priority
	os.WriteFile(filepath.Join(dir, "MANIFEST.adoc"), []byte("= Root"), 0644)

	result, err = FindSpecInDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Dir(result) != dir {
		t.Errorf("expected root MANIFEST.adoc to take priority, got %s", result)
	}
}

func TestFindSpecInDir_NotFound(t *testing.T) {
	dir := t.TempDir()

	_, err := FindSpecInDir(dir)
	if err == nil {
		t.Error("expected error for empty directory")
	}
}

func TestFindSpecInDir_SpecYamlBadPath(t *testing.T) {
	dir := t.TempDir()

	// .spec.yaml points to nonexistent file
	specYaml := filepath.Join(dir, ".spec.yaml")
	os.WriteFile(specYaml, []byte("spec: ./nonexistent.adoc"), 0644)

	_, err := FindSpecInDir(dir)
	if err == nil {
		t.Error("expected error for bad spec path")
	}
}

func TestFindSpecInDir_ReturnsAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "MANIFEST.adoc"), []byte("= Test"), 0644)

	result, err := FindSpecInDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got %s", result)
	}
}

func TestFindSpec_UsesCurrentDir(t *testing.T) {
	// FindSpec() should behave same as FindSpecInDir(".")
	// This is a sanity check that the wrapper works

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	dir := t.TempDir()
	os.Chdir(dir)
	os.WriteFile("MANIFEST.adoc", []byte("= Test"), 0644)

	result, err := FindSpec()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filepath.Base(result) != "MANIFEST.adoc" {
		t.Errorf("expected MANIFEST.adoc, got %s", result)
	}
}

func TestFindSpecInDir_ProfileOnlyConfigUsesConvention(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "MANIFEST.adoc"), []byte("= Manifest"), 0644)
	os.WriteFile(filepath.Join(dir, ".spec.yaml"), []byte("profile: architecture\n"), 0644)

	result, err := FindSpecInDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(result) != "MANIFEST.adoc" {
		t.Errorf("expected MANIFEST.adoc by convention, got %s", result)
	}
}

func TestLoadProfile(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string // "" means no .spec.yaml
		want    string
		wantErr bool
	}{
		{name: "no config", yaml: "", want: ""},
		{name: "no profile", yaml: "spec: ./MANIFEST.adoc\n", want: ""},
		{name: "architecture", yaml: "profile: architecture\n", want: ProfileArchitecture},
		{name: "unknown", yaml: "profile: novel\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.yaml != "" {
				os.WriteFile(filepath.Join(dir, ".spec.yaml"), []byte(tt.yaml), 0644)
			}
			got, err := LoadProfile(dir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
