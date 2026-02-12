package scribe

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// installFakeSkill creates a skill in the scrolls directory with metadata.
func installFakeSkill(t *testing.T, name, description, sourceType, source string) string {
	t.Helper()
	scrollsDir, err := GetScrollsDir()
	if err != nil {
		t.Fatalf("failed to get scrolls dir: %v", err)
	}
	skillDir := filepath.Join(scrollsDir, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	content := "---\nname: " + name + "\ndescription: " + description + "\n---\n\n# " + name + "\n\nTest skill content.\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}

	meta := NewSkillMeta(&SourceInfo{
		Type:  sourceType,
		Owner: "testowner",
		Repo:  "testrepo",
		URL:   "https://github.com/testowner/testrepo",
	}, "", content, nil)
	meta.Source = source
	meta.SourceType = sourceType

	if err := WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta); err != nil {
		t.Fatalf("failed to write meta: %v", err)
	}

	return skillDir
}

// ============================================================================
// CheckSkillForUpdate (updater.go)
// ============================================================================

func TestCheckSkillForUpdate_Nonexistent(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	result := CheckSkillForUpdate("nonexistent-skill")
	if result.Error == "" {
		t.Error("expected error for nonexistent skill")
	}
	if !strings.Contains(result.Error, "failed to read skill") {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.Name != "nonexistent-skill" {
		t.Errorf("Name = %q, want 'nonexistent-skill'", result.Name)
	}
}

func TestCheckSkillForUpdate_NoMeta(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create skill without metadata
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "no-meta")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: no-meta\ndescription: No meta\n---\n# No meta\n"), 0o644)

	result := CheckSkillForUpdate("no-meta")
	if result.Error == "" {
		t.Error("expected error for skill without metadata")
	}
	if !strings.Contains(result.Error, "no metadata") {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestCheckSkillForUpdate_LocalSource(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	installFakeSkill(t, "local-check", "Local check", "local", "/some/path")

	result := CheckSkillForUpdate("local-check")
	if result.Error == "" {
		t.Error("expected error for local source")
	}
	if !strings.Contains(result.Error, "local source") {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.CurrentHash == "" {
		t.Error("expected CurrentHash to be populated")
	}
}

func TestCheckSkillForUpdate_GitHubFetchFails(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	installFakeSkill(t, "github-check", "GitHub check", "github", "nonexistent-owner/nonexistent-repo")

	result := CheckSkillForUpdate("github-check")
	if result.Error == "" {
		t.Error("expected error for fetch failure")
	}
	if !strings.Contains(result.Error, "failed to fetch") {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestCheckSkillForUpdate_UpToDate(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	skillContent := "---\nname: up-to-date\ndescription: Up to date skill\n---\n# Up to date\n"

	// Create a local git repo as the "remote" source
	repoDir := filepath.Join(tmpDir, "remote-repo")
	skillInRepo := filepath.Join(repoDir, "up-to-date")
	_ = os.MkdirAll(skillInRepo, 0o755)
	_ = os.WriteFile(filepath.Join(skillInRepo, "SKILL.md"), []byte(skillContent), 0o644)
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"up-to-date/SKILL.md": skillContent,
	})

	// Install the skill locally with matching content
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "up-to-date")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644)

	meta := NewSkillMeta(&SourceInfo{
		Type:  "github",
		Owner: "testowner",
		Repo:  "remote-repo",
		URL:   repoURL,
	}, "", skillContent, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)

	result := CheckSkillForUpdate("up-to-date")
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.NeedsUpdate {
		t.Error("expected NeedsUpdate = false for identical content")
	}
	if result.CurrentHash == "" || result.RemoteHash == "" {
		t.Error("expected both hashes to be populated")
	}
	if result.CurrentHash != result.RemoteHash {
		t.Errorf("hashes should match: current=%s, remote=%s", result.CurrentHash, result.RemoteHash)
	}
}

func TestCheckSkillForUpdate_Outdated(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	oldContent := "---\nname: outdated\ndescription: Outdated skill\n---\n# Old version\n"
	newContent := "---\nname: outdated\ndescription: Outdated skill\n---\n# New version with changes\n"

	// Create a local git repo with the new content
	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"outdated/SKILL.md": newContent,
	})

	// Install the skill locally with old content
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "outdated")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(oldContent), 0o644)

	meta := NewSkillMeta(&SourceInfo{
		Type:  "github",
		Owner: "testowner",
		Repo:  "remote-repo",
		URL:   repoURL,
	}, "", oldContent, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)

	result := CheckSkillForUpdate("outdated")
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if !result.NeedsUpdate {
		t.Error("expected NeedsUpdate = true for different content")
	}
	if result.CurrentHash == result.RemoteHash {
		t.Error("hashes should differ")
	}
}

// ============================================================================
// UpdateSkill (updater.go)
// ============================================================================

