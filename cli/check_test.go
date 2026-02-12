package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	scribe "github.com/usescrolls/scribe/internal"
)

// ---------------------------------------------------------------------------
// Tests for truncateHash (check.go)
// ---------------------------------------------------------------------------

func TestTruncateHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty string", input: "", expected: "-"},
		{name: "short hash", input: "abc123", expected: "abc123"},
		{name: "exactly 20 chars", input: "12345678901234567890", expected: "12345678901234567890"},
		{name: "21 chars truncated", input: "123456789012345678901", expected: "12345678901234567890..."},
		{name: "long sha256 hash", input: "sha256:abcdef1234567890abcdef1234567890abcdef1234567890", expected: "sha256:abcdef1234567..."},
		{name: "single char", input: "x", expected: "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateHash(tt.input)
			if result != tt.expected {
				t.Errorf("truncateHash(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for reconstructSource (check.go)
// ---------------------------------------------------------------------------

func TestReconstructSource(t *testing.T) {
	tests := []struct {
		name          string
		meta          *scribe.SkillMeta
		expectType    string
		expectOwner   string
		expectRepo    string
		expectRef     string
		expectURL     string
		expectSubpath string
		expectLocal   string
	}{
		{
			name: "github owner/repo",
			meta: &scribe.SkillMeta{
				Source:     "myowner/myrepo",
				SourceType: "github",
			},
			expectType:  "github",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectURL:   "https://github.com/myowner/myrepo",
		},
		{
			name: "github owner/repo with ref",
			meta: &scribe.SkillMeta{
				Source:     "myowner/myrepo#v2.0",
				SourceType: "github",
			},
			expectType:  "github",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectRef:   "v2.0",
			expectURL:   "https://github.com/myowner/myrepo",
		},
		{
			name: "github with existing URL",
			meta: &scribe.SkillMeta{
				Source:     "myowner/myrepo",
				SourceType: "github",
				SourceURL:  "https://github.com/myowner/myrepo",
			},
			expectType:  "github",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectURL:   "https://github.com/myowner/myrepo",
		},
		{
			name: "github with subpath in source",
			meta: &scribe.SkillMeta{
				Source:     "myowner/myrepo/skills/react",
				SourceType: "github",
			},
			expectType:    "github",
			expectOwner:   "myowner",
			expectRepo:    "myrepo",
			expectSubpath: "skills/react",
			expectURL:     "https://github.com/myowner/myrepo",
		},
		{
			name: "gitlab owner/repo",
			meta: &scribe.SkillMeta{
				Source:     "glowner/glrepo",
				SourceType: "gitlab",
			},
			expectType:  "gitlab",
			expectOwner: "glowner",
			expectRepo:  "glrepo",
			expectURL:   "https://gitlab.com/glowner/glrepo",
		},
		{
			name: "zip source",
			meta: &scribe.SkillMeta{
				Source:     "https://example.com/skills.zip",
				SourceType: "zip",
			},
			expectType: "zip",
			expectURL:  "https://example.com/skills.zip",
		},
		{
			name: "zip with existing URL",
			meta: &scribe.SkillMeta{
				Source:     "https://example.com/skills.zip",
				SourceType: "zip",
				SourceURL:  "https://example.com/skills.zip",
			},
			expectType: "zip",
			expectURL:  "https://example.com/skills.zip",
		},
		{
			name: "local source",
			meta: &scribe.SkillMeta{
				Source:     "/home/user/my-skills",
				SourceType: "local",
			},
			expectType:  "local",
			expectLocal: "/home/user/my-skills",
		},
		{
			name: "url type",
			meta: &scribe.SkillMeta{
				Source:     "https://example.com/my-skills",
				SourceType: "url",
			},
			expectType: "url",
			expectURL:  "https://example.com/my-skills",
		},
		{
			name: "well-known type",
			meta: &scribe.SkillMeta{
				Source:     "https://example.com",
				SourceType: "well-known",
			},
			expectType: "well-known",
			expectURL:  "https://example.com",
		},
		{
			name: "skillPath overrides subpath",
			meta: &scribe.SkillMeta{
				Source:     "myowner/myrepo",
				SourceType: "github",
				SkillPath:  "deep/path/to/skill",
			},
			expectType:    "github",
			expectOwner:   "myowner",
			expectRepo:    "myrepo",
			expectSubpath: "deep/path/to/skill",
			expectURL:     "https://github.com/myowner/myrepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scribe.ReconstructSource(tt.meta)
			if result.Type != tt.expectType {
				t.Errorf("Type = %q, expected %q", result.Type, tt.expectType)
			}
			if tt.expectOwner != "" && result.Owner != tt.expectOwner {
				t.Errorf("Owner = %q, expected %q", result.Owner, tt.expectOwner)
			}
			if tt.expectRepo != "" && result.Repo != tt.expectRepo {
				t.Errorf("Repo = %q, expected %q", result.Repo, tt.expectRepo)
			}
			if tt.expectRef != "" && result.Ref != tt.expectRef {
				t.Errorf("Ref = %q, expected %q", result.Ref, tt.expectRef)
			}
			if tt.expectURL != "" && result.URL != tt.expectURL {
				t.Errorf("URL = %q, expected %q", result.URL, tt.expectURL)
			}
			if tt.expectSubpath != "" && result.Subpath != tt.expectSubpath {
				t.Errorf("Subpath = %q, expected %q", result.Subpath, tt.expectSubpath)
			}
			if tt.expectLocal != "" && result.LocalPath != tt.expectLocal {
				t.Errorf("LocalPath = %q, expected %q", result.LocalPath, tt.expectLocal)
			}
		})
	}
}

func TestReconstructSourceUnknownType(t *testing.T) {
	meta := &scribe.SkillMeta{
		Source:     "something",
		SourceType: "unknown",
	}
	result := scribe.ReconstructSource(meta)
	if result.Type != "unknown" {
		t.Errorf("expected type 'unknown', got %s", result.Type)
	}
}

func TestReconstructSourceGitLabWithURL(t *testing.T) {
	meta := &scribe.SkillMeta{
		Source:     "owner/repo",
		SourceType: "gitlab",
		SourceURL:  "https://gitlab.com/owner/repo",
	}
	result := scribe.ReconstructSource(meta)
	if result.URL != "https://gitlab.com/owner/repo" {
		t.Errorf("expected existing URL to be preserved, got %s", result.URL)
	}
}

func TestReconstructSourceURLType(t *testing.T) {
	meta := &scribe.SkillMeta{
		Source:     "https://example.com/skills",
		SourceType: "url",
	}
	result := scribe.ReconstructSource(meta)
	if result.URL != "https://example.com/skills" {
		t.Errorf("expected URL to be set from source, got %s", result.URL)
	}
}

func TestReconstructSourceZipWithURL(t *testing.T) {
	meta := &scribe.SkillMeta{
		Source:     "https://example.com/skills.zip",
		SourceType: "zip",
		SourceURL:  "https://example.com/skills.zip",
	}
	result := scribe.ReconstructSource(meta)
	// URL should be preserved from SourceURL
	if result.URL != "https://example.com/skills.zip" {
		t.Errorf("expected URL from SourceURL, got %s", result.URL)
	}
}

// ---------------------------------------------------------------------------
// Tests for checkOutputJSON (check.go)
// ---------------------------------------------------------------------------

func TestCheckOutputJSON(t *testing.T) {
	tests := []struct {
		name            string
		results         []scribe.CheckResult
		expectOutdated  int
		expectUpToDate  int
		expectErrors    int
		expectTotal     int
		expectSubstring string
	}{
		{
			name:           "empty results",
			results:        []scribe.CheckResult{},
			expectOutdated: 0,
			expectUpToDate: 0,
			expectErrors:   0,
			expectTotal:    0,
		},
		{
			name: "mixed results",
			results: []scribe.CheckResult{
				{Name: "skill-a", NeedsUpdate: true, CurrentHash: "hash1", RemoteHash: "hash2"},
				{Name: "skill-b", NeedsUpdate: false, CurrentHash: "hash3", RemoteHash: "hash3"},
				{Name: "skill-c", Error: "some error"},
			},
			expectOutdated: 1,
			expectUpToDate: 1,
			expectErrors:   1,
			expectTotal:    3,
		},
		{
			name: "all up-to-date",
			results: []scribe.CheckResult{
				{Name: "skill-x", NeedsUpdate: false, CurrentHash: "aaa", RemoteHash: "aaa"},
				{Name: "skill-y", NeedsUpdate: false, CurrentHash: "bbb", RemoteHash: "bbb"},
			},
			expectOutdated: 0,
			expectUpToDate: 2,
			expectErrors:   0,
			expectTotal:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := checkOutputJSON(tt.results)

			_ = w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Parse the JSON to verify structure
			var parsed struct {
				Results []scribe.CheckResult `json:"results"`
				Summary struct {
					Total    int `json:"total"`
					Outdated int `json:"outdated"`
					UpToDate int `json:"upToDate"`
					Errors   int `json:"errors"`
				} `json:"summary"`
			}
			if err := json.Unmarshal([]byte(output), &parsed); err != nil {
				t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			if parsed.Summary.Total != tt.expectTotal {
				t.Errorf("summary.total = %d, expected %d", parsed.Summary.Total, tt.expectTotal)
			}
			if parsed.Summary.Outdated != tt.expectOutdated {
				t.Errorf("summary.outdated = %d, expected %d", parsed.Summary.Outdated, tt.expectOutdated)
			}
			if parsed.Summary.UpToDate != tt.expectUpToDate {
				t.Errorf("summary.upToDate = %d, expected %d", parsed.Summary.UpToDate, tt.expectUpToDate)
			}
			if parsed.Summary.Errors != tt.expectErrors {
				t.Errorf("summary.errors = %d, expected %d", parsed.Summary.Errors, tt.expectErrors)
			}
			if len(parsed.Results) != len(tt.results) {
				t.Errorf("results length = %d, expected %d", len(parsed.Results), len(tt.results))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for checkOutputTable (check.go)
// ---------------------------------------------------------------------------

func TestCheckOutputTable(t *testing.T) {
	// Save and restore quiet flag
	origQuiet := quiet
	defer func() { quiet = origQuiet }()

	t.Run("table with mixed results", func(t *testing.T) {
		quiet = false
		results := []scribe.CheckResult{
			{Name: "skill-a", NeedsUpdate: true, CurrentHash: "oldhash", RemoteHash: "newhash"},
			{Name: "skill-b", NeedsUpdate: false, CurrentHash: "samehash", RemoteHash: "samehash"},
			{Name: "skill-c", Error: "failed to fetch"},
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := checkOutputTable(results)

		_ = w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify header
		if !strings.Contains(output, "NAME") || !strings.Contains(output, "STATUS") {
			t.Error("expected table header with NAME and STATUS")
		}
		// Verify skill names appear
		if !strings.Contains(output, "skill-a") {
			t.Error("expected skill-a in output")
		}
		if !strings.Contains(output, "skill-b") {
			t.Error("expected skill-b in output")
		}
		if !strings.Contains(output, "skill-c") {
			t.Error("expected skill-c in output")
		}
		// Verify statuses
		if !strings.Contains(output, "outdated") {
			t.Error("expected 'outdated' status in output")
		}
		if !strings.Contains(output, "up-to-date") {
			t.Error("expected 'up-to-date' status in output")
		}
		if !strings.Contains(output, "error:") {
			t.Error("expected 'error:' status in output")
		}
		// Verify summary
		if !strings.Contains(output, "3 skill(s) checked") {
			t.Errorf("expected summary line, got: %s", output)
		}
	})

	t.Run("table in quiet mode", func(t *testing.T) {
		quiet = true
		results := []scribe.CheckResult{
			{Name: "skill-a", NeedsUpdate: false, CurrentHash: "x", RemoteHash: "x"},
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_ = checkOutputTable(results)

		_ = w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// In quiet mode, the summary line should not appear
		if strings.Contains(output, "skill(s) checked") {
			t.Error("quiet mode should suppress summary line")
		}
	})
}

func TestCheckOutputTableWithErrors(t *testing.T) {
	saveAndRestoreFlags(t)
	quiet = false

	results := []scribe.CheckResult{
		{Name: "err-skill-1", Error: "fetch failed"},
		{Name: "err-skill-2", Error: "timeout"},
	}

	output := captureStdout(t, func() {
		_ = checkOutputTable(results)
	})

	if !strings.Contains(output, "2 error(s)") {
		t.Errorf("expected '2 error(s)' in summary, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// Tests for checkSkill (check.go)
// ---------------------------------------------------------------------------

func TestCheckSkillNonexistent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	result := scribe.CheckSkillForUpdate("nonexistent-skill")
	if result.Error == "" {
		t.Error("expected error for nonexistent skill")
	}
	if !strings.Contains(result.Error, "failed to read skill") {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestCheckSkillNoMeta(t *testing.T) {
	tmpDir, cleanup := setupTempHome(t)
	defer cleanup()
	_ = tmpDir

	// Install a skill without metadata
	scrollsDir, _ := scribe.GetScrollsDir()
	skillDir := filepath.Join(scrollsDir, "no-meta-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	content := "---\nname: no-meta-skill\ndescription: A test skill\n---\n\nContent here.\n"
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644)
	// No .scribe-meta.json file

	result := scribe.CheckSkillForUpdate("no-meta-skill")
	if result.Error == "" {
		t.Error("expected error for skill without metadata")
	}
	if !strings.Contains(result.Error, "no metadata") {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestCheckSkillLocalSource(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	installFakeSkill(t, "local-skill", "A local skill", "local", "/some/local/path")

	result := scribe.CheckSkillForUpdate("local-skill")
	if result.Error == "" {
		t.Error("expected error for local source skill")
	}
	if !strings.Contains(result.Error, "local source") {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestCheckSkillGitHubSourceFetchFails(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// Install a skill with github source type pointing to invalid repo
	installFakeSkill(t, "github-check-skill", "GitHub check test", "github", "nonexistent-owner/nonexistent-repo")

	result := scribe.CheckSkillForUpdate("github-check-skill")
	// Should fail at the fetch step
	if result.Error == "" {
		t.Error("expected error when fetching nonexistent github repo")
	}
	if !strings.Contains(result.Error, "failed to fetch") {
		t.Errorf("expected fetch error, got: %s", result.Error)
	}
}

// ---------------------------------------------------------------------------
// Tests for runCheck (check.go)
// ---------------------------------------------------------------------------

func TestRunCheckNoSkillsInstalled(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "No skills installed") {
		t.Errorf("expected 'No skills installed' message, got: %s", output)
	}
}

func TestRunCheckNoSkillsQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = false

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "No skills installed") {
		t.Error("quiet mode should suppress 'No skills installed' message")
	}
}

func TestRunCheckSingleSkillNonexistent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = false

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{"nonexistent"})
		if err != nil {
			t.Errorf("unexpected error from runCheck: %v", err)
		}
	})

	// Should show an error in the table for the nonexistent skill
	if !strings.Contains(output, "nonexistent") {
		t.Errorf("expected skill name in output, got: %s", output)
	}
}

func TestRunCheckSingleSkillJSON(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = true

	// Install a local skill (which can't be checked remotely)
	installFakeSkill(t, "check-me", "Check me", "local", "/local/path")

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{"check-me"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	var parsed struct {
		Results []scribe.CheckResult `json:"results"`
		Summary struct {
			Total  int `json:"total"`
			Errors int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Summary.Total != 1 {
		t.Errorf("expected total 1, got %d", parsed.Summary.Total)
	}
	if parsed.Summary.Errors != 1 {
		t.Errorf("expected 1 error (local source), got %d", parsed.Summary.Errors)
	}
}

func TestRunCheckAllSkillsWithLocalSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false

	installFakeSkill(t, "skill-alpha", "Alpha skill", "local", "/path/alpha")
	installFakeSkill(t, "skill-beta", "Beta skill", "local", "/path/beta")

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Checking 2 skill(s)") {
		t.Errorf("expected 'Checking 2 skill(s)' in output, got: %s", output)
	}
	if !strings.Contains(output, "skill-alpha") {
		t.Errorf("expected 'skill-alpha' in output, got: %s", output)
	}
	if !strings.Contains(output, "skill-beta") {
		t.Errorf("expected 'skill-beta' in output, got: %s", output)
	}
}

func TestRunCheckAllSkillsJSON(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = true

	installFakeSkill(t, "json-skill", "JSON check skill", "local", "/local/path")

	output := captureStdout(t, func() {
		err := runCheck(checkCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	var parsed struct {
		Results []scribe.CheckResult `json:"results"`
		Summary struct {
			Total int `json:"total"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON output: %v\n%s", err, output)
	}
	if parsed.Summary.Total != 1 {
		t.Errorf("expected 1 skill, got %d", parsed.Summary.Total)
	}
}
