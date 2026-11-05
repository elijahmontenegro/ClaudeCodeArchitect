package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SpecConfig represents the .spec.yaml configuration file
type SpecConfig struct {
	Spec string `yaml:"spec"`

	// Profile names the kind of doc set, which decides what semantic
	// validation checks. Empty means general docs: structural checks only.
	Profile string `yaml:"profile"`
}

// ProfileArchitecture is the profile for architecture specifications, which
// `cca validate` checks against the completeness checklist.
const ProfileArchitecture = "architecture"

// LoadProfile returns the profile named in dir's .spec.yaml, or "" when there
// is no .spec.yaml or it names none. An unknown profile is an error.
func LoadProfile(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, ".spec.yaml"))
	if err != nil {
		return "", nil
	}
	var config SpecConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("invalid .spec.yaml: %v", err)
	}
	switch config.Profile {
	case "", ProfileArchitecture:
		return config.Profile, nil
	default:
		return "", fmt.Errorf("unknown profile %q in .spec.yaml (known: %s)", config.Profile, ProfileArchitecture)
	}
}

// FindSpec discovers the specification file location in the current directory
func FindSpec() (string, error) {
	return FindSpecInDir(".")
}

// FindSpecInDir discovers the specification file location in the given directory
// It checks .spec.yaml first, then falls back to conventions
func FindSpecInDir(dir string) (string, error) {
	// 1. Check for .spec.yaml config
	configPath := filepath.Join(dir, ".spec.yaml")
	if data, err := os.ReadFile(configPath); err == nil {
		var config SpecConfig
		// A .spec.yaml that sets only a profile leaves the spec to convention.
		if err := yaml.Unmarshal(data, &config); err == nil && config.Spec != "" {
			path := config.Spec
			if !filepath.IsAbs(path) {
				path = filepath.Join(dir, path)
			}
			if _, err := os.Stat(path); err == nil {
				absPath, _ := filepath.Abs(path)
				return absPath, nil
			}
			return "", fmt.Errorf("spec file not found at configured path: %s", path)
		}
	}

	// 2. Check conventions
	conventions := []string{
		"MANIFEST.adoc",
		"spec/MANIFEST.adoc",
		"plan/MANIFEST.adoc",
	}

	for _, name := range conventions {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath, nil
		}
	}

	return "", fmt.Errorf("spec not found in %s - checked: %v\nCreate .spec.yaml or use MANIFEST.adoc", dir, conventions)
}

// LoadSpecConfig loads .spec.yaml configuration from the current directory
func LoadSpecConfig() (*SpecConfig, error) {
	data, err := os.ReadFile(".spec.yaml")
	if err != nil {
		return nil, err
	}

	var config SpecConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetSpecRoot returns the directory containing the MANIFEST.adoc file
func GetSpecRoot(manifestPath string) string {
	return filepath.Dir(manifestPath)
}
