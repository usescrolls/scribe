package scribe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// InstallSkill (installer.go)
// ============================================================================

func TestBoost_InstallSkill_Success(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create a source skill
	srcDir := filepath.Join(tmpDir, "src-skill")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("---\nname: install-test\ndescription: Install test\n---\n# Install\n"), 0o644)

	skill := &Skill{
		Name:        "install-test",
		Description: "Install test",
		Path:        srcDir,
	}
	source := &SourceInfo{
		Type:  "github",
		Owner: "testuser",
		Repo:  "testrepo",
		URL:   "https://github.com/testuser/testrepo",
	}

	err := InstallSkill(skill, source, InstallOptions{}, nil)
	if err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	// Verify installation
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "install-test")
	if _, err := os.Stat(filepath.Join(skillDir, "SKILL.md")); err != nil {
		t.Error("SKILL.md not found in installed location")
	}
	if _, err := os.Stat(filepath.Join(skillDir, ".scribe-meta.json")); err != nil {
		t.Error("metadata not created during install")
	}
}

func TestBoost_InstallSkill_AlreadyExists(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create an existing skill in scrolls
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls", "existing-skill")
	_ = os.MkdirAll(scrollsDir, 0o755)
	_ = os.WriteFile(filepath.Join(scrollsDir, "SKILL.md"), []byte("# Existing"), 0o644)

	skill := &Skill{
		Name:        "existing-skill",
		Description: "Already exists",
		Path:        "/tmp/whatever",
	}
	source := &SourceInfo{Type: "github", Owner: "u", Repo: "r"}

	err := InstallSkill(skill, source, InstallOptions{}, nil)
	if err == nil {
		t.Error("expected error when skill already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want 'already exists'", err.Error())
	}
}

func TestBoost_InstallSkill_WithSubpath(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	srcDir := filepath.Join(tmpDir, "sub-src")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("---\nname: subpath-install\ndescription: Subpath install\n---\n# Subpath\n"), 0o644)

	skill := &Skill{
		Name:        "subpath-install",
		Description: "Subpath install",
		Path:        srcDir,
	}
	source := &SourceInfo{
		Type:    "github",
		Owner:   "u",
		Repo:    "r",
		URL:     "https://github.com/u/r",
		Subpath: "skills/subpath-install",
	}

	err := InstallSkill(skill, source, InstallOptions{}, nil)
	if err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	// Verify meta has subpath
	metaPath := filepath.Join(tmpDir, ".scribe", "scrolls", "subpath-install", ".scribe-meta.json")
	meta, err := ReadSkillMeta(metaPath)
	if err != nil {
		t.Fatalf("ReadSkillMeta error: %v", err)
	}
	if meta.SkillPath != "skills/subpath-install" {
		t.Errorf("meta.SkillPath = %q, want 'skills/subpath-install'", meta.SkillPath)
	}
}

func TestBoost_InstallSkill_WithSpecificAgents(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create agent dir
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)

	srcDir := filepath.Join(tmpDir, "agent-src")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("---\nname: agent-skill\ndescription: Agent skill\n---\n# Agent\n"), 0o644)

	skill := &Skill{
		Name:        "agent-skill",
		Description: "Agent skill",
		Path:        srcDir,
	}
	source := &SourceInfo{Type: "local", LocalPath: srcDir}

	err := InstallSkill(skill, source, InstallOptions{}, nil)
	if err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	// Verify the skill was synced to claude-code
	agentSkillDir := filepath.Join(tmpDir, ".claude", "skills", "agent-skill")
	if _, err := os.Stat(agentSkillDir); err != nil {
		t.Error("skill not synced to specified agent")
	}
}

func TestBoost_InstallSkill_IsPrivate(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	srcDir := filepath.Join(tmpDir, "private-src")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("---\nname: private-skill\ndescription: Private skill\n---\n# Private\n"), 0o644)

	skill := &Skill{
		Name:        "private-skill",
		Description: "Private skill",
		Path:        srcDir,
	}
	source := &SourceInfo{
		Type:  "github",
		Owner: "user",
		Repo:  "private-repo",
		URL:   "git@github.com:user/private-repo.git",
	}

	err := InstallSkill(skill, source, InstallOptions{IsPrivate: true}, nil)
	if err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	// Verify IsPrivate was persisted in metadata
	metaPath := filepath.Join(tmpDir, ".scribe", "scrolls", "private-skill", ".scribe-meta.json")
	meta, err := ReadSkillMeta(metaPath)
	if err != nil {
		t.Fatalf("ReadSkillMeta error: %v", err)
	}
	if !meta.IsPrivate {
		t.Error("meta.IsPrivate = false, want true")
	}
}

