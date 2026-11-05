package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFindEntry_ConfigBeatsConvention(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{
		"MANIFEST.adoc":   "= Manifest",
		"docs/index.adoc": "= Index",
		File:              "entry: docs/index.adoc\n",
	})

	got, err := FindEntry(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "docs", "index.adoc"); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestFindEntry_ConventionOrder(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{"docs/MANIFEST.adoc": "= Docs"})

	got, err := FindEntry(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "docs", "MANIFEST.adoc"); got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	write(t, dir, map[string]string{"MANIFEST.adoc": "= Root"})
	got, err = FindEntry(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "MANIFEST.adoc"); got != want {
		t.Errorf("root MANIFEST.adoc should win, got %s", got)
	}
}

func TestFindEntry_ConfigWithoutEntryUsesConvention(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{"MANIFEST.adoc": "= Manifest", File: "# no settings\n"})

	got, err := FindEntry(dir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "MANIFEST.adoc" {
		t.Errorf("got %s, want MANIFEST.adoc by convention", got)
	}
}

func TestFindEntry_Errors(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		wantErr string
	}{
		{name: "nothing found", files: nil, wantErr: "no docs found"},
		{name: "entry missing", files: map[string]string{File: "entry: gone.adoc\n"}, wantErr: "not found"},
		{name: "unknown key", files: map[string]string{File: "spec: MANIFEST.adoc\n"}, wantErr: "invalid .cca.yaml"},
		{name: "retired profile key", files: map[string]string{File: "profile: architecture\n"}, wantErr: "invalid .cca.yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, tt.files)
			_, err := FindEntry(dir)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoad_EmptyOrMissingFileIsZeroConfig(t *testing.T) {
	dir := t.TempDir()
	if cfg, err := Load(dir); err != nil || cfg != (Config{}) {
		t.Fatalf("missing file: cfg=%+v err=%v", cfg, err)
	}
	write(t, dir, map[string]string{File: ""})
	if cfg, err := Load(dir); err != nil || cfg != (Config{}) {
		t.Fatalf("empty file: cfg=%+v err=%v", cfg, err)
	}
}
