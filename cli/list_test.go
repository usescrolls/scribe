package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

// ---------------------------------------------------------------------------
// Tests for filterSkills (list.go)
// ---------------------------------------------------------------------------

// TestFilterSkills tests the filterSkills function
func TestFilterSkills(t *testing.T) {
	skills := []*scribe.Skill{
		{Name: "react-best-practices", Description: "React patterns"},
		{Name: "typescript-patterns", Description: "TypeScript tips"},
		{Name: "go-patterns", Description: "Go idioms"},
	}

	t.Run("filter single skill", func(t *testing.T) {
		filtered := scribe.FilterSkillsByName(skills, []string{"react-best-practices"})
		if len(filtered) != 1 {
			t.Errorf("expected 1 skill, got %d", len(filtered))
		}
		if filtered[0].Name != "react-best-practices" {
			t.Errorf("expected react-best-practices, got %s", filtered[0].Name)
		}
	})

	t.Run("filter multiple skills", func(t *testing.T) {
		filtered := scribe.FilterSkillsByName(skills, []string{"react-best-practices", "go-patterns"})
		if len(filtered) != 2 {
			t.Errorf("expected 2 skills, got %d", len(filtered))
		}
	})

	t.Run("filter non-existent skill", func(t *testing.T) {
		filtered := scribe.FilterSkillsByName(skills, []string{"non-existent"})
		if len(filtered) != 0 {
			t.Errorf("expected 0 skills, got %d", len(filtered))
		}
	})
}

