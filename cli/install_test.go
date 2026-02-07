package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	scribe "github.com/usescrolls/scribe/internal"
)

// ---------------------------------------------------------------------------
// Tests for formatSourceInfo (install.go)
// ---------------------------------------------------------------------------

// TestFormatSourceInfo tests the formatSourceInfo function
func TestFormatSourceInfo(t *testing.T) {
	tests := []struct {
		name     string
		source   *scribe.SourceInfo
		expected string
	}{
		{
			name: "github source",
			source: &scribe.SourceInfo{
				Type:  "github",
				Owner: "user",
				Repo:  "repo",
			},
			expected: "github:user/repo",
		},
		{
			name: "github source with ref",
			source: &scribe.SourceInfo{
				Type:  "github",
				Owner: "user",
				Repo:  "repo",
				Ref:   "v1.0.0",
			},
			expected: "github:user/repo#v1.0.0",
		},
		{
			name: "local source",
			source: &scribe.SourceInfo{
				Type:      "local",
				LocalPath: "/path/to/skills",
			},
			expected: "local:/path/to/skills",
		},
		{
			name: "zip source",
			source: &scribe.SourceInfo{
				Type: "zip",
				URL:  "https://example.com/plugin.zip",
			},
			expected: "zip:https://example.com/plugin.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSourceInfo(tt.source)
			if result != tt.expected {
				t.Errorf("formatSourceInfo() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestFormatSourceInfoEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		source   *scribe.SourceInfo
		expected string
	}{
		{
			name: "github with subpath",
			source: &scribe.SourceInfo{
				Type:    "github",
				Owner:   "user",
				Repo:    "repo",
				Subpath: "skills/react",
			},
			expected: "github:user/repo/skills/react",
		},
		{
			name: "github with ref and subpath",
			source: &scribe.SourceInfo{
				Type:    "github",
				Owner:   "user",
				Repo:    "repo",
				Ref:     "v2.0",
				Subpath: "skills/react",
			},
			expected: "github:user/repo#v2.0/skills/react",
		},
		{
			name: "well-known source",
			source: &scribe.SourceInfo{
				Type: "well-known",
				URL:  "https://example.com/.well-known/skills",
			},
			expected: "https://example.com/.well-known/skills",
		},
		{
			name: "unknown type falls through to URL",
			source: &scribe.SourceInfo{
				Type: "other",
				URL:  "https://custom.example.com/skills",
			},
			expected: "https://custom.example.com/skills",
		},
		{
			name: "gitlab source",
			source: &scribe.SourceInfo{
				Type:  "gitlab",
				Owner: "gluser",
				Repo:  "glrepo",
			},
			expected: "gitlab:gluser/glrepo",
		},
		{
			name: "empty fields",
			source: &scribe.SourceInfo{
				Type:  "github",
				Owner: "",
				Repo:  "",
			},
			expected: "github:/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSourceInfo(tt.source)
			if result != tt.expected {
				t.Errorf("formatSourceInfo() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for parseSource (install.go)
// ---------------------------------------------------------------------------

// TestParseSource tests the parseSource function
func TestParseSource(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectType  string
		expectOwner string
		expectRepo  string
		expectRef   string
		expectError bool
	}{
		{
			name:        "github shorthand",
			input:       "owner/repo",
			expectType:  "github",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:        "github shorthand with ref",
			input:       "owner/repo#v1.0.0",
			expectType:  "github",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectRef:   "v1.0.0",
		},
		{
			name:        "github URL",
			input:       "https://github.com/owner/repo",
			expectType:  "github",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:       "local path relative",
			input:      "./my-skills",
			expectType: "local",
		},
		{
			name:       "local path absolute",
			input:      "/absolute/path",
			expectType: "local",
		},
		{
			name:       "zip URL",
			input:      "https://example.com/skills.zip",
			expectType: "zip",
		},
		{
			name:       "well-known URL",
			input:      "https://example.com",
			expectType: "well-known",
		},
		{
			name:        "invalid shorthand",
			input:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := parseSource(tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if source.Type != tt.expectType {
				t.Errorf("expected type %q, got %q", tt.expectType, source.Type)
			}
			if tt.expectOwner != "" && source.Owner != tt.expectOwner {
				t.Errorf("expected owner %q, got %q", tt.expectOwner, source.Owner)
			}
			if tt.expectRepo != "" && source.Repo != tt.expectRepo {
				t.Errorf("expected repo %q, got %q", tt.expectRepo, source.Repo)
			}
			if tt.expectRef != "" && source.Ref != tt.expectRef {
				t.Errorf("expected ref %q, got %q", tt.expectRef, source.Ref)
			}
		})
	}
}

func TestParseSourceTildePath(t *testing.T) {
	source, err := parseSource("~/my-skills")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.Type != "local" {
		t.Errorf("expected type 'local', got %s", source.Type)
	}
}

func TestParseSourceGitHubWithSubpath(t *testing.T) {
	source, err := parseSource("owner/repo/skills/react")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.Type != "github" {
		t.Errorf("expected type 'github', got %s", source.Type)
	}
	if source.Owner != "owner" {
		t.Errorf("expected owner 'owner', got %s", source.Owner)
	}
	if source.Repo != "repo" {
		t.Errorf("expected repo 'repo', got %s", source.Repo)
	}
	if source.Subpath != "skills/react" {
		t.Errorf("expected subpath 'skills/react', got %s", source.Subpath)
	}
}

func TestParseSourceGitHubShorthandWithRefAndSubpath(t *testing.T) {
	source, err := parseSource("owner/repo/skills/react#v2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.Ref != "v2.0" {
		t.Errorf("expected ref 'v2.0', got %s", source.Ref)
	}
	if source.Subpath != "skills/react" {
		t.Errorf("expected subpath 'skills/react', got %s", source.Subpath)
	}
}

func TestParseSourceGitLabURL(t *testing.T) {
	source, err := parseSource("https://gitlab.com/myowner/myrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.Type != "gitlab" {
		t.Errorf("expected type 'gitlab', got %s", source.Type)
	}
	if source.Owner != "myowner" {
		t.Errorf("expected owner 'myowner', got %s", source.Owner)
	}
}

func TestParseSourceHTTPURL(t *testing.T) {
	source, err := parseSource("http://example.com/skills")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.Type != "well-known" {
		t.Errorf("expected type 'well-known', got %s", source.Type)
	}
}

// ---------------------------------------------------------------------------
// Tests for parseGitLabURL (install.go)
// ---------------------------------------------------------------------------

func TestParseGitLabURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectOwner string
		expectRepo  string
		expectType  string
		expectURL   string
	}{
		{
			name:        "basic gitlab URL",
			url:         "https://gitlab.com/myowner/myrepo",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectType:  "gitlab",
			expectURL:   "https://gitlab.com/myowner/myrepo",
		},
		{
			name:        "gitlab URL with .git suffix",
			url:         "https://gitlab.com/myowner/myrepo.git",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectType:  "gitlab",
			expectURL:   "https://gitlab.com/myowner/myrepo.git",
		},
		{
			name:        "gitlab URL with subpath",
			url:         "https://gitlab.com/myowner/myrepo/sub/path",
			expectOwner: "myowner",
			expectRepo:  "myrepo",
			expectType:  "gitlab",
			expectURL:   "https://gitlab.com/myowner/myrepo/sub/path",
		},
		{
			name:       "gitlab URL with only owner",
			url:        "https://gitlab.com/onlyowner",
			expectType: "gitlab",
			expectURL:  "https://gitlab.com/onlyowner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := parseGitLabURL(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if source.Type != tt.expectType {
				t.Errorf("Type = %q, expected %q", source.Type, tt.expectType)
			}
			if source.URL != tt.expectURL {
				t.Errorf("URL = %q, expected %q", source.URL, tt.expectURL)
			}
			if tt.expectOwner != "" && source.Owner != tt.expectOwner {
				t.Errorf("Owner = %q, expected %q", source.Owner, tt.expectOwner)
			}
			if tt.expectRepo != "" && source.Repo != tt.expectRepo {
				t.Errorf("Repo = %q, expected %q", source.Repo, tt.expectRepo)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for parseGitHubURL edge cases (install.go)
// ---------------------------------------------------------------------------

func TestParseGitHubURLEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectOwner   string
		expectRepo    string
		expectRef     string
		expectSubpath string
	}{
		{
			name:        "basic github URL",
			url:         "https://github.com/owner/repo",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:        "github URL with .git suffix",
			url:         "https://github.com/owner/repo.git",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:          "github URL with tree/branch",
			url:           "https://github.com/owner/repo/tree/main",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectRef:     "main",
			expectSubpath: "",
		},
		{
			name:          "github URL with tree/branch/path",
			url:           "https://github.com/owner/repo/tree/develop/skills/react",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectRef:     "develop",
			expectSubpath: "skills/react",
		},
		{
			name:          "github URL with subpath no tree",
			url:           "https://github.com/owner/repo/skills/react",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectSubpath: "skills/react",
		},
		{
			name:          "github URL tree with nested path",
			url:           "https://github.com/owner/repo/tree/v1.0/deep/nested/path",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectRef:     "v1.0",
			expectSubpath: "deep/nested/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := parseGitHubURL(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if source.Type != "github" {
				t.Errorf("Type = %q, expected %q", source.Type, "github")
			}
			if source.Owner != tt.expectOwner {
				t.Errorf("Owner = %q, expected %q", source.Owner, tt.expectOwner)
			}
			if source.Repo != tt.expectRepo {
				t.Errorf("Repo = %q, expected %q", source.Repo, tt.expectRepo)
			}
			if source.Ref != tt.expectRef {
				t.Errorf("Ref = %q, expected %q", source.Ref, tt.expectRef)
			}
			if source.Subpath != tt.expectSubpath {
				t.Errorf("Subpath = %q, expected %q", source.Subpath, tt.expectSubpath)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for runInstall (install.go)
// ---------------------------------------------------------------------------

func TestRunInstallFromLocalDir(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installYes = true
	installAll = false
	installListOnly = false
	installAgents = ""
	installSkills = ""

	// Create a local skill source directory
	tmpSrc, err := os.MkdirTemp("", "scribe-install-src-*")
	if err != nil {
		t.Fatalf("failed to create temp source dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skillContent := "---\nname: local-install-test\ndescription: Test skill for local install\n---\n\n# Local Install Test\n\nHello.\n"
	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte(skillContent), 0o644)

	output := captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	// Install should fail because no agents detected in temp HOME
	if err != nil {
		if !strings.Contains(err.Error(), "no coding agents detected") {
			t.Errorf("unexpected error: %v", err)
		}
		return
	}

	// If we got here, install succeeded (some agent dirs exist)
	_ = output
	exists, _ := scribe.SkillExists("local-install-test")
	if !exists {
		t.Error("expected skill to be installed")
	}
}

func TestRunInstallListOnly(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	installYes = false
	installAll = false
	installListOnly = true
	installAgents = ""
	installSkills = ""

	// Create a local skill source directory
	tmpSrc, err := os.MkdirTemp("", "scribe-install-list-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skillContent := "---\nname: list-only-test\ndescription: List only test skill\n---\n\n# Test\n"
	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte(skillContent), 0o644)

	output := captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Found 1 skill(s)") {
		t.Errorf("expected 'Found 1 skill(s)' in output, got: %s", output)
	}
	if !strings.Contains(output, "list-only-test") {
		t.Errorf("expected skill name in output, got: %s", output)
	}

	// Skill should NOT be installed in list-only mode
	exists, _ := scribe.SkillExists("list-only-test")
	if exists {
		t.Error("skill should not be installed in list-only mode")
	}
}

func TestRunInstallInvalidSource(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installListOnly = false

	err := runInstall(installCmd, []string{"invalid"})
	if err == nil {
		t.Error("expected error for invalid source")
	}
	if !strings.Contains(err.Error(), "invalid source") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunInstallNoSkillsFound(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installListOnly = false

	// Create an empty local source directory
	tmpSrc, err := os.MkdirTemp("", "scribe-install-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	err = runInstall(installCmd, []string{tmpSrc})
	if err == nil {
		t.Error("expected error when no skills found")
	}
	if !strings.Contains(err.Error(), "no skills found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunInstallFilterSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installYes = true
	installListOnly = false
	installSkills = "nonexistent-skill"
	installAgents = ""
	installAll = false

	// Create a local skill source directory with a skill
	tmpSrc, err := os.MkdirTemp("", "scribe-install-filter-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skillContent := "---\nname: real-skill\ndescription: A real skill\n---\n\n# Real\n"
	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte(skillContent), 0o644)

	err = runInstall(installCmd, []string{tmpSrc})
	if err == nil {
		t.Error("expected error when filtering for nonexistent skill")
	}
	if !strings.Contains(err.Error(), "no matching skills") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunInstallWithAgentFilter(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installYes = true
	installListOnly = false
	installSkills = ""
	installAgents = "claude-code"
	installAll = false

	// Create agent config dir so agent is "detected"
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Create a local skill source
	tmpSrc, err := os.MkdirTemp("", "scribe-install-agent-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skillContent := "---\nname: agent-filter-skill\ndescription: Skill with agent filter\n---\n\n# Test\n"
	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte(skillContent), 0o644)

	captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, _ := scribe.SkillExists("agent-filter-skill")
	if !exists {
		t.Error("expected skill to be installed")
	}
}

func TestRunInstallLocalWithMultipleSkills(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	installYes = true
	installListOnly = false
	installSkills = ""
	installAgents = ""
	installAll = false

	// Create agent config dir
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Create source with multiple skills
	tmpSrc, err := os.MkdirTemp("", "scribe-multi-install-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	// Create two skills in subdirectories
	skill1Dir := filepath.Join(tmpSrc, "skills", "skill-one")
	skill2Dir := filepath.Join(tmpSrc, "skills", "skill-two")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: skill-one\ndescription: First skill\n---\n\nContent 1\n"), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("---\nname: skill-two\ndescription: Second skill\n---\n\nContent 2\n"), 0o644)

	output := captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "skill-one") {
		t.Errorf("expected 'skill-one' in output, got: %s", output)
	}
	if !strings.Contains(output, "skill-two") {
		t.Errorf("expected 'skill-two' in output, got: %s", output)
	}
}

func TestRunInstallFilterBySpecificSkill(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installYes = true
	installListOnly = false
	installSkills = "pick-me"
	installAgents = ""
	installAll = false

	// Create agent config dir
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Create source with two skills
	tmpSrc, err := os.MkdirTemp("", "scribe-filter-install-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skill1Dir := filepath.Join(tmpSrc, "skills", "pick-me")
	skill2Dir := filepath.Join(tmpSrc, "skills", "skip-me")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: pick-me\ndescription: Pick this one\n---\n\nPicked\n"), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("---\nname: skip-me\ndescription: Skip this one\n---\n\nSkipped\n"), 0o644)

	captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, _ := scribe.SkillExists("pick-me")
	if !exists {
		t.Error("expected 'pick-me' to be installed")
	}
	exists, _ = scribe.SkillExists("skip-me")
	if exists {
		t.Error("expected 'skip-me' to NOT be installed")
	}
}

func TestRunInstallDuplicateSkill(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	installYes = true
	installListOnly = false
	installSkills = ""
	installAgents = ""
	installAll = false

	// Create agent config dir
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Pre-install the skill
	installFakeSkill(t, "already-here", "Already installed", "github", "owner/repo")

	// Create source with same skill name
	tmpSrc, err := os.MkdirTemp("", "scribe-dup-install-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte("---\nname: already-here\ndescription: Duplicate\n---\n\nDup\n"), 0o644)

	// Capture stderr too since "Failed to install" goes to stderr
	oldStderr := os.Stderr
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	output := captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	_ = stderrW.Close()
	os.Stderr = oldStderr
	var stderrBuf bytes.Buffer
	_, _ = stderrBuf.ReadFrom(stderrR)
	stderrOutput := stderrBuf.String()

	// Should not fail overall, but the skill should fail to install
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "Failed to install" goes to stderr, "Installed 0/1" goes to stdout
	if !strings.Contains(output, "0/1") {
		t.Errorf("expected '0/1' installed count, got stdout: %s", output)
	}
	if !strings.Contains(stderrOutput, "already exists") {
		t.Errorf("expected 'already exists' in stderr, got: %s", stderrOutput)
	}
}

func TestRunInstallVerboseOutput(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	installYes = true
	installListOnly = false
	installSkills = ""
	installAgents = ""
	installAll = false

	// Create agent config dir
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Create a local skill source
	tmpSrc, err := os.MkdirTemp("", "scribe-verbose-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte("---\nname: verbose-skill\ndescription: Verbose test skill\n---\n\n# Verbose\n"), 0o644)

	output := captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verbose output should contain "Fetching skills from"
	if !strings.Contains(output, "Fetching skills from") {
		t.Errorf("expected 'Fetching skills from', got: %s", output)
	}
	// Should list found skills
	if !strings.Contains(output, "Found 1 skill(s) to install") {
		t.Errorf("expected 'Found 1 skill(s) to install', got: %s", output)
	}
	// Should list detected agents
	if !strings.Contains(output, "Detected") {
		t.Errorf("expected 'Detected' in output, got: %s", output)
	}
	// Should show install progress
	if !strings.Contains(output, "Installing verbose-skill") {
		t.Errorf("expected 'Installing verbose-skill', got: %s", output)
	}
	// Should show success count
	if !strings.Contains(output, "Installed 1/1 skill(s)") {
		t.Errorf("expected 'Installed 1/1 skill(s)', got: %s", output)
	}
}