func TestBoost_InstallSkill_PublicNotPrivate(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	srcDir := filepath.Join(tmpDir, "public-src")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("---\nname: public-skill\ndescription: Public skill\n---\n# Public\n"), 0o644)

	skill := &Skill{
		Name:        "public-skill",
		Description: "Public skill",
		Path:        srcDir,
	}
	source := &SourceInfo{
		Type:  "github",
		Owner: "user",
		Repo:  "public-repo",
		URL:   "https://github.com/user/public-repo",
	}

	err := InstallSkill(skill, source, InstallOptions{}, nil)
	if err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	// Verify IsPrivate defaults to false
	metaPath := filepath.Join(tmpDir, ".scribe", "scrolls", "public-skill", ".scribe-meta.json")
	meta, err := ReadSkillMeta(metaPath)
	if err != nil {
		t.Fatalf("ReadSkillMeta error: %v", err)
	}
	if meta.IsPrivate {
		t.Error("meta.IsPrivate = true, want false for public install")
	}
}

// ============================================================================
// SyncSkillToAgents (installer.go)
// ============================================================================

func TestBoost_SyncSkillToAgents_UnknownAgent(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create a skill in scrolls
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls", "sync-test")
	_ = os.MkdirAll(scrollsDir, 0o755)
	_ = os.WriteFile(filepath.Join(scrollsDir, "SKILL.md"), []byte("# Test"), 0o644)

	// Syncing to unknown agent should silently skip
	err := SyncSkillToAgents("sync-test", []string{"unknown-agent-xyz"})
	if err != nil {
		t.Errorf("SyncSkillToAgents with unknown agent error: %v", err)
	}
}

func TestBoost_SyncSkillToAgents_EmptyAgents(t *testing.T) {
	err := SyncSkillToAgents("any-skill", []string{})
	if err != nil {
		t.Errorf("SyncSkillToAgents with empty agents error: %v", err)
	}
}

func TestBoost_SyncSkillToAgents_MultipleAgents(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create skill in scrolls
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls", "multi-sync")
	_ = os.MkdirAll(scrollsDir, 0o755)
	_ = os.WriteFile(filepath.Join(scrollsDir, "SKILL.md"), []byte("# Multi"), 0o644)

	// Create multiple agent dirs
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, ".cursor"), 0o755)

	err := SyncSkillToAgents("multi-sync", []string{"claude-code", "cursor"})
	if err != nil {
		t.Fatalf("SyncSkillToAgents error: %v", err)
	}

	// Verify synced to both agents
	claudeSkill := filepath.Join(tmpDir, ".claude", "skills", "multi-sync")
	cursorSkill := filepath.Join(tmpDir, ".cursor", "skills", "multi-sync")

	if _, err := os.Stat(claudeSkill); err != nil {
		t.Error("skill not synced to claude-code")
	}
	if _, err := os.Stat(cursorSkill); err != nil {
		t.Error("skill not synced to cursor")
	}
}

func TestBoost_SyncSkillToAgents_FallbackToCopy(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create skill in scrolls
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "fallback-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: fallback-skill\ndescription: F\n---\n# F\n"), 0o644)

	// Create agent skills dir
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude", "skills"), 0o755)

	err := SyncSkillToAgents("fallback-skill", []string{"claude-code"})
	if err != nil {
		t.Fatalf("SyncSkillToAgents error: %v", err)
	}
}

// ============================================================================
// CreateSymlink (installer.go)
// ============================================================================

func TestBoost_CreateSymlink_Basic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-symlink-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	targetDir := filepath.Join(tmpDir, "target")
	_ = os.MkdirAll(targetDir, 0o755)
	_ = os.WriteFile(filepath.Join(targetDir, "file.txt"), []byte("hello"), 0o644)

	linkPath := filepath.Join(tmpDir, "link")
	err = CreateSymlink(targetDir, linkPath)
	if err != nil {
		t.Fatalf("CreateSymlink() error: %v", err)
	}

	// Verify the symlink works
	if !IsSymlink(linkPath) {
		t.Error("link is not a symlink")
	}

	// Verify we can read through the symlink
	data, err := os.ReadFile(filepath.Join(linkPath, "file.txt"))
	if err != nil {
		t.Fatalf("read through symlink: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("content = %q, want 'hello'", string(data))
	}
}

func TestBoost_CreateSymlink_RelativePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-symlink-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	targetDir := filepath.Join(tmpDir, "sub", "target")
	_ = os.MkdirAll(targetDir, 0o755)

	linkPath := filepath.Join(tmpDir, "sub", "link")
	err = CreateSymlink(targetDir, linkPath)
	if err != nil {
		t.Fatalf("CreateSymlink() error: %v", err)
	}

	// Verify the link target is relative
	target, err := GetSymlinkTarget(linkPath)
	if err != nil {
		t.Fatalf("GetSymlinkTarget error: %v", err)
	}
	if filepath.IsAbs(target) {
		t.Errorf("expected relative symlink, got absolute: %q", target)
	}
}

