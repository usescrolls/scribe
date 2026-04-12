package scribe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newUpdateServer(t *testing.T, statusCode int, rel releaseManifest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases/latest" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(rel); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
}

func TestCheckForUpdate_UpdateAvailable(t *testing.T) {
	srv := newUpdateServer(t, http.StatusOK, releaseManifest{
		TagName:     "v99.0.0",
		HTMLURL:     "https://gitlab.com/usescrolls/scribe/-/releases/v99.0.0",
		PublishedAt: "2026-01-01T00:00:00Z",
	})
	defer srv.Close()

	old := Version
	Version = "1.0.0"
	defer func() { Version = old }()

	info, err := CheckForUpdate(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.UpdateAvailable {
		t.Error("expected UpdateAvailable to be true")
	}
	if info.LatestVersion != "v99.0.0" {
		t.Errorf("LatestVersion = %q, want %q", info.LatestVersion, "v99.0.0")
	}
	if info.CurrentVersion != "1.0.0" {
		t.Errorf("CurrentVersion = %q, want %q", info.CurrentVersion, "1.0.0")
	}
	if info.ReleaseURL == "" {
		t.Error("expected non-empty ReleaseURL")
	}
}

func TestCheckForUpdate_UpToDate(t *testing.T) {
	srv := newUpdateServer(t, http.StatusOK, releaseManifest{
		TagName:     "v1.0.0",
		HTMLURL:     "https://gitlab.com/usescrolls/scribe/-/releases/v1.0.0",
		PublishedAt: "2026-01-01T00:00:00Z",
	})
	defer srv.Close()

	old := Version
	Version = "1.0.0"
	defer func() { Version = old }()

	info, err := CheckForUpdate(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.UpdateAvailable {
		t.Error("expected UpdateAvailable to be false")
	}
}

func TestCheckForUpdate_CurrentNewer(t *testing.T) {
	srv := newUpdateServer(t, http.StatusOK, releaseManifest{
		TagName: "v1.0.0",
	})
	defer srv.Close()

	old := Version
	Version = "2.0.0"
	defer func() { Version = old }()

	info, err := CheckForUpdate(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.UpdateAvailable {
		t.Error("expected UpdateAvailable to be false when current is newer")
	}
}

func TestCheckForUpdate_DevVersion(t *testing.T) {
	old := Version
	Version = "dev"
	defer func() { Version = old }()

	info, err := CheckForUpdate("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.UpdateAvailable {
		t.Error("expected UpdateAvailable to be false for dev version")
	}
}

func TestCheckForUpdate_DevSuffixedVersion_SameRelease(t *testing.T) {
	srv := newUpdateServer(t, http.StatusOK, releaseManifest{
		TagName: "v1.17.0",
		HTMLURL: "https://gitlab.com/usescrolls/scribe/-/releases/v1.17.0",
	})
	defer srv.Close()

	old := Version
	Version = "1.17.0-dev"
	defer func() { Version = old }()

	info, err := CheckForUpdate(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.UpdateAvailable {
		t.Error("1.17.0-dev should NOT show update when latest is v1.17.0")
	}
}

func TestCheckForUpdate_DevSuffixedVersion_NewerRelease(t *testing.T) {
	srv := newUpdateServer(t, http.StatusOK, releaseManifest{
		TagName: "v1.18.0",
		HTMLURL: "https://gitlab.com/usescrolls/scribe/-/releases/v1.18.0",
	})
	defer srv.Close()

	old := Version
	Version = "1.17.0-dev"
	defer func() { Version = old }()

	info, err := CheckForUpdate(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.UpdateAvailable {
		t.Error("1.17.0-dev SHOULD show update when latest is v1.18.0")
	}
}

func TestCheckForUpdate_APIError(t *testing.T) {
	srv := newUpdateServer(t, http.StatusNotFound, releaseManifest{})
	defer srv.Close()

	old := Version
	Version = "1.0.0"
	defer func() { Version = old }()

	_, err := CheckForUpdate(srv.URL)
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestCheckForUpdate_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	old := Version
	Version = "1.0.0"
	defer func() { Version = old }()

	_, err := CheckForUpdate(srv.URL)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestBuildReleaseManifestURL(t *testing.T) {
	tests := []struct {
		baseURL string
		want    string
	}{
		{"https://cdn.example.com", "https://cdn.example.com/releases/latest"},
		{"https://cdn.example.com/", "https://cdn.example.com/releases/latest"},
		{"http://localhost:1234", "http://localhost:1234/releases/latest"},
	}
	for _, tt := range tests {
		got := buildReleaseManifestURL(tt.baseURL)
		if got != tt.want {
			t.Errorf("buildReleaseManifestURL(%q) = %q, want %q", tt.baseURL, got, tt.want)
		}
	}
}

// --- Semver tests ---

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.17.0", "1.18.0", -1},
		{"1.18.0", "1.17.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"10.0.0", "9.0.0", 1},
		{"0.0.0", "0.0.0", 0},
		{"1.17.0-dev", "1.17.0", 0},
		{"1.17.0-dev", "1.18.0", -1},
		{"1.18.0", "1.17.0-dev", 1},
	}

	for _, tt := range tests {
		got := compareSemver(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareSemver(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestParseSemverParts(t *testing.T) {
	tests := []struct {
		input string
		want  [3]int
	}{
		{"1.17.0", [3]int{1, 17, 0}},
		{"0.0.0", [3]int{0, 0, 0}},
		{"10.20.30", [3]int{10, 20, 30}},
		{"dev", [3]int{0, 0, 0}},
		{"abc.def.ghi", [3]int{0, 0, 0}},
		{"1", [3]int{1, 0, 0}},
		{"1.2", [3]int{1, 2, 0}},
		{"1.17.0-dev", [3]int{1, 17, 0}},
		{"1.17.0-rc1", [3]int{1, 17, 0}},
		{"0.0.0-dev", [3]int{0, 0, 0}},
		{"2.0.0-beta.1", [3]int{2, 0, 0}},
	}

	for _, tt := range tests {
		got := parseSemverParts(tt.input)
		if got != tt.want {
			t.Errorf("parseSemverParts(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