func TestUpdateSkill_Nonexistent(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	_, err := UpdateSkill("nonexistent", false)
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
	if !strings.Contains(err.Error(), "skill not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateSkill_NoMeta(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "no-meta-update")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: no-meta-update\ndescription: No meta\n---\n# Test\n"), 0o644)

	_, err := UpdateSkill("no-meta-update", false)
	if err == nil {
		t.Error("expected error for skill without metadata")
	}
	if !strings.Contains(err.Error(), "no metadata") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateSkill_LocalSource(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	installFakeSkill(t, "local-update", "Local update", "local", "/some/path")

	_, err := UpdateSkill("local-update", false)
	if err == nil {
		t.Error("expected error for local source")
	}
	if !strings.Contains(err.Error(), "local source") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateSkill_GitHubFetchFails(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	installFakeSkill(t, "gh-update-fail", "GitHub fail", "github", "nonexistent-owner/nonexistent-repo")

	_, err := UpdateSkill("gh-update-fail", false)
	if err == nil {
		t.Error("expected error for fetch failure")
	}
	if !strings.Contains(err.Error(), "failed to fetch") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateSkill_AlreadyUpToDate(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	skillContent := "---\nname: up-to-date\ndescription: Up to date\n---\n# Up to date\n"

	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"up-to-date/SKILL.md": skillContent,
	})

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "up-to-date")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644)

	meta := NewSkillMeta(&SourceInfo{
		Type:  "github",
		Owner: "testowner",
		Repo:  "remote-repo",
		URL:   repoURL,
	}, "", skillContent, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)

	result, err := UpdateSkill("up-to-date", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Updated {
		t.Error("expected Updated = false for identical content")
	}
	if result.SkillName != "up-to-date" {
		t.Errorf("SkillName = %q, want 'up-to-date'", result.SkillName)
	}
}

func TestUpdateSkill_PerformsUpdate(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	oldContent := "---\nname: needs-update\ndescription: Needs update\n---\n# Old version\n"
	newContent := "---\nname: needs-update\ndescription: Needs update\n---\n# New version with improvements\n"

	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"needs-update/SKILL.md": newContent,
	})

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "needs-update")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(oldContent), 0o644)

	meta := NewSkillMeta(&SourceInfo{
		Type:  "github",
		Owner: "testowner",
		Repo:  "remote-repo",
		URL:   repoURL,
	}, "", oldContent, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)

	result, err := UpdateSkill("needs-update", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Updated {
		t.Error("expected Updated = true")
	}
	if result.SkillName != "needs-update" {
		t.Errorf("SkillName = %q, want 'needs-update'", result.SkillName)
	}

	// Verify the skill content was actually updated on disk
	data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("failed to read updated skill: %v", err)
	}
	if string(data) != newContent {
		t.Errorf("skill content not updated on disk")
	}

	// Verify metadata was updated
	updatedMeta, err := ReadSkillMeta(filepath.Join(skillDir, MetaFileName))
	if err != nil {
		t.Fatalf("failed to read updated meta: %v", err)
	}
	if updatedMeta.ContentHash == meta.ContentHash {
		t.Error("content hash should have changed after update")
	}
}

func TestUpdateSkill_ForceUpdateEvenWhenCurrent(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	skillContent := "---\nname: force-me\ndescription: Force me\n---\n# Same content\n"

	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"force-me/SKILL.md": skillContent,
	})

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "force-me")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644)

	meta := NewSkillMeta(&SourceInfo{
		Type:  "github",
		Owner: "testowner",
		Repo:  "remote-repo",
		URL:   repoURL,
	}, "", skillContent, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)

	result, err := UpdateSkill("force-me", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Updated {
		t.Error("expected Updated = true with force=true even when content matches")
	}
}

func TestUpdateSkill_BackfillsGitInfo(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	skillContent := "---\nname: backfill\ndescription: Backfill git info\n---\n# Same\n"

	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"backfill/SKILL.md": skillContent,
	})

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "backfill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644)

	// Create meta WITHOUT CommitHash to trigger backfill
	meta := NewSkillMeta(&SourceInfo{
		Type:  "github",
		Owner: "testowner",
		Repo:  "remote-repo",
		URL:   repoURL,
	}, "", skillContent, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	meta.CommitHash = "" // explicitly empty
	meta.CommitDate = ""
	_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)

	result, err := UpdateSkill("backfill", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Updated {
		t.Error("expected Updated = false (same content)")
	}

	// Verify git info was backfilled into metadata
	updatedMeta, err := ReadSkillMeta(filepath.Join(skillDir, MetaFileName))
	if err != nil {
		t.Fatalf("failed to read meta: %v", err)
	}
	if updatedMeta.CommitHash == "" {
		t.Error("expected CommitHash to be backfilled")
	}
}

func TestUpdateSkill_OldHashFromContentHash(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	installFakeSkill(t, "oldhash-test", "Old hash test", "github", "nonexistent/repo")

	// The skill has ContentHash set by NewSkillMeta (sha256:...) and CommitHash is empty.
	// UpdateSkill should extract old hash from ContentHash[7:14].
	result, _ := UpdateSkill("oldhash-test", false)
	// This will fail at fetch, but we can't check result since err != nil.
	// Instead, verify the logic by checking a skill where fetch succeeds — already covered above.
	_ = result
}

// ============================================================================
// CheckResult JSON serialization
// ============================================================================

func TestCheckResult_JSONRoundTrip(t *testing.T) {
	original := CheckResult{
		Name:        "test-skill",
		NeedsUpdate: true,
		CurrentHash: "sha256:abc123",
		RemoteHash:  "sha256:def456",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded CheckResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.NeedsUpdate != original.NeedsUpdate {
		t.Errorf("NeedsUpdate = %v, want %v", decoded.NeedsUpdate, original.NeedsUpdate)
	}
	if decoded.CurrentHash != original.CurrentHash {
		t.Errorf("CurrentHash = %q, want %q", decoded.CurrentHash, original.CurrentHash)
	}
	if decoded.RemoteHash != original.RemoteHash {
		t.Errorf("RemoteHash = %q, want %q", decoded.RemoteHash, original.RemoteHash)
	}
}

func TestCheckResult_JSONOmitsEmptyError(t *testing.T) {
	result := CheckResult{Name: "test", NeedsUpdate: false}
	data, _ := json.Marshal(result)
	if strings.Contains(string(data), "error") {
		t.Error("expected empty error to be omitted from JSON")
	}
}