// ---------------------------------------------------------------------------
// Tests for truncateString (list.go)
// ---------------------------------------------------------------------------

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{name: "short string fits", input: "hello", maxLen: 10, expected: "hello"},
		{name: "exact length", input: "hello", maxLen: 5, expected: "hello"},
		{name: "needs truncation", input: "hello world", maxLen: 8, expected: "hello..."},
		{name: "maxLen 3 no ellipsis", input: "hello", maxLen: 3, expected: "hel"},
		{name: "maxLen 2 no ellipsis", input: "hello", maxLen: 2, expected: "he"},
		{name: "maxLen 1 no ellipsis", input: "hello", maxLen: 1, expected: "h"},
		{name: "maxLen 4 with ellipsis", input: "hello", maxLen: 4, expected: "h..."},
		{name: "empty string", input: "", maxLen: 10, expected: ""},
		{name: "long description", input: "This is a very long description that should be truncated", maxLen: 20, expected: "This is a very lo..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, expected %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for formatInstalledAt (list.go)
// ---------------------------------------------------------------------------

func TestFormatInstalledAt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty string", input: "", expected: "-"},
		{name: "full ISO8601", input: "2024-01-15T10:30:00Z", expected: "2024-01-15"},
		{name: "date only", input: "2024-01-15", expected: "2024-01-15"},
		{name: "short timestamp", input: "2024-01", expected: "2024-01"},
		{name: "very short", input: "2024", expected: "2024"},
		{name: "exactly 10 chars", input: "2024-01-15", expected: "2024-01-15"},
		{name: "longer than 10", input: "2024-01-15T00:00:00+05:30", expected: "2024-01-15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatInstalledAt(tt.input)
			if result != tt.expected {
				t.Errorf("formatInstalledAt(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for formatSkillSource (list.go)
// ---------------------------------------------------------------------------

// TestFormatSkillSource tests the formatSkillSource function
func TestFormatSkillSource(t *testing.T) {
	tests := []struct {
		name     string
		info     scribe.SkillInfo
		expected string
	}{
		{
			name: "github source",
			info: scribe.SkillInfo{
				Source:     "owner/repo",
				SourceType: "github",
			},
			expected: "github:owner/repo",
		},
		{
			name: "gitlab source",
			info: scribe.SkillInfo{
				Source:     "owner/repo",
				SourceType: "gitlab",
			},
			expected: "gitlab:owner/repo",
		},
		{
			name: "local source",
			info: scribe.SkillInfo{
				Source:     "/path/to/skill",
				SourceType: "local",
			},
			expected: "local:/path/to/skill",
		},
		{
			name: "zip source",
			info: scribe.SkillInfo{
				Source:     "https://example.com/plugin.zip",
				SourceType: "zip",
			},
			expected: "https://example.com/plugin.zip",
		},
		{
			name: "empty source",
			info: scribe.SkillInfo{
				Source:     "",
				SourceType: "",
			},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSkillSource(tt.info)
			if result != tt.expected {
				t.Errorf("formatSkillSource() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestFormatSkillSourceURL(t *testing.T) {
	info := scribe.SkillInfo{
		Source:     "https://example.com/skills",
		SourceType: "url",
	}
	result := formatSkillSource(info)
	if result != "https://example.com/skills" {
		t.Errorf("expected URL directly, got %s", result)
	}
}

func TestFormatSkillSourceUnknownType(t *testing.T) {
	info := scribe.SkillInfo{
		Source:     "custom-source",
		SourceType: "custom",
	}
	result := formatSkillSource(info)
	if result != "custom-source" {
		t.Errorf("expected raw source for unknown type, got %s", result)
	}
}

// ---------------------------------------------------------------------------
// Tests for listSkillsJSON (list.go)
// ---------------------------------------------------------------------------

func TestListSkillsJSON(t *testing.T) {
	skills := []scribe.SkillInfo{
		{
			Name:        "react-patterns",
			Description: "React best practices",
			Source:      "owner/repo",
			SourceType:  "github",
			InstalledAt: "2024-06-15T10:00:00Z",
			Agents:      []string{"claude-code"},
		},
		{
			Name:        "go-idioms",
			Description: "Go idiomatic patterns",
			Source:      "/local/path",
			SourceType:  "local",
			InstalledAt: "2024-07-01T12:00:00Z",
			Agents:      []string{"cursor", "claude-code"},
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsJSON(skills)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Skills []struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Source      string   `json:"source"`
			SourceType  string   `json:"sourceType"`
			InstalledAt string   `json:"installedAt"`
			Agents      []string `json:"agents"`
		} `json:"skills"`
		Count int `json:"count"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Count != 2 {
		t.Errorf("count = %d, expected 2", parsed.Count)
	}
	if len(parsed.Skills) != 2 {
		t.Errorf("skills length = %d, expected 2", len(parsed.Skills))
	}
	if parsed.Skills[0].Name != "react-patterns" {
		t.Errorf("first skill name = %q, expected %q", parsed.Skills[0].Name, "react-patterns")
	}
	if parsed.Skills[1].SourceType != "local" {
		t.Errorf("second skill sourceType = %q, expected %q", parsed.Skills[1].SourceType, "local")
	}
}

func TestListSkillsJSONEmpty(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsJSON([]scribe.SkillInfo{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Skills []any `json:"skills"`
		Count  int   `json:"count"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Count != 0 {
		t.Errorf("count = %d, expected 0", parsed.Count)
	}
	// skills should be an empty array, not null
	if parsed.Skills == nil {
		t.Error("expected empty array for skills, got null")
	}
}

// ---------------------------------------------------------------------------
// Tests for listSkillsNamesOnly (list.go)
// ---------------------------------------------------------------------------

func TestListSkillsNamesOnly(t *testing.T) {
	skills := []scribe.SkillInfo{
		{Name: "alpha-skill"},
		{Name: "beta-skill"},
		{Name: "gamma-skill"},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsNamesOnly(skills)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "alpha-skill" {
		t.Errorf("line 0 = %q, expected %q", lines[0], "alpha-skill")
	}
	if lines[1] != "beta-skill" {
		t.Errorf("line 1 = %q, expected %q", lines[1], "beta-skill")
	}
	if lines[2] != "gamma-skill" {
		t.Errorf("line 2 = %q, expected %q", lines[2], "gamma-skill")
	}
}

func TestListSkillsNamesOnlyEmpty(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsNamesOnly([]scribe.SkillInfo{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// Tests for listSkillsTable (list.go)
// ---------------------------------------------------------------------------

func TestListSkillsTableWithSkills(t *testing.T) {
	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = false

	skills := []scribe.SkillInfo{
		{
			Name:        "react-patterns",
			Description: "React best practices and patterns",
			Source:      "owner/repo",
			SourceType:  "github",
			InstalledAt: "2024-06-15T10:00:00Z",
			Agents:      []string{"claude-code", "cursor"},
		},
		{
			Name:        "go-idioms",
			Description: "Go idiomatic patterns",
			Source:      "/local/path",
			SourceType:  "local",
			InstalledAt: "2024-07-01",
			Agents:      []string{},
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsTable(skills)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check header
	if !strings.Contains(output, "NAME") {
		t.Error("expected NAME header")
	}
	if !strings.Contains(output, "DESCRIPTION") {
		t.Error("expected DESCRIPTION header")
	}
	if !strings.Contains(output, "SOURCE") {
		t.Error("expected SOURCE header")
	}
	// Check skill data
	if !strings.Contains(output, "react-patterns") {
		t.Error("expected react-patterns in output")
	}
	if !strings.Contains(output, "go-idioms") {
		t.Error("expected go-idioms in output")
	}
	// Check installed count
	if !strings.Contains(output, "2 skill(s) installed") {
		t.Errorf("expected '2 skill(s) installed' in output, got: %s", output)
	}
}

func TestListSkillsTableEmpty(t *testing.T) {
	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = false

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsTable([]scribe.SkillInfo{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "No skills installed") {
		t.Errorf("expected 'No skills installed', got: %s", output)
	}
}

func TestListSkillsTableEmptyQuiet(t *testing.T) {
	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = true

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listSkillsTable([]scribe.SkillInfo{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(output, "No skills installed") {
		t.Error("quiet mode should suppress 'No skills installed' message")
	}
}

func TestListSkillsTableNoAgentsColumn(t *testing.T) {
	saveAndRestoreFlags(t)
	quiet = false

	skills := []scribe.SkillInfo{
		{
			Name:        "no-agents-skill",
			Description: "Skill with no agents",
			Source:      "owner/repo",
			SourceType:  "github",
			InstalledAt: "2024-06-15",
			Agents:      []string{},
		},
	}

	output := captureStdout(t, func() {
		_ = listSkillsTable(skills)
	})

	// Empty agents should display as "-"
	if !strings.Contains(output, "-") {
		t.Errorf("expected '-' for empty agents column, got: %s", output)
	}
}

func TestListSkillsTableQuiet(t *testing.T) {
	saveAndRestoreFlags(t)
	quiet = true

	skills := []scribe.SkillInfo{
		{
			Name:        "quiet-table-skill",
			Description: "Quiet table test",
			Source:      "owner/repo",
			SourceType:  "github",
			InstalledAt: "2024-06-15",
			Agents:      []string{"claude-code"},
		},
	}

	output := captureStdout(t, func() {
		_ = listSkillsTable(skills)
	})

	// Should still have table but not the summary count
	if strings.Contains(output, "skill(s) installed") {
		t.Error("quiet mode should suppress summary")
	}
	if !strings.Contains(output, "quiet-table-skill") {
		t.Error("should still show skill data in quiet mode")
	}
}

// ---------------------------------------------------------------------------
// Tests for runList (list.go)
// ---------------------------------------------------------------------------

// TestListEmptyOutput tests the list command with no skills installed
func TestListEmptyOutput(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "scribe-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override home directory for test
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	scribe.InitLoggerCLI(false)

	t.Run("empty list", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		jsonOutput = false
		namesOnly = false
		quiet = false
		_ = runList(listCmd, []string{})

		_ = w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "No skills installed") {
			t.Errorf("expected 'No skills installed' message, got: %s", output)
		}
	})
}

// TestQuietMode tests that quiet mode suppresses output
func TestQuietMode(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "scribe-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override home directory for test
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	scribe.InitLoggerCLI(false)

	t.Run("quiet list with no skills", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		quiet = true
		jsonOutput = false
		namesOnly = false
		_ = runList(listCmd, []string{})

		_ = w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// In quiet mode, "No skills installed" should not be printed
		if strings.Contains(output, "No skills installed") {
			t.Error("quiet mode should suppress 'No skills installed' message")
		}
	})
}

func TestRunListWithInstalledSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false
	namesOnly = false

	installFakeSkill(t, "list-skill-a", "First list skill", "github", "owner/repo-a")
	installFakeSkill(t, "list-skill-b", "Second list skill", "local", "/local/path")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "list-skill-a") {
		t.Errorf("expected 'list-skill-a' in output, got: %s", output)
	}
	if !strings.Contains(output, "list-skill-b") {
		t.Errorf("expected 'list-skill-b' in output, got: %s", output)
	}
	if !strings.Contains(output, "2 skill(s) installed") {
		t.Errorf("expected '2 skill(s) installed' in output, got: %s", output)
	}
}

func TestRunListJSON(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = true
	namesOnly = false

	installFakeSkill(t, "json-list-skill", "JSON list skill", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	var parsed struct {
		Skills []skillJSON `json:"skills"`
		Count  int         `json:"count"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Count != 1 {
		t.Errorf("expected count 1, got %d", parsed.Count)
	}
	if parsed.Skills[0].Name != "json-list-skill" {
		t.Errorf("expected 'json-list-skill', got %s", parsed.Skills[0].Name)
	}
}

func TestRunListNamesOnly(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = false
	namesOnly = true

	installFakeSkill(t, "name-only-a", "Name A", "github", "owner/repo")
	installFakeSkill(t, "name-only-b", "Name B", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(lines), lines)
	}
}

func TestRunListSearchArgumentFiltersInstalledSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false
	namesOnly = false

	installFakeSkill(t, "react-patterns", "React component patterns", "github", "owner/react")
	installFakeSkill(t, "typescript-tips", "TypeScript tips", "github", "owner/typescript")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{"rct"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "react-patterns") {
		t.Errorf("expected fuzzy match in output, got: %s", output)
	}
	if strings.Contains(output, "typescript-tips") {
		t.Errorf("did not expect non-matching skill in output, got: %s", output)
	}
	if !strings.Contains(output, "1 skill(s) installed") {
		t.Errorf("expected filtered count in output, got: %s", output)
	}
}

func TestRunListSearchFlagFiltersJSON(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = true
	namesOnly = false
	listSearch = "react source"

	installFakeSkill(t, "react-patterns", "React component patterns", "github", "owner/react-source")
	installFakeSkill(t, "go-idioms", "Go idioms", "github", "owner/go-source")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	var parsed struct {
		Skills []skillJSON `json:"skills"`
		Count  int         `json:"count"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Count != 1 {
		t.Fatalf("expected count 1, got %d", parsed.Count)
	}
	if parsed.Skills[0].Name != "react-patterns" {
		t.Errorf("expected react-patterns, got %s", parsed.Skills[0].Name)
	}
}

func TestRunListSearchNoMatches(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false
	namesOnly = false

	installFakeSkill(t, "react-patterns", "React component patterns", "github", "owner/react")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{"python"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `No skills match "python"`) {
		t.Errorf("expected no-match message, got: %s", output)
	}
}

func TestRunListSearchMatchesSkillMdContent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false
	namesOnly = false

	installFakeSkillWithBody(t, "content-search-skill", "Metadata without phrase", "github", "owner/content", "Use browser automation for visual checks.")
	installFakeSkillWithBody(t, "other-search-skill", "Other metadata", "github", "owner/other", "Use database migrations.")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{"browser automation"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "content-search-skill") {
		t.Errorf("expected SKILL.md content match in output, got: %s", output)
	}
	if strings.Contains(output, "other-search-skill") {
		t.Errorf("did not expect non-matching skill in output, got: %s", output)
	}
}

func TestRunListSearchRejectsDuplicateQuery(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	listSearch = "react"

	err := runList(listCmd, []string{"go"})
	if err == nil {
		t.Fatal("expected duplicate query error")
	}
	if !strings.Contains(err.Error(), "specified twice") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunListSortsAlphabetically(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	jsonOutput = false
	namesOnly = true

	// Install in reverse alphabetical order
	installFakeSkill(t, "zebra-skill", "Zebra", "github", "o/r")
	installFakeSkill(t, "alpha-skill", "Alpha", "github", "o/r")
	installFakeSkill(t, "middle-skill", "Middle", "github", "o/r")

	output := captureStdout(t, func() {
		err := runList(listCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "alpha-skill" {
		t.Errorf("expected first skill 'alpha-skill', got '%s'", lines[0])
	}
	if lines[1] != "middle-skill" {
		t.Errorf("expected second skill 'middle-skill', got '%s'", lines[1])
	}
	if lines[2] != "zebra-skill" {
		t.Errorf("expected third skill 'zebra-skill', got '%s'", lines[2])
	}
}
