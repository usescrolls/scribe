package scribe

import (
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
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create skill directories
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skill1Dir := filepath.Join(scrollsDir, "skill-one")
	skill2Dir := filepath.Join(scrollsDir, "skill-two")

	os.MkdirAll(skill1Dir, 0755)
	os.MkdirAll(skill2Dir, 0755)

	// Create SKILL.md files
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Skill One"), 0644)
	os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Skill Two"), 0644)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

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
	os.MkdirAll(scrollsDir, 0755)
	os.WriteFile(filepath.Join(scrollsDir, "SKILL.md"), []byte("# Existing"), 0644)

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

	// Should have 45 agents
	if len(agents) != 45 {
		t.Errorf("GetAllAgents() returned %d agents, want 45", len(agents))
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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// No agents installed initially
	agents := DetectInstalledAgents()
	if len(agents) != 0 {
		t.Errorf("DetectInstalledAgents() = %d agents, want 0", len(agents))
	}

	// Create Claude config directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

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
		{"~/.claude/skills", filepath.Join(home, ".claude/skills")},
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
	defer os.RemoveAll(tmpDir)

	// Create a valid SKILL.md
	skillDir := filepath.Join(tmpDir, "test-skill")
	os.MkdirAll(skillDir, 0755)
	skillContent := `---
name: test-skill
description: A test skill for testing
---

# Test Skill

This is a test skill.
`
	skillPath := filepath.Join(skillDir, "SKILL.md")
	os.WriteFile(skillPath, []byte(skillContent), 0644)

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
	defer os.RemoveAll(tmpDir)

	skillDir := filepath.Join(tmpDir, "no-frontmatter")
	os.MkdirAll(skillDir, 0755)
	skillContent := `# No Frontmatter

This skill has no YAML frontmatter.
`
	skillPath := filepath.Join(skillDir, "SKILL.md")
	os.WriteFile(skillPath, []byte(skillContent), 0644)

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
	defer os.RemoveAll(tmpDir)

	// Create skill directories with SKILL.md files
	skill1Dir := filepath.Join(tmpDir, "skills", "skill-one")
	skill2Dir := filepath.Join(tmpDir, "skills", "skill-two")
	os.MkdirAll(skill1Dir, 0755)
	os.MkdirAll(skill2Dir, 0755)

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
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skill1Content), 0644)
	os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skill2Content), 0644)

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
	defer os.RemoveAll(tmpDir)

	metaPath := filepath.Join(tmpDir, ".scribe-meta.json")

	// Create meta
	source := &SourceInfo{
		Type:  "github",
		Owner: "user",
		Repo:  "repo",
	}
	meta := NewSkillMeta(source, "skills/test", "test content")

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Ensure directories exist
	EnsureScribeDirs()

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create workspace
	ws := &Workspace{
		Name:   "test-ws",
		Skills: []string{},
	}
	CreateWorkspace(ws)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create workspace with skills
	ws := &Workspace{
		Name:   "test-ws",
		Skills: []string{"skill-1", "skill-2", "skill-3"},
	}
	CreateWorkspace(ws)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create a source skill directory
	sourceDir := filepath.Join(tmpDir, "source", "test-skill")
	os.MkdirAll(sourceDir, 0755)
	skillContent := `---
name: test-skill
description: Test skill for installation
---
# Test Skill
`
	os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(skillContent), 0644)

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
	err = InstallSkill(skill, source, opts)
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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create a skill in scrolls
	skillDir, _ := GetSkillDir("sync-test")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Test"), 0644)

	// Create Claude agent directory
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	os.MkdirAll(claudeSkillsDir, 0755)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create a skill symlink in Claude directory
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	os.MkdirAll(claudeSkillsDir, 0755)
	linkPath := filepath.Join(claudeSkillsDir, "remove-test")
	os.MkdirAll(linkPath, 0755) // Create as directory for test

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create Claude agent directory (simulating Claude Code is installed)
	claudeDir := filepath.Join(tmpDir, ".claude")
	claudeSkillsDir := filepath.Join(claudeDir, "skills")
	os.MkdirAll(claudeSkillsDir, 0755)

	// Create source skill
	sourceDir := filepath.Join(tmpDir, "source", "my-skill")
	os.MkdirAll(sourceDir, 0755)
	skillContent := `---
name: my-skill
description: My awesome skill
---
# My Skill

This skill does cool things.
`
	os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(skillContent), 0644)

	// Parse and install
	skill, _ := ParseSkillMd(filepath.Join(sourceDir, "SKILL.md"))
	source := &SourceInfo{Type: "local", LocalPath: sourceDir}

	// Install to all detected agents
	detected := DetectInstalledAgents()
	agentIDs := make([]string, len(detected))
	for i, a := range detected {
		agentIDs[i] = a.ID
	}

	err = InstallSkill(skill, source, InstallOptions{Agents: agentIDs})
	if err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	// Add to workspace
	EnsureDefaultWorkspace()
	AddSkillToActiveAndDefaultWorkspace("my-skill")

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
	RemoveSkillFromAllWorkspaces("my-skill")
	UninstallSkill("my-skill")

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")

	// No agents installed
	statuses := GetAgentStatus(scrollsDir)
	if len(statuses) != 45 {
		t.Errorf("GetAgentStatus() returned %d statuses, want 45", len(statuses))
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
	os.MkdirAll(claudeSkillsDir, 0755)

	// Create a skill directory
	skill1Dir := filepath.Join(claudeSkillsDir, "test-skill")
	os.MkdirAll(skill1Dir, 0755)
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Test"), 0644)

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
	expected := filepath.Join(home, ".claude/skills")

	if result != expected {
		t.Errorf("ExpandAgentPath() = %q, want %q", result, expected)
	}
}

func TestCountSkillsInDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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

	os.MkdirAll(skill1Dir, 0755)
	os.MkdirAll(skill2Dir, 0755)
	os.MkdirAll(noSkillDir, 0755)

	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Skill 1"), 0644)
	os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Skill 2"), 0644)
	os.WriteFile(filepath.Join(noSkillDir, "README.md"), []byte("# Not a skill"), 0644)

	count = countSkillsInDir(tmpDir)
	if count != 2 {
		t.Errorf("countSkillsInDir() = %d, want 2", count)
	}
}

func TestIsSymlink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a regular file
	regularFile := filepath.Join(tmpDir, "regular.txt")
	os.WriteFile(regularFile, []byte("test"), 0644)

	if IsSymlink(regularFile) {
		t.Error("Regular file should not be a symlink")
	}

	// Create a directory
	dir := filepath.Join(tmpDir, "dir")
	os.MkdirAll(dir, 0755)

	if IsSymlink(dir) {
		t.Error("Directory should not be a symlink")
	}

	// Create a symlink
	symlinkPath := filepath.Join(tmpDir, "link")
	os.Symlink(regularFile, symlinkPath)

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
	defer os.RemoveAll(tmpDir)

	// Create target file
	targetFile := filepath.Join(tmpDir, "target.txt")
	os.WriteFile(targetFile, []byte("test"), 0644)

	// Create symlink
	symlinkPath := filepath.Join(tmpDir, "link")
	os.Symlink(targetFile, symlinkPath)

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
	meta := NewSkillMeta(source, "skills/test", "original content")

	originalHash := meta.ContentHash

	// Update with new content
	UpdateSkillMeta(meta, "new content")

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
	defer os.RemoveAll(tmpDir)

	skillDir := filepath.Join(tmpDir, "test-skill")
	os.MkdirAll(skillDir, 0755)

	// Create SKILL.md
	skillContent := `---
name: test-skill
description: A test skill
---
# Test Skill
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)

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
	meta := NewSkillMeta(source, "", skillContent)
	WriteSkillMeta(filepath.Join(skillDir, ".scribe-meta.json"), meta)

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
	defer os.RemoveAll(tmpDir)

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
	os.MkdirAll(scrollsDir, 0755)

	skill1Dir := filepath.Join(scrollsDir, "skill-one")
	skill2Dir := filepath.Join(scrollsDir, "skill-two")
	os.MkdirAll(skill1Dir, 0755)
	os.MkdirAll(skill2Dir, 0755)

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
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skill1Content), 0644)
	os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skill2Content), 0644)

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
	defer os.RemoveAll(tmpDir)

	skillDir := filepath.Join(tmpDir, "test-skill")
	os.MkdirAll(skillDir, 0755)

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
	meta := NewSkillMeta(source, "", content)
	WriteSkillMeta(filepath.Join(skillDir, ".scribe-meta.json"), meta)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create a test workspace
	ws := &Workspace{
		Name:        "test-workspace",
		Description: "A test workspace",
		Skills:      []string{"skill-1", "skill-2"},
	}
	CreateWorkspace(ws)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create skills
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skill1Dir := filepath.Join(scrollsDir, "skill-one")
	skill2Dir := filepath.Join(scrollsDir, "skill-two")
	os.MkdirAll(skill1Dir, 0755)
	os.MkdirAll(skill2Dir, 0755)

	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Skill 1"), 0644)
	os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Skill 2"), 0644)

	// Create agent directories
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	cursorSkillsDir := filepath.Join(tmpDir, ".cursor", "skills")
	os.MkdirAll(claudeSkillsDir, 0755)
	os.MkdirAll(cursorSkillsDir, 0755)

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
	defer os.RemoveAll(tmpDir)

	// Create target directory
	targetDir := filepath.Join(tmpDir, "target")
	os.MkdirAll(targetDir, 0755)
	os.WriteFile(filepath.Join(targetDir, "test.txt"), []byte("test"), 0644)

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
	defer os.RemoveAll(tmpDir)

	// Directory exists
	if !dirExists(tmpDir) {
		t.Error("dirExists() should return true for existing directory")
	}

	// File is not a directory
	filePath := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(filePath, []byte("test"), 0644)
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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create a skill with valid SKILL.md
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "test-skill")
	os.MkdirAll(skillDir, 0755)

	skillContent := `---
