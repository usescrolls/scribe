package scribe

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// Storage Tests
// ============================================================================

func TestGetScrollsDir(t *testing.T) {
	dir, err := GetScrollsDir()
	if err != nil {
		t.Fatalf("GetScrollsDir() error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".scribe", "scrolls")
	if dir != expected {
		t.Errorf("GetScrollsDir() = %q, want %q", dir, expected)
	}
}

func TestGetSkillDir(t *testing.T) {
	dir, err := GetSkillDir("test-skill")
	if err != nil {
		t.Fatalf("GetSkillDir() error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".scribe", "scrolls", "test-skill")
	if dir != expected {
		t.Errorf("GetSkillDir() = %q, want %q", dir, expected)
	}
}

func TestGetSkillPath(t *testing.T) {
	path, err := GetSkillPath("test-skill")
	if err != nil {
		t.Fatalf("GetSkillPath() error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".scribe", "scrolls", "test-skill", "SKILL.md")
	if path != expected {
		t.Errorf("GetSkillPath() = %q, want %q", path, expected)
	}
}

func TestGetMetaPath(t *testing.T) {
	path, err := GetMetaPath("test-skill")
	if err != nil {
		t.Fatalf("GetMetaPath() error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".scribe", "scrolls", "test-skill", ".scribe-meta.json")
	if path != expected {
		t.Errorf("GetMetaPath() = %q, want %q", path, expected)
	}
}

func TestListInstalledSkills_Empty(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override home directory for test
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	skills, err := ListInstalledSkills()
	if err != nil {
		t.Fatalf("ListInstalledSkills() error: %v", err)
	}

	if len(skills) != 0 {
		t.Errorf("ListInstalledSkills() returned %d skills, want 0", len(skills))
	}
}

func TestListInstalledSkills_WithSkills(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Create skill directories
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skill1Dir := filepath.Join(scrollsDir, "skill-one")
	skill2Dir := filepath.Join(scrollsDir, "skill-two")

	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)

	// Create SKILL.md files
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Skill One"), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Skill Two"), 0o644)

	skills, err := ListInstalledSkills()
	if err != nil {
		t.Fatalf("ListInstalledSkills() error: %v", err)
	}

	if len(skills) != 2 {
		t.Errorf("ListInstalledSkills() returned %d skills, want 2", len(skills))
	}
}

func TestSkillExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Test non-existent skill
	exists, err := SkillExists("non-existent")
	if err != nil {
		t.Fatalf("SkillExists() error: %v", err)
	}
	if exists {
		t.Error("SkillExists() = true for non-existent skill")
	}

	// Create a skill
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls", "existing-skill")
	_ = os.MkdirAll(scrollsDir, 0o755)
	_ = os.WriteFile(filepath.Join(scrollsDir, "SKILL.md"), []byte("# Existing"), 0o644)

	// Test existing skill
	exists, err = SkillExists("existing-skill")
	if err != nil {
		t.Fatalf("SkillExists() error: %v", err)
	}
	if !exists {
		t.Error("SkillExists() = false for existing skill")
	}
}

// ============================================================================
// Agent Tests
// ============================================================================

func TestGetAgent(t *testing.T) {
	tests := []struct {
		id       string
		wantNil  bool
		wantName string
	}{
		{"claude-code", false, "Claude Code"},
		{"cursor", false, "Cursor"},
		{"non-existent", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			agent := GetAgent(tt.id)
			if tt.wantNil {
				if agent != nil {
					t.Errorf("GetAgent(%q) = %v, want nil", tt.id, agent)
				}
			} else {
				if agent == nil {
					t.Fatalf("GetAgent(%q) = nil, want non-nil", tt.id)
				}
				if agent.DisplayName != tt.wantName {
					t.Errorf("GetAgent(%q).DisplayName = %q, want %q", tt.id, agent.DisplayName, tt.wantName)
				}
			}
		})
	}
}

func TestGetAllAgents(t *testing.T) {
	agents := GetAllAgents()

	if len(agents) == 0 {
		t.Error("GetAllAgents() returned empty slice")
	}

	// Should have 39 agents (synced with Vercel skills CLI)
	if len(agents) != 39 {
		t.Errorf("GetAllAgents() returned %d agents, want 39", len(agents))
	}

	// Verify first agent is claude-code
	if agents[0].ID != "claude-code" {
		t.Errorf("First agent ID = %q, want 'claude-code'", agents[0].ID)
	}
}

func TestAgentHasNoProjectSkillsDir(t *testing.T) {
	// Verify agents don't have ProjectSkillsDir field (global-only)
	agent := GetAgent("claude-code")
	if agent == nil {
		t.Fatal("GetAgent('claude-code') = nil")
	}

	// The struct should only have ID, DisplayName, GlobalSkillsDir, GlobalConfigDir
	if agent.GlobalSkillsDir == "" {
		t.Error("Agent.GlobalSkillsDir is empty")
	}
	if agent.GlobalConfigDir == "" {
		t.Error("Agent.GlobalConfigDir is empty")
	}
}

func TestDetectInstalledAgents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// No agents installed initially
	agents := DetectInstalledAgents()
	if len(agents) != 0 {
		t.Errorf("DetectInstalledAgents() = %d agents, want 0", len(agents))
	}

	// Create Claude config directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Now should detect Claude
	agents = DetectInstalledAgents()
	if len(agents) != 1 {
		t.Errorf("DetectInstalledAgents() = %d agents, want 1", len(agents))
	}
	if len(agents) > 0 && agents[0].ID != "claude-code" {
		t.Errorf("Detected agent ID = %q, want 'claude-code'", agents[0].ID)
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/.claude/skills", filepath.Join(home, ".claude", "skills")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Skills Parsing Tests
// ============================================================================

func TestParseSkillMd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a valid SKILL.md
	skillDir := filepath.Join(tmpDir, "test-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	skillContent := `---
name: test-skill
description: A test skill for testing
---

# Test Skill

This is a test skill.
`
	skillPath := filepath.Join(skillDir, "SKILL.md")
	_ = os.WriteFile(skillPath, []byte(skillContent), 0o644)

	skill, err := ParseSkillMd(skillPath)
	if err != nil {
		t.Fatalf("ParseSkillMd() error: %v", err)
	}

	if skill.Name != "test-skill" {
		t.Errorf("skill.Name = %q, want 'test-skill'", skill.Name)
	}
	if skill.Description != "A test skill for testing" {
		t.Errorf("skill.Description = %q, want 'A test skill for testing'", skill.Description)
	}
}

func TestParseSkillMd_MissingFrontmatter(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	skillDir := filepath.Join(tmpDir, "no-frontmatter")
	_ = os.MkdirAll(skillDir, 0o755)
	skillContent := `# No Frontmatter

This skill has no YAML frontmatter.
`
	skillPath := filepath.Join(skillDir, "SKILL.md")
	_ = os.WriteFile(skillPath, []byte(skillContent), 0o644)

	_, err = ParseSkillMd(skillPath)
	if err == nil {
		t.Error("ParseSkillMd() should error on missing frontmatter")
	}
}

func TestDiscoverSkills(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create skill directories with SKILL.md files
	skill1Dir := filepath.Join(tmpDir, "skills", "skill-one")
	skill2Dir := filepath.Join(tmpDir, "skills", "skill-two")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)

	skill1Content := `---
name: skill-one
description: First skill
---
# Skill One
`
	skill2Content := `---
name: skill-two
description: Second skill
---
# Skill Two
`
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skill1Content), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skill2Content), 0o644)

	skills, err := DiscoverSkills(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverSkills() error: %v", err)
	}

	if len(skills) != 2 {
		t.Errorf("DiscoverSkills() found %d skills, want 2", len(skills))
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Name", "simple-name"},
		{"UPPERCASE", "uppercase"},
		{"with spaces", "with-spaces"},
		{"with_underscores", "with-underscores"},
		{"path/traversal", "pathtraversal"},
		{"../dangerous", "dangerous"},
		{"./relative", "relative"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Meta Tests
// ============================================================================

func TestComputeContentHash(t *testing.T) {
	hash1 := ComputeContentHash("test content")
	hash2 := ComputeContentHash("test content")
	hash3 := ComputeContentHash("different content")

	if hash1 != hash2 {
		t.Error("Same content should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("Different content should produce different hash")
	}
	if len(hash1) != 71 { // "sha256:" + 64 hex chars
		t.Errorf("Hash length = %d, want 71", len(hash1))
	}
}

func TestSkillMetaReadWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	metaPath := filepath.Join(tmpDir, ".scribe-meta.json")

	// Create meta
	source := &SourceInfo{
		Type:  "github",
		Owner: "user",
		Repo:  "repo",
	}
	meta := NewSkillMeta(source, "skills/test", "test content", nil)

	// Write
	err = WriteSkillMeta(metaPath, meta)
	if err != nil {
		t.Fatalf("WriteSkillMeta() error: %v", err)
	}

	// Read
	readMeta, err := ReadSkillMeta(metaPath)
	if err != nil {
		t.Fatalf("ReadSkillMeta() error: %v", err)
	}

	if readMeta.Source != meta.Source {
		t.Errorf("Source = %q, want %q", readMeta.Source, meta.Source)
	}
	if readMeta.SourceType != meta.SourceType {
		t.Errorf("SourceType = %q, want %q", readMeta.SourceType, meta.SourceType)
	}
	if readMeta.ContentHash != meta.ContentHash {
		t.Errorf("ContentHash = %q, want %q", readMeta.ContentHash, meta.ContentHash)
	}
}

// ============================================================================
// Workspace Tests
// ============================================================================

func TestWorkspaceCRUD(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Ensure directories exist
	_ = EnsureScribeDirs()

	// Create workspace
	ws := &Workspace{
		Name:        "test-workspace",
		Description: "A test workspace",
		Skills:      []string{"skill-1", "skill-2"},
	}

	err = CreateWorkspace(ws)
	if err != nil {
		t.Fatalf("CreateWorkspace() error: %v", err)
	}

	// Read workspace
	readWs, err := GetWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("GetWorkspace() error: %v", err)
	}

	if readWs.Name != ws.Name {
		t.Errorf("Name = %q, want %q", readWs.Name, ws.Name)
	}
	if readWs.Description != ws.Description {
		t.Errorf("Description = %q, want %q", readWs.Description, ws.Description)
	}
	if len(readWs.Skills) != 2 {
		t.Errorf("Skills count = %d, want 2", len(readWs.Skills))
	}

	// Update workspace
	ws.Skills = append(ws.Skills, "skill-3")
	err = UpdateWorkspace(ws)
	if err != nil {
		t.Fatalf("UpdateWorkspace() error: %v", err)
	}

	readWs, _ = GetWorkspace("test-workspace")
	if len(readWs.Skills) != 3 {
		t.Errorf("Skills count after update = %d, want 3", len(readWs.Skills))
	}

	// Delete workspace
	err = DeleteWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("DeleteWorkspace() error: %v", err)
	}

	_, err = GetWorkspace("test-workspace")
	if err == nil {
		t.Error("GetWorkspace() should error after delete")
	}
}

func TestDeleteDefaultWorkspaceFails(t *testing.T) {
	err := DeleteWorkspace("default")
	if err == nil {
		t.Error("DeleteWorkspace('default') should fail")
	}
}

func TestAddSkillToWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create workspace
	ws := &Workspace{
		Name:   "test-ws",
		Skills: []string{},
	}
	_ = CreateWorkspace(ws)

	// Add skill
	err = AddSkillToWorkspace("new-skill", "test-ws")
	if err != nil {
		t.Fatalf("AddSkillToWorkspace() error: %v", err)
	}

	// Verify
	readWs, _ := GetWorkspace("test-ws")
	if len(readWs.Skills) != 1 {
		t.Errorf("Skills count = %d, want 1", len(readWs.Skills))
	}
	if readWs.Skills[0] != "new-skill" {
		t.Errorf("Skill = %q, want 'new-skill'", readWs.Skills[0])
	}

	// Adding same skill again should be idempotent
	err = AddSkillToWorkspace("new-skill", "test-ws")
	if err != nil {
		t.Fatalf("AddSkillToWorkspace() second call error: %v", err)
	}
	readWs, _ = GetWorkspace("test-ws")
	if len(readWs.Skills) != 1 {
		t.Errorf("Skills count after duplicate add = %d, want 1", len(readWs.Skills))
	}
}

func TestRemoveSkillFromWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create workspace with skills
	ws := &Workspace{
		Name:   "test-ws",
		Skills: []string{"skill-1", "skill-2", "skill-3"},
	}
	_ = CreateWorkspace(ws)

	// Remove skill
	err = RemoveSkillFromWorkspace("skill-2", "test-ws")
	if err != nil {
		t.Fatalf("RemoveSkillFromWorkspace() error: %v", err)
	}

	// Verify
	readWs, _ := GetWorkspace("test-ws")
	if len(readWs.Skills) != 2 {
		t.Errorf("Skills count = %d, want 2", len(readWs.Skills))
	}
	for _, s := range readWs.Skills {
		if s == "skill-2" {
			t.Error("skill-2 should have been removed")
		}
	}
}

// ============================================================================
// Installer Tests
// ============================================================================

func TestInstallAndUninstallSkill(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create a source skill directory
	sourceDir := filepath.Join(tmpDir, "source", "test-skill")
	_ = os.MkdirAll(sourceDir, 0o755)
	skillContent := `---
name: test-skill
description: Test skill for installation
---
# Test Skill
`
	_ = os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(skillContent), 0o644)

	// Parse the skill
	skill, err := ParseSkillMd(filepath.Join(sourceDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("ParseSkillMd() error: %v", err)
	}

	source := &SourceInfo{
		Type:      "local",
		LocalPath: sourceDir,
	}

	opts := InstallOptions{
		Agents: []string{}, // Empty means no agents to sync to
	}

	// Install
	err = InstallSkill(skill, source, opts, nil)
	if err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	// Verify installed
	exists, _ := SkillExists("test-skill")
	if !exists {
		t.Error("Skill should exist after install")
	}

	// Verify meta was written
	metaPath, _ := GetMetaPath("test-skill")
	if _, err := os.Stat(metaPath); err != nil {
		t.Error("Meta file should exist after install")
	}

	// Uninstall
	err = UninstallSkill("test-skill")
	if err != nil {
		t.Fatalf("UninstallSkill() error: %v", err)
	}

	// Verify uninstalled
	exists, _ = SkillExists("test-skill")
	if exists {
		t.Error("Skill should not exist after uninstall")
	}
}

func TestSyncSkillToAgents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create a skill in scrolls
	skillDir, _ := GetSkillDir("sync-test")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Test"), 0o644)

	// Create Claude agent directory
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	_ = os.MkdirAll(claudeSkillsDir, 0o755)

	// Sync to Claude
	err = SyncSkillToAgents("sync-test", []string{"claude-code"})
	if err != nil {
		t.Fatalf("SyncSkillToAgents() error: %v", err)
	}

	// Verify symlink was created
	linkPath := filepath.Join(claudeSkillsDir, "sync-test")
	if _, err := os.Lstat(linkPath); err != nil {
		t.Error("Symlink should exist after sync")
	}
}

func TestRemoveSkillFromAgents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Create a skill symlink in Claude directory
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	_ = os.MkdirAll(claudeSkillsDir, 0o755)
	linkPath := filepath.Join(claudeSkillsDir, "remove-test")
	_ = os.MkdirAll(linkPath, 0o755) // Create as directory for test

	// Remove
	err = RemoveSkillFromAgents("remove-test", []string{"claude-code"})
	if err != nil {
		t.Fatalf("RemoveSkillFromAgents() error: %v", err)
	}

	// Verify removed
	if _, err := os.Stat(linkPath); !os.IsNotExist(err) {
		t.Error("Skill directory should be removed")
	}
}

// ============================================================================
// Config Tests
// ============================================================================

func TestLoadConfigDefault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Load config when file doesn't exist
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if config.ActiveWorkspace != "default" {
		t.Errorf("ActiveWorkspace = %q, want 'default'", config.ActiveWorkspace)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	config := &Config{
		ActiveWorkspace: "custom-workspace",
	}

	err = SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if loaded.ActiveWorkspace != config.ActiveWorkspace {
		t.Errorf("ActiveWorkspace = %q, want %q", loaded.ActiveWorkspace, config.ActiveWorkspace)
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestFullInstallWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create Claude agent directory (simulating Claude Code is installed)
	claudeDir := filepath.Join(tmpDir, ".claude")
	claudeSkillsDir := filepath.Join(claudeDir, "skills")
	_ = os.MkdirAll(claudeSkillsDir, 0o755)

	// Create source skill
	sourceDir := filepath.Join(tmpDir, "source", "my-skill")
	_ = os.MkdirAll(sourceDir, 0o755)
	skillContent := `---
name: my-skill
description: My awesome skill
---
# My Skill

This skill does cool things.
`
	_ = os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(skillContent), 0o644)

	// Parse and install
	skill, _ := ParseSkillMd(filepath.Join(sourceDir, "SKILL.md"))
	source := &SourceInfo{Type: "local", LocalPath: sourceDir}

	// Install to all detected agents
	detected := DetectInstalledAgents()
	agentIDs := make([]string, len(detected))
	for i, a := range detected {
		agentIDs[i] = a.ID
	}

	err = InstallSkill(skill, source, InstallOptions{Agents: agentIDs}, nil)
	if err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	// Add to workspace
	_ = EnsureDefaultWorkspace()
	_ = AddSkillToActiveAndDefaultWorkspace("my-skill")

	// Verify installation
	exists, _ := SkillExists("my-skill")
	if !exists {
		t.Error("Skill should exist")
	}

	// Verify symlink in Claude directory
	linkPath := filepath.Join(claudeSkillsDir, "my-skill")
	if _, err := os.Lstat(linkPath); err != nil {
		t.Error("Symlink should exist in Claude skills directory")
	}

	// Verify in default workspace
	ws, _ := GetWorkspace("default")
	found := false
	for _, s := range ws.Skills {
		if s == "my-skill" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Skill should be in default workspace")
	}

	// Uninstall
	_ = RemoveSkillFromAllWorkspaces("my-skill")
	_ = UninstallSkill("my-skill")

	// Verify uninstallation
	exists, _ = SkillExists("my-skill")
	if exists {
		t.Error("Skill should not exist after uninstall")
	}

	if _, err := os.Stat(linkPath); !os.IsNotExist(err) {
		t.Error("Symlink should be removed after uninstall")
	}
}

// ============================================================================
// Additional Coverage Tests
// ============================================================================

func TestGetAgentStatus(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")

	// No agents installed
	statuses := GetAgentStatus(scrollsDir)
	if len(statuses) != 39 {
		t.Errorf("GetAgentStatus() returned %d statuses, want 39", len(statuses))
	}

	// All should be not installed
	for _, s := range statuses {
		if s.Installed {
			t.Errorf("Agent %s should not be installed", s.ID)
		}
	}

	// Create Claude directory with some skills
	claudeDir := filepath.Join(tmpDir, ".claude")
	claudeSkillsDir := filepath.Join(claudeDir, "skills")
	_ = os.MkdirAll(claudeSkillsDir, 0o755)

	// Create a skill directory
	skill1Dir := filepath.Join(claudeSkillsDir, "test-skill")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Test"), 0o644)

	statuses = GetAgentStatus(scrollsDir)

	// Find Claude status
	var claudeStatus *AgentStatus
	for i := range statuses {
		if statuses[i].ID == "claude-code" {
			claudeStatus = &statuses[i]
			break
		}
	}

	if claudeStatus == nil {
		t.Fatal("Claude status not found")
	}
	if !claudeStatus.Installed {
		t.Error("Claude should be installed")
	}
	if claudeStatus.SkillCount != 1 {
		t.Errorf("Claude skill count = %d, want 1", claudeStatus.SkillCount)
	}
}

func TestExpandAgentPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	result := ExpandAgentPath("~/.claude/skills")
	expected := filepath.Join(home, ".claude", "skills")

	if result != expected {
		t.Errorf("ExpandAgentPath() = %q, want %q", result, expected)
	}
}

func TestCountSkillsInDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Empty directory
	count := countSkillsInDir(tmpDir)
	if count != 0 {
		t.Errorf("countSkillsInDir() = %d, want 0", count)
	}

	// Non-existent directory
	count = countSkillsInDir(filepath.Join(tmpDir, "non-existent"))
	if count != 0 {
		t.Errorf("countSkillsInDir() for non-existent = %d, want 0", count)
	}

	// Create some skill directories
	skill1Dir := filepath.Join(tmpDir, "skill1")
	skill2Dir := filepath.Join(tmpDir, "skill2")
	noSkillDir := filepath.Join(tmpDir, "no-skill")

	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)
	_ = os.MkdirAll(noSkillDir, 0o755)

	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Skill 1"), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Skill 2"), 0o644)
	_ = os.WriteFile(filepath.Join(noSkillDir, "README.md"), []byte("# Not a skill"), 0o644)

	count = countSkillsInDir(tmpDir)
	if count != 2 {
		t.Errorf("countSkillsInDir() = %d, want 2", count)
	}

	// Symlinked skill directories (the real-world case: agent dirs contain symlinks to ~/.scribe/scrolls/)
	symlinkDir, err := os.MkdirTemp("", "scribe-test-symlinks-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(symlinkDir) }()

	// Create symlinks pointing to the real skill directories
	_ = os.Symlink(skill1Dir, filepath.Join(symlinkDir, "skill1"))
	_ = os.Symlink(skill2Dir, filepath.Join(symlinkDir, "skill2"))
	_ = os.Symlink(noSkillDir, filepath.Join(symlinkDir, "no-skill"))

	count = countSkillsInDir(symlinkDir)
	if count != 2 {
		t.Errorf("countSkillsInDir() with symlinks = %d, want 2", count)
	}
}

func TestIsSymlink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a regular file
	regularFile := filepath.Join(tmpDir, "regular.txt")
	_ = os.WriteFile(regularFile, []byte("test"), 0o644)

	if IsSymlink(regularFile) {
		t.Error("Regular file should not be a symlink")
	}

	// Create a directory
	dir := filepath.Join(tmpDir, "dir")
	_ = os.MkdirAll(dir, 0o755)

	if IsSymlink(dir) {
		t.Error("Directory should not be a symlink")
	}

	// Create a symlink
	symlinkPath := filepath.Join(tmpDir, "link")
	_ = os.Symlink(regularFile, symlinkPath)

	if !IsSymlink(symlinkPath) {
		t.Error("Symlink should be detected as symlink")
	}

	// Non-existent path
	if IsSymlink(filepath.Join(tmpDir, "non-existent")) {
		t.Error("Non-existent path should not be a symlink")
	}
}

func TestGetSymlinkTarget(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create target file
	targetFile := filepath.Join(tmpDir, "target.txt")
	_ = os.WriteFile(targetFile, []byte("test"), 0o644)

	// Create symlink
	symlinkPath := filepath.Join(tmpDir, "link")
	_ = os.Symlink(targetFile, symlinkPath)

	target, err := GetSymlinkTarget(symlinkPath)
	if err != nil {
		t.Fatalf("GetSymlinkTarget() error: %v", err)
	}

	if target != targetFile {
		t.Errorf("GetSymlinkTarget() = %q, want %q", target, targetFile)
	}

	// Non-symlink should error
	_, err = GetSymlinkTarget(targetFile)
	if err == nil {
		t.Error("GetSymlinkTarget() should error for non-symlink")
	}
}

func TestUpdateSkillMeta(t *testing.T) {
	source := &SourceInfo{
		Type:  "github",
		Owner: "user",
		Repo:  "repo",
	}
	meta := NewSkillMeta(source, "skills/test", "original content", nil)

	originalHash := meta.ContentHash

	// Update with new content
	UpdateSkillMeta(meta, "new content", nil)

	if meta.ContentHash == originalHash {
		t.Error("ContentHash should change after update")
	}
	// Note: UpdatedAt may not change if test runs within same second,
	// so we just verify it's set (not empty)
	if meta.UpdatedAt == "" {
		t.Error("UpdatedAt should be set after update")
	}
}

func TestLoadSkillWithMeta(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	skillDir := filepath.Join(tmpDir, "test-skill")
	_ = os.MkdirAll(skillDir, 0o755)

	// Create SKILL.md
	skillContent := `---
name: test-skill
description: A test skill
---
# Test Skill
`
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644)

	// Load without meta
	skill, err := LoadSkillWithMeta(skillDir)
	if err != nil {
		t.Fatalf("LoadSkillWithMeta() error: %v", err)
	}

	if skill.Name != "test-skill" {
		t.Errorf("skill.Name = %q, want 'test-skill'", skill.Name)
	}
	if skill.Meta != nil {
		t.Error("skill.Meta should be nil when no meta file exists")
	}

	// Create meta file
	source := &SourceInfo{Type: "local", LocalPath: skillDir}
	meta := NewSkillMeta(source, "", skillContent, nil)
	_ = WriteSkillMeta(filepath.Join(skillDir, ".scribe-meta.json"), meta)

	// Load with meta
	skill, err = LoadSkillWithMeta(skillDir)
	if err != nil {
		t.Fatalf("LoadSkillWithMeta() with meta error: %v", err)
	}

	if skill.Meta == nil {
		t.Error("skill.Meta should not be nil when meta file exists")
	}
}

func TestListSkillsWithMeta(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	scrollsDir := filepath.Join(tmpDir, "scrolls")

	// Non-existent directory should return empty
	skills, err := ListSkillsWithMeta(filepath.Join(tmpDir, "non-existent"))
	if err != nil {
		t.Fatalf("ListSkillsWithMeta() error: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("ListSkillsWithMeta() = %d skills, want 0", len(skills))
	}

	// Create skills directory with skills
	_ = os.MkdirAll(scrollsDir, 0o755)

	skill1Dir := filepath.Join(scrollsDir, "skill-one")
	skill2Dir := filepath.Join(scrollsDir, "skill-two")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)

	skill1Content := `---
name: skill-one
description: First skill
---
# Skill One
`
	skill2Content := `---
name: skill-two
description: Second skill
---
# Skill Two
`
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skill1Content), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skill2Content), 0o644)

	skills, err = ListSkillsWithMeta(scrollsDir)
	if err != nil {
		t.Fatalf("ListSkillsWithMeta() error: %v", err)
	}

	if len(skills) != 2 {
		t.Errorf("ListSkillsWithMeta() = %d skills, want 2", len(skills))
	}
}

func TestSkillNeedsUpdate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	skillDir := filepath.Join(tmpDir, "test-skill")
	_ = os.MkdirAll(skillDir, 0o755)

	content := "original content"

	// No meta file - should need update
	needsUpdate, err := SkillNeedsUpdate(skillDir, content)
	if err != nil {
		t.Fatalf("SkillNeedsUpdate() error: %v", err)
	}
	if !needsUpdate {
		t.Error("Should need update when no meta file exists")
	}

	// Create meta file
	source := &SourceInfo{Type: "local", LocalPath: skillDir}
	meta := NewSkillMeta(source, "", content, nil)
	_ = WriteSkillMeta(filepath.Join(skillDir, ".scribe-meta.json"), meta)

	// Same content - should not need update
	needsUpdate, err = SkillNeedsUpdate(skillDir, content)
	if err != nil {
		t.Fatalf("SkillNeedsUpdate() error: %v", err)
	}
	if needsUpdate {
		t.Error("Should not need update when content is same")
	}

	// Different content - should need update
	needsUpdate, err = SkillNeedsUpdate(skillDir, "different content")
	if err != nil {
		t.Fatalf("SkillNeedsUpdate() error: %v", err)
	}
	if !needsUpdate {
		t.Error("Should need update when content is different")
	}
}

func TestValidateSkill(t *testing.T) {
	// Nil skill
	err := ValidateSkill(nil)
	if err != ErrInvalidSkill {
		t.Errorf("ValidateSkill(nil) = %v, want ErrInvalidSkill", err)
	}

	// Missing name
	err = ValidateSkill(&Skill{Description: "desc"})
	if err != ErrMissingName {
		t.Errorf("ValidateSkill(no name) = %v, want ErrMissingName", err)
	}

	// Missing description
	err = ValidateSkill(&Skill{Name: "name"})
	if err != ErrMissingDesc {
		t.Errorf("ValidateSkill(no desc) = %v, want ErrMissingDesc", err)
	}

	// Valid skill
	err = ValidateSkill(&Skill{Name: "name", Description: "desc"})
	if err != nil {
		t.Errorf("ValidateSkill(valid) = %v, want nil", err)
	}
}

func TestFormatSource(t *testing.T) {
	tests := []struct {
		name     string
		source   *SourceInfo
		expected string
	}{
		{
			name: "github default branch",
			source: &SourceInfo{
				Type:  "github",
				Owner: "user",
				Repo:  "repo",
				Ref:   "main",
			},
			expected: "user/repo",
		},
		{
			name: "github with custom ref",
			source: &SourceInfo{
				Type:  "github",
				Owner: "user",
				Repo:  "repo",
				Ref:   "v1.0.0",
			},
			expected: "user/repo#v1.0.0",
		},
		{
			name: "gitlab",
			source: &SourceInfo{
				Type:  "gitlab",
				Owner: "group",
				Repo:  "project",
				Ref:   "develop",
			},
			expected: "group/project#develop",
		},
		{
			name: "local",
			source: &SourceInfo{
				Type:      "local",
				LocalPath: "/path/to/skill",
			},
			expected: "/path/to/skill",
		},
		{
			name: "url",
			source: &SourceInfo{
				Type: "url",
				URL:  "https://example.com/skill.zip",
			},
			expected: "https://example.com/skill.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSource(tt.source)
			if result != tt.expected {
				t.Errorf("formatSource() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetWorkspaceInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create a test workspace
	ws := &Workspace{
		Name:        "test-workspace",
		Description: "A test workspace",
		Skills:      []string{"skill-1", "skill-2"},
	}
	_ = CreateWorkspace(ws)

	// GetWorkspaceInfo returns info for all workspaces
	infos, err := GetWorkspaceInfo()
	if err != nil {
		t.Fatalf("GetWorkspaceInfo() error: %v", err)
	}

	// Should have at least the test workspace and default
	if len(infos) < 1 {
		t.Errorf("GetWorkspaceInfo() returned %d infos, want at least 1", len(infos))
	}

	// Find our test workspace
	var found *WorkspaceInfo
	for i := range infos {
		if infos[i].Name == "test-workspace" {
			found = &infos[i]
			break
		}
	}

	if found == nil {
		t.Fatal("test-workspace not found in infos")
	}
	if found.Description != ws.Description {
		t.Errorf("info.Description = %q, want %q", found.Description, ws.Description)
	}
	if len(found.Skills) != 2 {
		t.Errorf("info.Skills count = %d, want 2", len(found.Skills))
	}
}

func TestGetSkillInfo(t *testing.T) {
	skill := &Skill{
		Name:        "test-skill",
		Description: "A test skill",
	}

	info := GetSkillInfo(skill)

	if info.Name != skill.Name {
		t.Errorf("info.Name = %q, want %q", info.Name, skill.Name)
	}
	if info.Description != skill.Description {
		t.Errorf("info.Description = %q, want %q", info.Description, skill.Description)
	}

	// With metadata
	skill.Meta = &SkillMeta{
		Source:      "user/repo",
		SourceType:  "github",
		InstalledAt: "2024-01-01T00:00:00Z",
	}

	info = GetSkillInfo(skill)

	if info.Source != skill.Meta.Source {
		t.Errorf("info.Source = %q, want %q", info.Source, skill.Meta.Source)
	}
	if info.SourceType != skill.Meta.SourceType {
		t.Errorf("info.SourceType = %q, want %q", info.SourceType, skill.Meta.SourceType)
	}
}

func TestSyncAllSkillsToAgents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create skills
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skill1Dir := filepath.Join(scrollsDir, "skill-one")
	skill2Dir := filepath.Join(scrollsDir, "skill-two")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)

	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Skill 1"), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Skill 2"), 0o644)

	// Create agent directories
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	cursorSkillsDir := filepath.Join(tmpDir, ".cursor", "skills")
	_ = os.MkdirAll(claudeSkillsDir, 0o755)
	_ = os.MkdirAll(cursorSkillsDir, 0o755)

	// Sync all
	err = SyncAllSkillsToAgents()
	if err != nil {
		t.Fatalf("SyncAllSkillsToAgents() error: %v", err)
	}

	// Verify skills were synced to Claude
	if _, err := os.Lstat(filepath.Join(claudeSkillsDir, "skill-one")); err != nil {
		t.Error("skill-one should be synced to Claude")
	}
	if _, err := os.Lstat(filepath.Join(claudeSkillsDir, "skill-two")); err != nil {
		t.Error("skill-two should be synced to Claude")
	}

	// Verify skills were synced to Cursor
	if _, err := os.Lstat(filepath.Join(cursorSkillsDir, "skill-one")); err != nil {
		t.Error("skill-one should be synced to Cursor")
	}
	if _, err := os.Lstat(filepath.Join(cursorSkillsDir, "skill-two")); err != nil {
		t.Error("skill-two should be synced to Cursor")
	}
}

func TestCreateSymlink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create target directory
	targetDir := filepath.Join(tmpDir, "target")
	_ = os.MkdirAll(targetDir, 0o755)
	_ = os.WriteFile(filepath.Join(targetDir, "test.txt"), []byte("test"), 0o644)

	// Create symlink
	linkPath := filepath.Join(tmpDir, "link")
	err = CreateSymlink(targetDir, linkPath)
	if err != nil {
		t.Fatalf("CreateSymlink() error: %v", err)
	}

	// Verify symlink exists
	if !IsSymlink(linkPath) {
		t.Error("Link should be a symlink")
	}

	// Verify we can read through the symlink
	content, err := os.ReadFile(filepath.Join(linkPath, "test.txt"))
	if err != nil {
		t.Fatalf("Failed to read through symlink: %v", err)
	}
	if string(content) != "test" {
		t.Errorf("Content = %q, want 'test'", string(content))
	}
}

func TestDirExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Directory exists
	if !dirExists(tmpDir) {
		t.Error("dirExists() should return true for existing directory")
	}

	// File is not a directory
	filePath := filepath.Join(tmpDir, "file.txt")
	_ = os.WriteFile(filePath, []byte("test"), 0o644)
	if dirExists(filePath) {
		t.Error("dirExists() should return false for file")
	}

	// Non-existent path
	if dirExists(filepath.Join(tmpDir, "non-existent")) {
		t.Error("dirExists() should return false for non-existent path")
	}
}

func TestParseSkillContent(t *testing.T) {
	content := `---
name: inline-skill
description: Parsed from content
custom_field: value
---
# Inline Skill

This is the body content.
`

	skill, err := ParseSkillContent(content, "/fake/path")
	if err != nil {
		t.Fatalf("ParseSkillContent() error: %v", err)
	}

	if skill.Name != "inline-skill" {
		t.Errorf("skill.Name = %q, want 'inline-skill'", skill.Name)
	}
	if skill.Description != "Parsed from content" {
		t.Errorf("skill.Description = %q, want 'Parsed from content'", skill.Description)
	}
	if skill.Path != "/fake/path" {
		t.Errorf("skill.Path = %q, want '/fake/path'", skill.Path)
	}
	if skill.Content != "# Inline Skill\n\nThis is the body content." {
		t.Errorf("skill.Content = %q", skill.Content)
	}

	// Check that custom field is in metadata
	if skill.Metadata == nil {
		t.Error("skill.Metadata should not be nil")
	} else if skill.Metadata["custom_field"] != "value" {
		t.Errorf("skill.Metadata['custom_field'] = %v, want 'value'", skill.Metadata["custom_field"])
	}
}

func TestParseSkillContent_Errors(t *testing.T) {
	// Missing name
	content := `---
description: No name
---
Body
`
	_, err := ParseSkillContent(content, "/path")
	if err != ErrMissingName {
		t.Errorf("Expected ErrMissingName, got %v", err)
	}

	// Missing description
	content = `---
name: no-desc
---
Body
`
	_, err = ParseSkillContent(content, "/path")
	if err != ErrMissingDesc {
		t.Errorf("Expected ErrMissingDesc, got %v", err)
	}
}

// ============================================================================
// Additional Coverage Tests - Phase 2
// ============================================================================

func TestReadSkill(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create a skill with valid SKILL.md
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "test-skill")
	_ = os.MkdirAll(skillDir, 0o755)

	skillContent := `---
name: test-skill
description: A test skill
---
# Test Skill Content
`
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644)

	// Read skill without meta
	skill, err := ReadSkill("test-skill")
	if err != nil {
		t.Fatalf("ReadSkill() error: %v", err)
	}

	if skill.Name != "test-skill" {
		t.Errorf("skill.Name = %q, want 'test-skill'", skill.Name)
	}

	// Add meta and read again
	source := &SourceInfo{Type: "local", LocalPath: skillDir}
	meta := NewSkillMeta(source, "", skillContent, nil)
	_ = WriteSkillMeta(filepath.Join(skillDir, ".scribe-meta.json"), meta)

	skill, err = ReadSkill("test-skill")
	if err != nil {
		t.Fatalf("ReadSkill() with meta error: %v", err)
	}

	if skill.Meta == nil {
		t.Error("skill.Meta should not be nil when meta file exists")
	}
}

func TestReadAllSkills(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Empty initially
	skills, err := ReadAllSkills()
	if err != nil {
		t.Fatalf("ReadAllSkills() error: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("ReadAllSkills() = %d skills, want 0", len(skills))
	}

	// Create skills
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")

	skill1Dir := filepath.Join(scrollsDir, "skill-1")
	skill2Dir := filepath.Join(scrollsDir, "skill-2")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)

	skillContent1 := `---
name: skill-1
description: First skill
---
# Skill 1
`
	skillContent2 := `---
name: skill-2
description: Second skill
---
# Skill 2
`
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skillContent1), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skillContent2), 0o644)

	skills, err = ReadAllSkills()
	if err != nil {
		t.Fatalf("ReadAllSkills() error: %v", err)
	}
	if len(skills) != 2 {
		t.Errorf("ReadAllSkills() = %d skills, want 2", len(skills))
	}
}

func TestGetAllSkillInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create a skill
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "info-skill")
	_ = os.MkdirAll(skillDir, 0o755)

	skillContent := `---
name: info-skill
description: Skill for info test
---
# Info Skill
`
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644)

	// Create Claude agent directory with the skill
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	_ = os.MkdirAll(claudeSkillsDir, 0o755)
	_ = os.MkdirAll(filepath.Join(claudeSkillsDir, "info-skill"), 0o755)
	_ = os.WriteFile(filepath.Join(claudeSkillsDir, "info-skill", "SKILL.md"), []byte(skillContent), 0o644)

	infos, err := GetAllSkillInfo()
	if err != nil {
		t.Fatalf("GetAllSkillInfo() error: %v", err)
	}

	if len(infos) != 1 {
		t.Errorf("GetAllSkillInfo() = %d infos, want 1", len(infos))
	}

	if len(infos) > 0 {
		if infos[0].Name != "info-skill" {
			t.Errorf("info.Name = %q, want 'info-skill'", infos[0].Name)
		}
		// Should detect Claude as an agent with this skill
		if len(infos[0].Agents) != 1 || infos[0].Agents[0] != "claude-code" {
			t.Errorf("info.Agents = %v, want ['claude-code']", infos[0].Agents)
		}
	}
}

func TestGetActiveWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Default should be active
	ws, err := GetActiveWorkspace()
	if err != nil {
		t.Fatalf("GetActiveWorkspace() error: %v", err)
	}

	if ws.Name != "default" {
		t.Errorf("Active workspace = %q, want 'default'", ws.Name)
	}
}

func TestSetActiveWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create a new workspace
	ws := &Workspace{
		Name:        "new-workspace",
		Description: "A new workspace",
		Skills:      []string{},
	}
	_ = CreateWorkspace(ws)

	// Set it as active
	err = SetActiveWorkspace("new-workspace")
	if err != nil {
		t.Fatalf("SetActiveWorkspace() error: %v", err)
	}

	// Verify it's now active
	config, _ := LoadConfig()
	if config.ActiveWorkspace != "new-workspace" {
		t.Errorf("ActiveWorkspace = %q, want 'new-workspace'", config.ActiveWorkspace)
	}

	// Switch back to default
	err = SetActiveWorkspace("default")
	if err != nil {
		t.Fatalf("SetActiveWorkspace('default') error: %v", err)
	}

	config, _ = LoadConfig()
	if config.ActiveWorkspace != "default" {
		t.Errorf("ActiveWorkspace = %q, want 'default'", config.ActiveWorkspace)
	}
}

func TestSyncWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create skills
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skill1Dir := filepath.Join(scrollsDir, "sync-skill-1")
	skill2Dir := filepath.Join(scrollsDir, "sync-skill-2")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Skill 1"), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Skill 2"), 0o644)

	// Create Claude directory
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	_ = os.MkdirAll(claudeSkillsDir, 0o755)

	current := &Workspace{Name: "current", Skills: []string{"sync-skill-1"}}
	target := &Workspace{Name: "target", Skills: []string{"sync-skill-2"}}

	err = SyncWorkspace(current, target)
	if err != nil {
		t.Fatalf("SyncWorkspace() error: %v", err)
	}

	// sync-skill-1 should be removed, sync-skill-2 should be added
	if _, err := os.Lstat(filepath.Join(claudeSkillsDir, "sync-skill-1")); !os.IsNotExist(err) {
		t.Error("sync-skill-1 should be removed after sync")
	}
	if _, err := os.Lstat(filepath.Join(claudeSkillsDir, "sync-skill-2")); err != nil {
		t.Error("sync-skill-2 should be added after sync")
	}
}

func TestResyncCurrentWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create skills in scrolls dir
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skill1Dir := filepath.Join(scrollsDir, "resync-skill-1")
	skill2Dir := filepath.Join(scrollsDir, "resync-skill-2")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Skill 1"), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Skill 2"), 0o644)

	// Create Claude directory (simulates installed agent)
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	_ = os.MkdirAll(claudeSkillsDir, 0o755)

	// Create and save a workspace with both skills
	ws := &Workspace{
		Name:   "default",
		Skills: []string{"resync-skill-1", "resync-skill-2"},
	}
	if err := saveWorkspace(ws); err != nil {
		t.Fatalf("failed to save workspace: %v", err)
	}

	// Set it as active
	config := &Config{ActiveWorkspace: "default"}
	if err := SaveConfig(config); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Skills not in agent dir yet (simulates missing symlinks)
	if _, err := os.Lstat(filepath.Join(claudeSkillsDir, "resync-skill-1")); !os.IsNotExist(err) {
		t.Error("resync-skill-1 should not exist before resync")
	}

	// Resync should create symlinks
	err = ResyncCurrentWorkspace()
	if err != nil {
		t.Fatalf("ResyncCurrentWorkspace() error: %v", err)
	}

	// Both skills should now be symlinked
	if _, err := os.Lstat(filepath.Join(claudeSkillsDir, "resync-skill-1")); err != nil {
		t.Error("resync-skill-1 should exist after resync")
	}
	if _, err := os.Lstat(filepath.Join(claudeSkillsDir, "resync-skill-2")); err != nil {
		t.Error("resync-skill-2 should exist after resync")
	}
}

func TestSkillDiff(t *testing.T) {
	a := []string{"skill-1", "skill-2", "skill-3"}
	b := []string{"skill-2", "skill-4"}

	diff := skillDiff(a, b)

	// Should contain skill-1 and skill-3 (in a but not in b)
	if len(diff) != 2 {
		t.Errorf("skillDiff() = %v, want 2 items", diff)
	}

	found1, found3 := false, false
	for _, s := range diff {
		if s == "skill-1" {
			found1 = true
		}
		if s == "skill-3" {
			found3 = true
		}
	}

	if !found1 {
		t.Error("skill-1 should be in diff")
	}
	if !found3 {
		t.Error("skill-3 should be in diff")
	}
}

func TestRebuildDefaultWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create some skills
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skill1Dir := filepath.Join(scrollsDir, "rebuild-skill-1")
	skill2Dir := filepath.Join(scrollsDir, "rebuild-skill-2")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Skill 1"), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Skill 2"), 0o644)

	// Rebuild default workspace
	err = RebuildDefaultWorkspace()
	if err != nil {
		t.Fatalf("RebuildDefaultWorkspace() error: %v", err)
	}

	// Default workspace should have both skills
	ws, err := GetWorkspace("default")
	if err != nil {
		t.Fatalf("GetWorkspace('default') error: %v", err)
	}

	if len(ws.Skills) != 2 {
		t.Errorf("default workspace skills = %d, want 2", len(ws.Skills))
	}
}

func TestCleanWorkspaces(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create a skill
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "existing-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Existing"), 0o644)

	// Create workspace with existing and non-existing skills
	ws := &Workspace{
		Name:   "test-ws",
		Skills: []string{"existing-skill", "deleted-skill"},
	}
	_ = CreateWorkspace(ws)

	// Clean workspaces
	err = CleanWorkspaces()
	if err != nil {
		t.Fatalf("CleanWorkspaces() error: %v", err)
	}

	// Workspace should only have existing-skill
	ws, _ = GetWorkspace("test-ws")
	if len(ws.Skills) != 1 {
		t.Errorf("workspace skills = %d, want 1", len(ws.Skills))
	}
	if ws.Skills[0] != "existing-skill" {
		t.Errorf("workspace skill = %q, want 'existing-skill'", ws.Skills[0])
	}
}

func TestSaveSkillWithMeta(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create source skill
	sourceDir := filepath.Join(tmpDir, "source")
	_ = os.MkdirAll(sourceDir, 0o755)

	skillContent := `---
name: save-skill
description: Skill for save test
---
# Save Skill Content
`
	_ = os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(skillContent), 0o644)

	skill := &Skill{
		Name:        "save-skill",
		Description: "Skill for save test",
		Path:        sourceDir,
	}

	source := &SourceInfo{
		Type:  "github",
		Owner: "user",
		Repo:  "repo",
	}

	// Save to target directory
	targetDir := filepath.Join(tmpDir, "target", "save-skill")
	err = SaveSkillWithMeta(targetDir, skill, source, "skills/save-skill", nil)
	if err != nil {
		t.Fatalf("SaveSkillWithMeta() error: %v", err)
	}

	// Verify SKILL.md was saved
	savedSkillPath := filepath.Join(targetDir, "SKILL.md")
	if _, err := os.Stat(savedSkillPath); err != nil {
		t.Error("SKILL.md should exist after save")
	}

	// Verify meta was saved
	savedMetaPath := filepath.Join(targetDir, ".scribe-meta.json")
	if _, err := os.Stat(savedMetaPath); err != nil {
		t.Error(".scribe-meta.json should exist after save")
	}

	// Verify meta content
	meta, err := ReadSkillMeta(savedMetaPath)
	if err != nil {
		t.Fatalf("ReadSkillMeta() error: %v", err)
	}
	if meta.Source != "user/repo" {
		t.Errorf("meta.Source = %q, want 'user/repo'", meta.Source)
	}
	if meta.SkillPath != "skills/save-skill" {
		t.Errorf("meta.SkillPath = %q, want 'skills/save-skill'", meta.SkillPath)
	}
}

// ======================================================================
// URL Scheme Tests
// ======================================================================

func TestParseInstallURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantType  string
		wantOwner string
		wantRepo  string
		wantSkill string
		wantRef   string
		wantErr   bool
	}{
		{
			name:      "GitHub shorthand",
			url:       "agenthub://install?repo=user/repo",
			wantType:  "github",
			wantOwner: "user",
			wantRepo:  "repo",
		},
		{
			name:      "GitHub with source type",
			url:       "agenthub://install?source=github&repo=owner/myrepo",
			wantType:  "github",
			wantOwner: "owner",
			wantRepo:  "myrepo",
		},
		{
			name:      "GitHub with skill filter",
			url:       "agenthub://install?repo=user/repo&name=my-skill",
			wantType:  "github",
			wantOwner: "user",
			wantRepo:  "repo",
			wantSkill: "my-skill",
		},
		{
			name:      "GitHub with ref",
			url:       "agenthub://install?repo=user/repo&ref=main",
			wantType:  "github",
			wantOwner: "user",
			wantRepo:  "repo",
			wantRef:   "main",
		},
		{
			name:      "GitLab source",
			url:       "agenthub://install?source=gitlab&repo=user/project",
			wantType:  "gitlab",
			wantOwner: "user",
			wantRepo:  "project",
		},
		{
			name:    "Missing repo parameter",
			url:     "agenthub://install?source=github",
			wantErr: true,
		},
		{
			name:    "Wrong scheme",
			url:     "https://install?repo=user/repo",
			wantErr: true,
		},
		{
			name:    "Invalid repo format",
			url:     "agenthub://install?repo=noslash",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, skill, err := ParseInstallURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseInstallURL() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseInstallURL() error = %v", err)
			}

			if source.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", source.Type, tt.wantType)
			}
			if source.Owner != tt.wantOwner {
				t.Errorf("Owner = %q, want %q", source.Owner, tt.wantOwner)
			}
			if source.Repo != tt.wantRepo {
				t.Errorf("Repo = %q, want %q", source.Repo, tt.wantRepo)
			}
			if skill != tt.wantSkill {
				t.Errorf("skillFilter = %q, want %q", skill, tt.wantSkill)
			}
			if source.Ref != tt.wantRef {
				t.Errorf("Ref = %q, want %q", source.Ref, tt.wantRef)
			}
		})
	}
}

func TestParseInstallURL_Subpath(t *testing.T) {
	source, _, err := ParseInstallURL("agenthub://install?repo=user/repo/path/to/skills")
	if err != nil {
		t.Fatalf("ParseInstallURL() error = %v", err)
	}

	if source.Owner != "user" {
		t.Errorf("Owner = %q, want 'user'", source.Owner)
	}
	if source.Repo != "repo" {
		t.Errorf("Repo = %q, want 'repo'", source.Repo)
	}
	if source.Subpath != "path/to/skills" {
		t.Errorf("Subpath = %q, want 'path/to/skills'", source.Subpath)
	}
}

func TestFilterSkillsByName(t *testing.T) {
	skills := []*Skill{
		{Name: "skill-a", Description: "A"},
		{Name: "skill-b", Description: "B"},
		{Name: "skill-c", Description: "C"},
	}

	filtered := filterSkillsByName(skills, "skill-b")
	if len(filtered) != 1 {
		t.Fatalf("filterSkillsByName() returned %d skills, want 1", len(filtered))
	}
	if filtered[0].Name != "skill-b" {
		t.Errorf("filterSkillsByName() returned %q, want 'skill-b'", filtered[0].Name)
	}

	// Non-existent skill
	filtered = filterSkillsByName(skills, "nonexistent")
	if len(filtered) != 0 {
		t.Errorf("filterSkillsByName() returned %d skills, want 0", len(filtered))
	}
}

// ======================================================================
// Integration Tests - Full Flows
// ======================================================================

// TestIntegration_MultiSkillInstallAndRemove tests installing multiple skills
// and then removing them, verifying the full lifecycle.
func TestIntegration_MultiSkillInstallAndRemove(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Initialize Scribe directories
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create mock agent directories (Claude and Cursor)
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	cursorSkillsDir := filepath.Join(tmpDir, ".cursor", "skills")
	_ = os.MkdirAll(claudeSkillsDir, 0o755)
	_ = os.MkdirAll(cursorSkillsDir, 0o755)

	// Create source directory with multiple skills
	sourceDir := filepath.Join(tmpDir, "source-repo")
	skills := []struct {
		name        string
		description string
		content     string
	}{
		{"react-patterns", "React best practices", "# React Patterns\nUse functional components."},
		{"typescript-tips", "TypeScript guidance", "# TypeScript Tips\nPrefer interfaces over types."},
		{"go-idioms", "Go coding patterns", "# Go Idioms\nHandle errors explicitly."},
	}

	for _, s := range skills {
		skillDir := filepath.Join(sourceDir, "skills", s.name)
		_ = os.MkdirAll(skillDir, 0o755)
		content := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n%s\n", s.name, s.description, s.content)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644)
	}

	// Discover skills from source
	discovered, err := DiscoverSkills(sourceDir)
	if err != nil {
		t.Fatalf("DiscoverSkills() error: %v", err)
	}
	if len(discovered) != 3 {
		t.Fatalf("Expected 3 skills, got %d", len(discovered))
	}

	// Install all skills
	source := &SourceInfo{Type: "local", LocalPath: sourceDir}
	for _, skill := range discovered {
		err = InstallSkill(skill, source, InstallOptions{}, nil)
		if err != nil {
			t.Fatalf("InstallSkill(%s) error: %v", skill.Name, err)
		}
		_ = AddSkillToActiveAndDefaultWorkspace(skill.Name)
	}

	// Verify all skills exist in canonical location
	for _, s := range skills {
		exists, _ := SkillExists(s.name)
		if !exists {
			t.Errorf("Skill %s should exist in canonical location", s.name)
		}
	}

	// Verify symlinks exist in both agent directories
	for _, s := range skills {
		claudeLink := filepath.Join(claudeSkillsDir, s.name)
		if !IsSymlink(claudeLink) {
			t.Errorf("Skill %s should have symlink in Claude directory", s.name)
		}
		cursorLink := filepath.Join(cursorSkillsDir, s.name)
		if !IsSymlink(cursorLink) {
			t.Errorf("Skill %s should have symlink in Cursor directory", s.name)
		}
	}

	// Verify all skills are in default workspace
	ws, _ := GetWorkspace("default")
	for _, s := range skills {
		found := false
		for _, wsSkill := range ws.Skills {
			if wsSkill == s.name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Skill %s should be in default workspace", s.name)
		}
	}

	// Remove one skill
	_ = RemoveSkillFromAllWorkspaces("typescript-tips")
	_ = UninstallSkill("typescript-tips")

	// Verify it's gone
	exists, _ := SkillExists("typescript-tips")
	if exists {
		t.Error("typescript-tips should not exist after uninstall")
	}
	if IsSymlink(filepath.Join(claudeSkillsDir, "typescript-tips")) {
		t.Error("typescript-tips symlink should be removed from Claude")
	}

	// Verify others still exist
	exists, _ = SkillExists("react-patterns")
	if !exists {
		t.Error("react-patterns should still exist")
	}
	exists, _ = SkillExists("go-idioms")
	if !exists {
		t.Error("go-idioms should still exist")
	}

	// Remove all remaining skills
	for _, name := range []string{"react-patterns", "go-idioms"} {
		_ = RemoveSkillFromAllWorkspaces(name)
		_ = UninstallSkill(name)
	}

	// Verify all gone
	installed, _ := ListInstalledSkills()
	if len(installed) != 0 {
		t.Errorf("Expected 0 installed skills, got %d", len(installed))
	}
}

// TestIntegration_WorkspaceSwitching tests creating workspaces, adding skills,
// and switching between workspaces.
func TestIntegration_WorkspaceSwitching(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Initialize
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create mock agent directory
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	_ = os.MkdirAll(claudeSkillsDir, 0o755)

	// Create and install 4 skills
	skillNames := []string{"frontend-skill", "backend-skill", "devops-skill", "testing-skill"}
	for _, name := range skillNames {
		sourceDir := filepath.Join(tmpDir, "source", name)
		_ = os.MkdirAll(sourceDir, 0o755)
		content := fmt.Sprintf("---\nname: %s\ndescription: %s skill\n---\n# %s\n", name, name, name)
		_ = os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(content), 0o644)

		skill, _ := ParseSkillMd(filepath.Join(sourceDir, "SKILL.md"))
		source := &SourceInfo{Type: "local", LocalPath: sourceDir}
		_ = InstallSkill(skill, source, InstallOptions{}, nil)
		_ = AddSkillToActiveAndDefaultWorkspace(name)
	}

	// Create workspaces with different skill sets
	_ = CreateWorkspace(&Workspace{
		Name:        "frontend",
		Description: "Frontend development",
		Skills:      []string{"frontend-skill", "testing-skill"},
	})

	_ = CreateWorkspace(&Workspace{
		Name:        "backend",
		Description: "Backend development",
		Skills:      []string{"backend-skill", "devops-skill", "testing-skill"},
	})

	// Verify default workspace has all skills
	defaultWs, _ := GetWorkspace("default")
	if len(defaultWs.Skills) != 4 {
		t.Errorf("Default workspace should have 4 skills, got %d", len(defaultWs.Skills))
	}

	// Switch to frontend workspace
	current, _ := GetActiveWorkspace()
	frontend, _ := GetWorkspace("frontend")
	_ = SyncWorkspace(current, frontend)
	_ = SetActiveWorkspace("frontend")

	// Verify only frontend skills are symlinked
	frontendLinks := []string{"frontend-skill", "testing-skill"}
	backendLinks := []string{"backend-skill", "devops-skill"}

	for _, name := range frontendLinks {
		if !IsSymlink(filepath.Join(claudeSkillsDir, name)) {
			t.Errorf("Skill %s should be symlinked in frontend workspace", name)
		}
	}
	for _, name := range backendLinks {
		if IsSymlink(filepath.Join(claudeSkillsDir, name)) {
			t.Errorf("Skill %s should NOT be symlinked in frontend workspace", name)
		}
	}

	// Switch to backend workspace
	current, _ = GetActiveWorkspace()
	backend, _ := GetWorkspace("backend")
	_ = SyncWorkspace(current, backend)
	_ = SetActiveWorkspace("backend")

	// Verify backend skills are symlinked
	expectedBackend := []string{"backend-skill", "devops-skill", "testing-skill"}
	notExpected := []string{"frontend-skill"}

	for _, name := range expectedBackend {
		if !IsSymlink(filepath.Join(claudeSkillsDir, name)) {
			t.Errorf("Skill %s should be symlinked in backend workspace", name)
		}
	}
	for _, name := range notExpected {
		if IsSymlink(filepath.Join(claudeSkillsDir, name)) {
			t.Errorf("Skill %s should NOT be symlinked in backend workspace", name)
		}
	}

	// Switch back to default
	current, _ = GetActiveWorkspace()
	defaultWs, _ = GetWorkspace("default")
	_ = SyncWorkspace(current, defaultWs)
	_ = SetActiveWorkspace("default")

	// Verify all skills are symlinked again
	for _, name := range skillNames {
		if !IsSymlink(filepath.Join(claudeSkillsDir, name)) {
			t.Errorf("Skill %s should be symlinked in default workspace", name)
		}
	}
}

// TestIntegration_SkillDiscoveryFromNestedStructure tests discovering skills
// from various nested directory structures.
func TestIntegration_SkillDiscoveryFromNestedStructure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a complex nested structure like a real repo might have
	// repo/
	//   skills/
	//     react/SKILL.md
	//     vue/SKILL.md
	//   .claude/skills/
	//     internal-skill/SKILL.md
	//   docs/  (no skills here)
	//   SKILL.md  (root level skill)

	createSkill := func(dir, name, desc string) {
		_ = os.MkdirAll(dir, 0o755)
		content := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n# %s\n", name, desc, name)
		_ = os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)
	}

	createSkill(filepath.Join(tmpDir, "skills", "react"), "react-skill", "React skill")
	createSkill(filepath.Join(tmpDir, "skills", "vue"), "vue-skill", "Vue skill")
	createSkill(filepath.Join(tmpDir, ".claude", "skills", "internal-skill"), "internal-skill", "Internal")
	createSkill(tmpDir, "root-skill", "Root level skill")

	// Create a docs directory with no skills
	_ = os.MkdirAll(filepath.Join(tmpDir, "docs"), 0o755)
	_ = os.WriteFile(filepath.Join(tmpDir, "docs", "README.md"), []byte("# Docs"), 0o644)

	// Discover all skills
	skills, err := DiscoverSkills(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverSkills() error: %v", err)
	}

	// Should find 4 skills
	if len(skills) != 4 {
		t.Errorf("Expected 4 skills, got %d", len(skills))
		for _, s := range skills {
			t.Logf("Found: %s", s.Name)
		}
	}

	// Verify all expected skills are found
	expectedNames := map[string]bool{
		"react-skill":    false,
		"vue-skill":      false,
		"internal-skill": false,
		"root-skill":     false,
	}

	for _, skill := range skills {
		if _, ok := expectedNames[skill.Name]; ok {
			expectedNames[skill.Name] = true
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected to find skill %s", name)
		}
	}
}

// TestIntegration_SkillMetadataTracking tests that skill metadata is properly
// tracked and can be used for update detection.
func TestIntegration_SkillMetadataTracking(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create and install a skill
	sourceDir := filepath.Join(tmpDir, "source", "tracked-skill")
	_ = os.MkdirAll(sourceDir, 0o755)
	originalContent := "---\nname: tracked-skill\ndescription: A tracked skill\n---\n# Version 1\n"
	_ = os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(originalContent), 0o644)

	skill, _ := ParseSkillMd(filepath.Join(sourceDir, "SKILL.md"))
	source := &SourceInfo{
		Type:  "github",
		Owner: "test-org",
		Repo:  "test-repo",
		URL:   "https://github.com/test-org/test-repo",
	}
	_ = InstallSkill(skill, source, InstallOptions{}, nil)

	// Read back the installed skill with metadata
	skillDir, _ := GetSkillDir("tracked-skill")
	installedSkill, err := LoadSkillWithMeta(skillDir)
	if err != nil {
		t.Fatalf("LoadSkillWithMeta() error: %v", err)
	}

	// Verify metadata
	if installedSkill.Meta == nil {
		t.Fatal("Skill should have metadata")
	}
	if installedSkill.Meta.Source != "test-org/test-repo" {
		t.Errorf("Meta.Source = %q, want 'test-org/test-repo'", installedSkill.Meta.Source)
	}
	if installedSkill.Meta.SourceType != "github" {
		t.Errorf("Meta.SourceType = %q, want 'github'", installedSkill.Meta.SourceType)
	}
	if installedSkill.Meta.ContentHash == "" {
		t.Error("Meta.ContentHash should not be empty")
	}

	// Simulate source update
	updatedContent := "---\nname: tracked-skill\ndescription: A tracked skill\n---\n# Version 2 - Updated!\n"

	// Check if skill needs update
	needsUpdate, _ := SkillNeedsUpdate(skillDir, updatedContent)
	if !needsUpdate {
		t.Error("Skill should need update when content changes")
	}

	// Same content should not need update
	needsUpdate, _ = SkillNeedsUpdate(skillDir, originalContent)
	if needsUpdate {
		t.Error("Skill should not need update when content is the same")
	}
}

// TestIntegration_MultiAgentSync tests that skills are properly synced
// to multiple agents.
func TestIntegration_MultiAgentSync(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create multiple mock agent directories
	agents := []struct {
		id        string
		skillsDir string
	}{
		{"claude-code", filepath.Join(tmpDir, ".claude", "skills")},
		{"cursor", filepath.Join(tmpDir, ".cursor", "skills")},
		{"windsurf", filepath.Join(tmpDir, ".codeium", "windsurf", "skills")},
	}

	for _, agent := range agents {
		_ = os.MkdirAll(agent.skillsDir, 0o755)
	}

	// Create and install a skill
	sourceDir := filepath.Join(tmpDir, "source", "multi-agent-skill")
	_ = os.MkdirAll(sourceDir, 0o755)
	content := "---\nname: multi-agent-skill\ndescription: Installed to all agents\n---\n# Multi Agent\n"
	_ = os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(content), 0o644)

	skill, _ := ParseSkillMd(filepath.Join(sourceDir, "SKILL.md"))
	source := &SourceInfo{Type: "local", LocalPath: sourceDir}
	_ = InstallSkill(skill, source, InstallOptions{}, nil)

	// Verify symlink exists in all agent directories
	for _, agent := range agents {
		linkPath := filepath.Join(agent.skillsDir, "multi-agent-skill")
		if !IsSymlink(linkPath) {
			t.Errorf("Skill should be symlinked to %s", agent.id)
		}

		// Verify symlink points to correct location
		target, err := GetSymlinkTarget(linkPath)
		if err != nil {
			t.Errorf("Failed to get symlink target for %s: %v", agent.id, err)
		}
		scrollsDir, _ := GetScrollsDir()
		expectedTarget := filepath.Join(scrollsDir, "multi-agent-skill")
		// Resolve to absolute path for comparison
		absTarget, _ := filepath.Abs(filepath.Join(filepath.Dir(linkPath), target))
		if absTarget != expectedTarget {
			t.Errorf("Symlink for %s points to %s, want %s", agent.id, absTarget, expectedTarget)
		}
	}

	// Uninstall and verify removal from all agents
	_ = RemoveSkillFromAllWorkspaces("multi-agent-skill")
	_ = UninstallSkill("multi-agent-skill")

	for _, agent := range agents {
		linkPath := filepath.Join(agent.skillsDir, "multi-agent-skill")
		if IsSymlink(linkPath) {
			t.Errorf("Skill symlink should be removed from %s", agent.id)
		}
	}
}

// TestIntegration_WorkspaceSkillAddRemove tests adding and removing skills
// from specific workspaces.
func TestIntegration_WorkspaceSkillAddRemove(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create agent directory
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	_ = os.MkdirAll(claudeSkillsDir, 0o755)

	// Install a skill
	sourceDir := filepath.Join(tmpDir, "source", "workspace-test-skill")
	_ = os.MkdirAll(sourceDir, 0o755)
	content := "---\nname: workspace-test-skill\ndescription: Test skill\n---\n# Test\n"
	_ = os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(content), 0o644)

	skill, _ := ParseSkillMd(filepath.Join(sourceDir, "SKILL.md"))
	source := &SourceInfo{Type: "local", LocalPath: sourceDir}
	_ = InstallSkill(skill, source, InstallOptions{}, nil)
	_ = AddSkillToActiveAndDefaultWorkspace("workspace-test-skill")

	// Create a custom workspace without the skill
	_ = CreateWorkspace(&Workspace{
		Name:        "custom",
		Description: "Custom workspace",
		Skills:      []string{},
	})

	// Add skill to custom workspace
	_ = AddSkillToWorkspace("workspace-test-skill", "custom")

	// Verify skill is in custom workspace
	customWs, _ := GetWorkspace("custom")
	found := false
	for _, s := range customWs.Skills {
		if s == "workspace-test-skill" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Skill should be in custom workspace after AddSkillToWorkspace")
	}

	// Remove skill from custom workspace
	_ = RemoveSkillFromWorkspace("workspace-test-skill", "custom")

	// Verify skill is removed from custom workspace
	customWs, _ = GetWorkspace("custom")
	for _, s := range customWs.Skills {
		if s == "workspace-test-skill" {
			t.Error("Skill should not be in custom workspace after RemoveSkillFromWorkspace")
		}
	}

	// Verify skill still exists in default workspace
	defaultWs, _ := GetWorkspace("default")
	found = false
	for _, s := range defaultWs.Skills {
		if s == "workspace-test-skill" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Skill should still be in default workspace")
	}
}

// TestIntegration_CleanupOrphanedWorkspaces tests that workspace cleanup
// properly handles orphaned skills.
func TestIntegration_CleanupOrphanedWorkspaces(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create a workspace with a skill that doesn't exist
	_ = CreateWorkspace(&Workspace{
		Name:        "orphaned",
		Description: "Has orphaned skills",
		Skills:      []string{"nonexistent-skill-1", "nonexistent-skill-2"},
	})

	// Also add orphaned skills to default
	defaultWs, _ := GetWorkspace("default")
	defaultWs.Skills = append(defaultWs.Skills, "nonexistent-skill-3")
	_ = UpdateWorkspace(defaultWs)

	// Run cleanup
	_ = CleanWorkspaces()

	// Verify orphaned skills are removed
	orphanedWs, _ := GetWorkspace("orphaned")
	if len(orphanedWs.Skills) != 0 {
		t.Errorf("Orphaned workspace should have 0 skills after cleanup, got %d", len(orphanedWs.Skills))
	}

	defaultWs, _ = GetWorkspace("default")
	for _, s := range defaultWs.Skills {
		if s == "nonexistent-skill-3" {
			t.Error("Orphaned skill should be removed from default workspace")
		}
	}
}

// TestIntegration_RebuildDefaultWorkspace tests rebuilding the default workspace
// to include all installed skills.
func TestIntegration_RebuildDefaultWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Install skills without adding to workspaces
	skillNames := []string{"rebuild-skill-1", "rebuild-skill-2", "rebuild-skill-3"}
	for _, name := range skillNames {
		sourceDir := filepath.Join(tmpDir, "source", name)
		_ = os.MkdirAll(sourceDir, 0o755)
		content := fmt.Sprintf("---\nname: %s\ndescription: Rebuild test\n---\n# Test\n", name)
		_ = os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(content), 0o644)

		skill, _ := ParseSkillMd(filepath.Join(sourceDir, "SKILL.md"))
		source := &SourceInfo{Type: "local", LocalPath: sourceDir}
		_ = InstallSkill(skill, source, InstallOptions{}, nil)
		// Intentionally NOT adding to workspace
	}

	// Default workspace should be empty or not have our skills
	defaultWs, _ := GetWorkspace("default")
	originalCount := len(defaultWs.Skills)

	// Rebuild default workspace
	_ = RebuildDefaultWorkspace()

	// Verify all installed skills are now in default workspace
	defaultWs, _ = GetWorkspace("default")
	if len(defaultWs.Skills) != originalCount+3 {
		t.Errorf("Default workspace should have %d skills after rebuild, got %d",
			originalCount+3, len(defaultWs.Skills))
	}

	for _, name := range skillNames {
		found := false
		for _, s := range defaultWs.Skills {
			if s == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Skill %s should be in default workspace after rebuild", name)
		}
	}
}
