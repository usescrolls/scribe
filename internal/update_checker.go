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

// releaseAsset is a normalized release asset (platform binary).
type releaseAsset struct {
	Name        string
	DownloadURL string
}

// release is the normalized release metadata used by update checking and self-update.
type release struct {
	TagName     string
	URL         string
	PublishedAt string
	Assets      []releaseAsset
}

// buildReleaseManifestURL returns the CDN manifest URL for the given base URL.
func buildReleaseManifestURL(baseURL string) string {
	if baseURL == "" {
		baseURL = PublicDownloadBase
	}
	return fmt.Sprintf("%s/releases/latest", strings.TrimSuffix(baseURL, "/"))
}

// fetchLatestRelease fetches the latest release manifest from the CDN.
// Pass "" for baseURL to use the configured PublicDownloadBase.
func fetchLatestRelease(baseURL string) (*release, error) {
	req, err := http.NewRequest("GET", buildReleaseManifestURL(baseURL), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("User-Agent", "Scribe-Skills-Manager")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("release manifest request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release manifest returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return parseReleaseManifest(body)
}

// --- Release manifest JSON ---

type releaseManifestAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type releaseManifest struct {
	TagName     string                 `json:"tag_name"`
	HTMLURL     string                 `json:"html_url"`
	PublishedAt string                 `json:"published_at"`
	Assets      []releaseManifestAsset `json:"assets"`
}

func parseReleaseManifest(data []byte) (*release, error) {
	var manifest releaseManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse release manifest: %w", err)
	}
	r := &release{
		TagName:     manifest.TagName,
		URL:         manifest.HTMLURL,
		PublishedAt: manifest.PublishedAt,
	}
	for _, a := range manifest.Assets {
		r.Assets = append(r.Assets, releaseAsset{
			Name:        a.Name,
			DownloadURL: a.BrowserDownloadURL,
		})
	}
	return r, nil
}

// --- Update check (public API) ---

// CheckForUpdate queries the configured CDN release manifest and compares the
// latest release tag against the installed product version.
// Pass "" for overrideURL to use PublicDownloadBase.
func CheckForUpdate(overrideURL string) (*UpdateInfo, error) {
	if Version == "dev" {
		return &UpdateInfo{
			CurrentVersion:  Version,
			LatestVersion:   "unknown",
			UpdateAvailable: false,
		}, nil
	}

	currentVersion := Version
	if manifest, err := ReadInstallManifest(); err == nil {
		if manifest.Version != "" {
			currentVersion = manifest.Version
		}
		if overrideURL == "" && manifest.PublicDownloadBase != "" {
			overrideURL = manifest.PublicDownloadBase
		}
	}

	rel, err := fetchLatestRelease(overrideURL)
	if err != nil {
		return nil, err
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	return &UpdateInfo{
		CurrentVersion:  currentVersion,
		LatestVersion:   rel.TagName,
		UpdateAvailable: compareSemver(current, latest) < 0,
		ReleaseURL:      rel.URL,
		PublishedAt:     rel.PublishedAt,
	}, nil
}

// --- Semver helpers ---

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

func parseSemverParts(v string) [3]int {
	var parts [3]int
	segments := strings.SplitN(v, ".", 3)
	for i, seg := range segments {
		if i >= 3 {
			break
		}
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
