package scribe

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// expandPath (agents.go)
// ============================================================================

func TestExpandPath_Tilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}
	result := expandPath("~/foo")
	expected := filepath.Join(home, "foo")
	if result != expected {
		t.Errorf("expandPath('~/foo') = %q, want %q", result, expected)
	}
}

func TestExpandPath_AbsolutePath(t *testing.T) {
	result := expandPath("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("expandPath('/absolute/path') = %q, want '/absolute/path'", result)
	}
}

func TestExpandPath_RelativePath(t *testing.T) {
	result := expandPath("relative/path")
	if result != "relative/path" {
		t.Errorf("expandPath('relative/path') = %q, want 'relative/path'", result)
	}
}

func TestExpandPath_TildeOnly(t *testing.T) {
	// "~" without "/" prefix should not be expanded
	result := expandPath("~")
	if result != "~" {
		t.Errorf("expandPath('~') = %q, want '~' (no expansion for bare tilde)", result)
	}
}

func TestExpandPath_EmptyString(t *testing.T) {
	result := expandPath("")
	if result != "" {
		t.Errorf("expandPath('') = %q, want ''", result)
	}
}

// ============================================================================
// dirExists (agents.go)
// ============================================================================

func TestDirExists_ExistingDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if !dirExists(tmpDir) {
		t.Errorf("dirExists(%q) = false, want true", tmpDir)
	}
}

func TestDirExists_NonExisting(t *testing.T) {
	if dirExists("/tmp/nonexistent-scribe-test-dir-999999") {
		t.Error("dirExists(non-existent) = true, want false")
	}
}

func TestDirExists_File(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "scribe-test-file-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if dirExists(tmpFile.Name()) {
		t.Errorf("dirExists(file) = true, want false (not a directory)")
	}
}

// ============================================================================
// countSkillsInDir (agents.go)
// ============================================================================

func TestCountSkillsInDir_WithSkills(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create 3 skill subdirectories with SKILL.md
	for _, name := range []string{"skill-a", "skill-b", "skill-c"} {
		skillDir := filepath.Join(tmpDir, name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# "+name), 0o644)
	}

	// Create a directory without SKILL.md (should not count)
	_ = os.MkdirAll(filepath.Join(tmpDir, "not-a-skill"), 0o755)

	// Create a file (should not count)
	_ = os.WriteFile(filepath.Join(tmpDir, "random.txt"), []byte("nope"), 0o644)

	count := countSkillsInDir(tmpDir)
	if count != 3 {
		t.Errorf("countSkillsInDir() = %d, want 3", count)
	}
}

func TestCountSkillsInDir_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	count := countSkillsInDir(tmpDir)
	if count != 0 {
		t.Errorf("countSkillsInDir(empty) = %d, want 0", count)
	}
}

func TestCountSkillsInDir_NonExistent(t *testing.T) {
	count := countSkillsInDir("/tmp/nonexistent-scribe-test-dir-888888")
	if count != 0 {
		t.Errorf("countSkillsInDir(non-existent) = %d, want 0", count)
	}
}

// ============================================================================
// GetAgent / GetAllAgents (agents.go)
// ============================================================================

func TestGetAgent_Known(t *testing.T) {
	agent := GetAgent("claude-code")
	if agent == nil {
		t.Fatal("GetAgent('claude-code') = nil, want non-nil")
	}
	if agent.DisplayName != "Claude Code" {
		t.Errorf("DisplayName = %q, want 'Claude Code'", agent.DisplayName)
	}
	if agent.GlobalSkillsDir == "" {
		t.Error("GlobalSkillsDir is empty")
	}
}

func TestGetAgent_Unknown(t *testing.T) {
	agent := GetAgent("nonexistent-agent-xyz")
	if agent != nil {
		t.Errorf("GetAgent('nonexistent') = %v, want nil", agent)
	}
}

func TestGetAgent_EmptyString(t *testing.T) {
	agent := GetAgent("")
	if agent != nil {
		t.Errorf("GetAgent('') = %v, want nil", agent)
	}
}

func TestGetAllAgents_ReturnsCopy(t *testing.T) {
	agents1 := GetAllAgents()
	agents2 := GetAllAgents()

	if len(agents1) != len(agents2) {
		t.Fatalf("GetAllAgents() returned different lengths: %d vs %d", len(agents1), len(agents2))
	}

	// Mutating the returned slice should not affect the global
	if len(agents1) > 0 {
		agents1[0].DisplayName = "MUTATED"
		original := GetAgent(agents1[0].ID)
		if original != nil && original.DisplayName == "MUTATED" {
			t.Error("GetAllAgents() did not return a copy; mutation leaked to global")
		}
	}
}

func TestGetAllAgents_NonEmpty(t *testing.T) {
	agents := GetAllAgents()
	if len(agents) == 0 {
		t.Error("GetAllAgents() returned empty slice")
	}
}

// ============================================================================
// GetAgentStatus (agents.go)
// ============================================================================

func TestGetAgentStatus_EmptyHome(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	InitLoggerCLI(false)

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	statuses := GetAgentStatus(scrollsDir)

	if len(statuses) != len(AllAgents) {
		t.Errorf("GetAgentStatus() returned %d statuses, want %d", len(statuses), len(AllAgents))
	}

	// With empty HOME, no agents should be installed
	for _, s := range statuses {
		if s.Installed {
			t.Errorf("agent %q reported as installed in empty HOME", s.ID)
		}
		if s.SkillCount != 0 {
			t.Errorf("agent %q has %d skills in empty HOME, want 0", s.ID, s.SkillCount)
		}
	}
}

func TestGetAgentStatus_WithAgent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	InitLoggerCLI(false)

	// Create claude-code config and skills directories
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)
	skillsDir := filepath.Join(tmpDir, ".claude", "skills")
	_ = os.MkdirAll(filepath.Join(skillsDir, "test-skill"), 0o755)
	_ = os.WriteFile(filepath.Join(skillsDir, "test-skill", "SKILL.md"), []byte("# Test"), 0o644)

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	statuses := GetAgentStatus(scrollsDir)

	// Find claude-code status
	var claudeStatus *AgentStatus
	for i := range statuses {
		if statuses[i].ID == "claude-code" {
			claudeStatus = &statuses[i]
			break
		}
	}

	if claudeStatus == nil {
		t.Fatal("claude-code not found in statuses")
	}
	if !claudeStatus.Installed {
		t.Error("claude-code should be installed")
	}
	if claudeStatus.SkillCount != 1 {
		t.Errorf("claude-code SkillCount = %d, want 1", claudeStatus.SkillCount)
	}
}

// ============================================================================
// ExpandAgentPath (agents.go)
// ============================================================================

func TestExpandAgentPath_Tilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}
	result := ExpandAgentPath("~/test/path")
	expected := filepath.Join(home, "test", "path")
	if result != expected {
		t.Errorf("ExpandAgentPath('~/test/path') = %q, want %q", result, expected)
	}
}

func TestExpandAgentPath_Absolute(t *testing.T) {
	result := ExpandAgentPath("/usr/local/bin")
	if result != "/usr/local/bin" {
		t.Errorf("ExpandAgentPath('/usr/local/bin') = %q, want '/usr/local/bin'", result)
	}
}

func TestExpandAgentPath_Relative(t *testing.T) {
	result := ExpandAgentPath("some/relative")
	if result != "some/relative" {
		t.Errorf("ExpandAgentPath('some/relative') = %q, want 'some/relative'", result)
	}
}
