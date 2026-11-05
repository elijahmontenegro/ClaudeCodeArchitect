package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// File is the optional config file in a doc set's root directory.
const File = ".cca.yaml"

// Config is a doc set's .cca.yaml.
type Config struct {
	// Entry is the document's entry point, for when it isn't found by convention.
	Entry string `yaml:"entry"`
}

// Conventions are where cca looks for the entry point, in order, when
// .cca.yaml doesn't name one.
var Conventions = []string{"MANIFEST.adoc", "docs/MANIFEST.adoc"}

// Load reads dir's .cca.yaml. Without one, it returns the zero Config. A key
// cca doesn't know is an error, so a misspelled setting isn't silently ignored.
func Load(dir string) (Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, File))
	if errors.Is(err, fs.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return Config{}, fmt.Errorf("invalid %s: %v", File, err)
	}
	return cfg, nil
}

// FindEntry returns the absolute path of the entry point of the doc set in
// dir: the file .cca.yaml names, or else the first convention that exists.
func FindEntry(dir string) (string, error) {
	cfg, err := Load(dir)
	if err != nil {
		return "", err
	}

	if cfg.Entry != "" {
		path := resolve(dir, cfg.Entry)
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("entry point %s, named in %s, not found", path, File)
		}
		return path, nil
	}

	for _, name := range Conventions {
		path := resolve(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no docs found in %s: checked %v\nCreate MANIFEST.adoc, or name the entry point in %s", dir, Conventions, File)
}

// resolve makes path absolute, relative to dir when it isn't already.
func resolve(dir, path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
