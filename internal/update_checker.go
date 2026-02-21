package scribe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ghRelease is the minimal GitHub release JSON shape needed for update checking.
type ghRelease struct {
	TagName     string `json:"tag_name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
}

// CheckForUpdate queries the GitHub releases API and compares the latest
// release tag against the current compiled version.
// Pass "" for baseURL to use the production GitHub API.
func CheckForUpdate(baseURL string) (*UpdateInfo, error) {
	// Dev builds should not trigger update notifications
	if Version == "dev" {
		return &UpdateInfo{
			CurrentVersion:  Version,
			LatestVersion:   "unknown",
			UpdateAvailable: false,
		}, nil
	}

	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	apiURL := fmt.Sprintf("%s/repos/usescrolls/scribe/releases/latest", baseURL)

	req, err := http.NewRequest("GET", apiURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Scribe-Skills-Manager")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var release ghRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release JSON: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(Version, "v")

	return &UpdateInfo{
		CurrentVersion:  Version,
		LatestVersion:   release.TagName,
		UpdateAvailable: compareSemver(current, latest) < 0,
		ReleaseURL:      release.HTMLURL,
		PublishedAt:     release.PublishedAt,
	}, nil
}

// compareSemver compares two semver strings (without "v" prefix).
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareSemver(a, b string) int {
	aParts := parseSemverParts(a)
	bParts := parseSemverParts(b)

	for i := range 3 {
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}
	return 0
}

// parseSemverParts splits "1.17.0" into [1, 17, 0].
// Pre-release suffixes (e.g. "1.17.0-dev", "1.17.0-rc1") are stripped
// so that "1.17.0-dev" parses as [1, 17, 0].
// Returns [0,0,0] on failure.
func parseSemverParts(v string) [3]int {
	var parts [3]int
	segments := strings.SplitN(v, ".", 3)
	for i, seg := range segments {
		if i >= 3 {
			break
		}
		// Strip pre-release suffix: "0-dev" -> "0", "0-rc1" -> "0"
		if idx := strings.IndexByte(seg, '-'); idx >= 0 {
			seg = seg[:idx]
		}
		n, err := strconv.Atoi(seg)
		if err != nil {
			return [3]int{}
		}
		parts[i] = n
	}
	return parts
}