// ============================================================================
// IsSymlink / GetSymlinkTarget (installer.go)
// ============================================================================

func TestBoost_IsSymlink_NotASymlink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	regularFile := filepath.Join(tmpDir, "regular.txt")
	_ = os.WriteFile(regularFile, []byte("data"), 0o644)

	if IsSymlink(regularFile) {
		t.Error("regular file should not be a symlink")
	}
}

func TestBoost_IsSymlink_NonExistent(t *testing.T) {
	if IsSymlink("/tmp/nonexistent-path-12345") {
		t.Error("nonexistent path should not be a symlink")
	}
}

func TestBoost_GetSymlinkTarget_NotASymlink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	regularFile := filepath.Join(tmpDir, "regular.txt")
	_ = os.WriteFile(regularFile, []byte("data"), 0o644)

	_, err = GetSymlinkTarget(regularFile)
	if err == nil {
		t.Error("expected error for non-symlink")
	}
}

// ============================================================================
// UninstallSkill (installer.go)
// ============================================================================

func TestBoost_UninstallSkill(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create skill in scrolls
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "uninstall-me")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Uninstall"), 0o644)

	err := UninstallSkill("uninstall-me")
	if err != nil {
		t.Fatalf("UninstallSkill() error: %v", err)
	}

	// Verify removed
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Error("skill directory still exists after uninstall")
	}
}

// ============================================================================
// RemoveSkillFromAgents (installer.go)
// ============================================================================

func TestBoost_RemoveSkillFromAgents(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)

	// Create agent skill directory
	agentSkillDir := filepath.Join(tmpDir, ".claude", "skills", "remove-agent-skill")
	_ = os.MkdirAll(agentSkillDir, 0o755)

	err := RemoveSkillFromAgents("remove-agent-skill", []string{"claude-code"})
	if err != nil {
		t.Fatalf("RemoveSkillFromAgents() error: %v", err)
	}

	// Verify removed
	if _, err := os.Stat(agentSkillDir); !os.IsNotExist(err) {
		t.Error("agent skill directory still exists after removal")
	}
}

// ============================================================================
// SyncAllSkillsToAgents (installer.go)
// ============================================================================

func TestBoost_SyncAllSkillsToAgents(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create some skills
	for _, name := range []string{"sync-all-a", "sync-all-b"} {
		skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: test\n---\n# Test\n"), 0o644)
	}

	// Create agent dir
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)

	err := SyncAllSkillsToAgents()
	if err != nil {
		t.Fatalf("SyncAllSkillsToAgents() error: %v", err)
	}

	// Verify skills were synced
	for _, name := range []string{"sync-all-a", "sync-all-b"} {
		path := filepath.Join(tmpDir, ".claude", "skills", name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("skill %q not synced to agent", name)
		}
	}
}

// ============================================================================
// copyFile / CopySkillDir (installer.go)
// ============================================================================

func TestBoost_CopySkillDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copy-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcDir := filepath.Join(tmpDir, "src")
	_ = os.MkdirAll(filepath.Join(srcDir, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("# Root"), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "sub", "nested.txt"), []byte("nested"), 0o644)

	dstDir := filepath.Join(tmpDir, "dst")
	err = CopySkillDir(srcDir, dstDir)
	if err != nil {
		t.Fatalf("CopySkillDir() error: %v", err)
	}

	// Verify root file
	data, err := os.ReadFile(filepath.Join(dstDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("read root file: %v", err)
	}
	if string(data) != "# Root" {
		t.Errorf("root file content = %q, want '# Root'", string(data))
	}

	// Verify nested file
	data, err = os.ReadFile(filepath.Join(dstDir, "sub", "nested.txt"))
	if err != nil {
		t.Fatalf("read nested file: %v", err)
	}
	if string(data) != "nested" {
		t.Errorf("nested file content = %q, want 'nested'", string(data))
	}
}

func TestBoost_CopyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copy-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcPath := filepath.Join(tmpDir, "src.txt")
	_ = os.WriteFile(srcPath, []byte("file content"), 0o644)

	dstPath := filepath.Join(tmpDir, "dst.txt")
	err = copyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("copyFile() error: %v", err)
	}

	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != "file content" {
		t.Errorf("content = %q, want 'file content'", string(data))
	}
}

func TestBoost_CopyFile_CreateParentDir(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "scribe-copy-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	src := filepath.Join(tmpDir, "src.txt")
	_ = os.WriteFile(src, []byte("data"), 0o644)

	dst := filepath.Join(tmpDir, "a", "b", "c", "dst.txt")
	err := copyFile(src, dst)
	if err != nil {
		t.Fatalf("copyFile with nested dir error: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "data" {
		t.Errorf("content = %q, want 'data'", string(data))
	}
}
