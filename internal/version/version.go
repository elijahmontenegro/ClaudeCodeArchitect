package version

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	repoOwner    = "elijahmontenegro"
	repoName     = "ClaudeCodeArchitect"
	cacheFile    = ".cca-version-check"
	cacheTTL     = 24 * time.Hour
	githubAPIURL = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
)

type cacheEntry struct {
	Version   string    `json:"version"`
	CheckedAt time.Time `json:"checked_at"`
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckForUpdate checks if a newer version is available
// Returns the latest version if an update is available, empty string otherwise
func CheckForUpdate(currentVersion string) string {
	// Skip for dev builds
	if currentVersion == "dev" {
		return ""
	}

	// Check cache first
	if cached := readCache(); cached != nil {
		if time.Since(cached.CheckedAt) < cacheTTL {
			if isNewer(cached.Version, currentVersion) {
				return cached.Version
			}
			return ""
		}
	}

	// Fetch latest from GitHub
	latest := fetchLatestVersion()
	if latest == "" {
		return ""
	}

	// Update cache
	writeCache(latest)

	if isNewer(latest, currentVersion) {
		return latest
	}
	return ""
}

func fetchLatestVersion() string {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(githubAPIURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	return strings.TrimPrefix(release.TagName, "v")
}

func isNewer(latest, current string) bool {
	// Simple semver comparison (handles x.y.z format)
	latestParts := strings.Split(strings.TrimPrefix(latest, "v"), ".")
	currentParts := strings.Split(strings.TrimPrefix(current, "v"), ".")

	for i := 0; i < len(latestParts) && i < len(currentParts); i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return len(latestParts) > len(currentParts)
}

func getCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, cacheFile)
}

func readCache() *cacheEntry {
	path := getCachePath()
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}
	return &entry
}

func writeCache(version string) {
	path := getCachePath()
	if path == "" {
		return
	}

	entry := cacheEntry{
		Version:   version,
		CheckedAt: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	os.WriteFile(path, data, 0644)
}