name: test-skill
description: A test skill
---
# Test Skill Content
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)

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
	meta := NewSkillMeta(source, "", skillContent)
	WriteSkillMeta(filepath.Join(skillDir, ".scribe-meta.json"), meta)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

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
	os.MkdirAll(skill1Dir, 0755)
	os.MkdirAll(skill2Dir, 0755)

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
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skillContent1), 0644)
	os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skillContent2), 0644)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create a skill
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "info-skill")
	os.MkdirAll(skillDir, 0755)

	skillContent := `---
name: info-skill
description: Skill for info test
---
# Info Skill
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)

	// Create Claude agent directory with the skill
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	os.MkdirAll(claudeSkillsDir, 0755)
	os.MkdirAll(filepath.Join(claudeSkillsDir, "info-skill"), 0755)
	os.WriteFile(filepath.Join(claudeSkillsDir, "info-skill", "SKILL.md"), []byte(skillContent), 0644)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()
	EnsureDefaultWorkspace()

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()
	EnsureDefaultWorkspace()

	// Create a new workspace
	ws := &Workspace{
		Name:        "new-workspace",
		Description: "A new workspace",
		Skills:      []string{},
	}
	CreateWorkspace(ws)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create skills
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skill1Dir := filepath.Join(scrollsDir, "sync-skill-1")
	skill2Dir := filepath.Join(scrollsDir, "sync-skill-2")
	os.MkdirAll(skill1Dir, 0755)
	os.MkdirAll(skill2Dir, 0755)
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Skill 1"), 0644)
	os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Skill 2"), 0644)

	// Create Claude directory
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	os.MkdirAll(claudeSkillsDir, 0755)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create some skills
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skill1Dir := filepath.Join(scrollsDir, "rebuild-skill-1")
	skill2Dir := filepath.Join(scrollsDir, "rebuild-skill-2")
	os.MkdirAll(skill1Dir, 0755)
	os.MkdirAll(skill2Dir, 0755)
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("# Skill 1"), 0644)
	os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Skill 2"), 0644)

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
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	EnsureScribeDirs()

	// Create a skill
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "existing-skill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Existing"), 0644)

	// Create workspace with existing and non-existing skills
	ws := &Workspace{
		Name:   "test-ws",
		Skills: []string{"existing-skill", "deleted-skill"},
	}
	CreateWorkspace(ws)

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
	defer os.RemoveAll(tmpDir)

	// Create source skill
	sourceDir := filepath.Join(tmpDir, "source")
	os.MkdirAll(sourceDir, 0755)

	skillContent := `---
name: save-skill
description: Skill for save test
---
# Save Skill Content
`
	os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(skillContent), 0644)

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
	err = SaveSkillWithMeta(targetDir, skill, source, "skills/save-skill")
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
		name        string
		url         string
		wantType    string
		wantOwner   string
		wantRepo    string
		wantSkill   string
		wantRef     string
		wantErr     bool
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
