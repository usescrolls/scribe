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
		{
			name: "bitbucket source",
			source: &scribe.SourceInfo{
				Type:  "bitbucket",
				Owner: "bbuser",
				Repo:  "bbrepo",
			},
			expected: "bitbucket:bbuser/bbrepo",
		},
		{
			name: "github SSH source",
			source: &scribe.SourceInfo{
				Type:  "github",
				Owner: "user",
				Repo:  "repo",
				URL:   "git@github.com:user/repo.git",
			},
			expected: "github(ssh):user/repo",
		},
		{
			name: "gitlab SSH source",
			source: &scribe.SourceInfo{
				Type:  "gitlab",
				Owner: "gluser",
				Repo:  "glrepo",
				URL:   "git@gitlab.com:gluser/glrepo.git",
			},
			expected: "gitlab(ssh):gluser/glrepo",
		},
		{
			name: "bitbucket SSH source",
			source: &scribe.SourceInfo{
				Type:  "bitbucket",
				Owner: "bbuser",
				Repo:  "bbrepo",
				URL:   "git@bitbucket.org:bbuser/bbrepo.git",
			},
			expected: "bitbucket(ssh):bbuser/bbrepo",
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

func TestRunInstallSyncsToDetectedAgents(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	installYes = true
	installListOnly = false
	installSkills = ""
	installAll = false

	// Create agent config dir so agent is "detected"
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Create a local skill source
	tmpSrc, err := os.MkdirTemp("", "scribe-install-agent-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skillContent := "---\nname: agent-sync-skill\ndescription: Skill synced to all agents\n---\n\n# Test\n"
	_ = os.WriteFile(filepath.Join(tmpSrc, "SKILL.md"), []byte(skillContent), 0o644)

	captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, _ := scribe.SkillExists("agent-sync-skill")
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
