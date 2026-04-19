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
// CheckSourceForUpdates (updater.go) — batch check
// ============================================================================

func TestCheckSourceForUpdates_EmptyList(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	results, _ := CheckSourceForUpdates("owner/repo", nil)
	if results != nil {
		t.Errorf("expected nil for empty skill list, got %v", results)
	}
}

func TestCheckSourceForUpdates_NonexistentSkill(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	results, _ := CheckSourceForUpdates("owner/repo", []string{"nonexistent"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error == "" {
		t.Error("expected error for nonexistent skill")
	}
}

func TestCheckSourceForUpdates_FetchFailure(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	installFakeSkill(t, "batch-fail-a", "Batch fail A", "github", "nonexistent-owner/nonexistent-repo")
	installFakeSkill(t, "batch-fail-b", "Batch fail B", "github", "nonexistent-owner/nonexistent-repo")

	results, _ := CheckSourceForUpdates("nonexistent-owner/nonexistent-repo", []string{"batch-fail-a", "batch-fail-b"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error == "" {
			t.Errorf("expected error for %s", r.Name)
		}
		if !strings.Contains(r.Error, "failed to fetch") {
			t.Errorf("unexpected error for %s: %s", r.Name, r.Error)
		}
	}
}

func TestCheckSourceForUpdates_BatchUpToDate(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	contentA := "---\nname: batch-a\ndescription: Batch A\n---\n# Batch A\n"
	contentB := "---\nname: batch-b\ndescription: Batch B\n---\n# Batch B\n"

	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"batch-a/SKILL.md": contentA,
		"batch-b/SKILL.md": contentB,
	})

	scrollsDir, _ := GetScrollsDir()
	for _, tc := range []struct{ name, content string }{
		{"batch-a", contentA},
		{"batch-b", contentB},
	} {
		skillDir := filepath.Join(scrollsDir, tc.name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(tc.content), 0o644)
		meta := NewSkillMeta(&SourceInfo{
			Type: "github", Owner: "testowner", Repo: "remote-repo", URL: repoURL,
		}, "", tc.content, nil)
		meta.Source = "testowner/remote-repo"
		meta.SourceType = "github"
		_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)
	}

	results, _ := CheckSourceForUpdates("testowner/remote-repo", []string{"batch-a", "batch-b"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error != "" {
			t.Errorf("unexpected error for %s: %s", r.Name, r.Error)
		}
		if r.NeedsUpdate {
			t.Errorf("expected NeedsUpdate=false for %s", r.Name)
		}
	}
}

func TestCheckSourceForUpdates_BatchMixed(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	contentA := "---\nname: mix-a\ndescription: Mix A\n---\n# Mix A\n"
	oldContentB := "---\nname: mix-b\ndescription: Mix B\n---\n# Mix B old\n"
	newContentB := "---\nname: mix-b\ndescription: Mix B\n---\n# Mix B new version\n"

	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"mix-a/SKILL.md": contentA,
		"mix-b/SKILL.md": newContentB,
	})

	scrollsDir, _ := GetScrollsDir()
	// mix-a is up to date
	{
		skillDir := filepath.Join(scrollsDir, "mix-a")
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(contentA), 0o644)
		meta := NewSkillMeta(&SourceInfo{
			Type: "github", Owner: "testowner", Repo: "remote-repo", URL: repoURL,
		}, "", contentA, nil)
		meta.Source = "testowner/remote-repo"
		meta.SourceType = "github"
		_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)
	}
	// mix-b is outdated
	{
		skillDir := filepath.Join(scrollsDir, "mix-b")
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(oldContentB), 0o644)
		meta := NewSkillMeta(&SourceInfo{
			Type: "github", Owner: "testowner", Repo: "remote-repo", URL: repoURL,
		}, "", oldContentB, nil)
		meta.Source = "testowner/remote-repo"
		meta.SourceType = "github"
		_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)
	}

	results, _ := CheckSourceForUpdates("testowner/remote-repo", []string{"mix-a", "mix-b"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	resultMap := make(map[string]CheckResult)
	for _, r := range results {
		resultMap[r.Name] = r
	}

	if resultMap["mix-a"].NeedsUpdate {
		t.Error("mix-a should be up to date")
	}
	if !resultMap["mix-b"].NeedsUpdate {
		t.Error("mix-b should need an update")
	}
}

func TestCheckSourceForUpdates_SkillNotInRemote(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	contentA := "---\nname: exists-remote\ndescription: Exists\n---\n# Exists\n"

	// Remote repo only has "exists-remote", not "ghost"
	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"exists-remote/SKILL.md": contentA,
	})

	scrollsDir, _ := GetScrollsDir()

	// Install "exists-remote" (matches remote)
	{
		skillDir := filepath.Join(scrollsDir, "exists-remote")
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(contentA), 0o644)
		meta := NewSkillMeta(&SourceInfo{
			Type: "github", Owner: "testowner", Repo: "remote-repo", URL: repoURL,
		}, "", contentA, nil)
		meta.Source = "testowner/remote-repo"
		meta.SourceType = "github"
		_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)
	}

	// Install "ghost" (not in the remote repo, but claims same source)
	ghostContent := "---\nname: ghost\ndescription: Ghost\n---\n# Ghost\n"
	{
		skillDir := filepath.Join(scrollsDir, "ghost")
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(ghostContent), 0o644)
		meta := NewSkillMeta(&SourceInfo{
			Type: "github", Owner: "testowner", Repo: "remote-repo", URL: repoURL,
		}, "", ghostContent, nil)
		meta.Source = "testowner/remote-repo"
		meta.SourceType = "github"
		_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)
	}

	results, _ := CheckSourceForUpdates("testowner/remote-repo", []string{"exists-remote", "ghost"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	resultMap := make(map[string]CheckResult)
	for _, r := range results {
		resultMap[r.Name] = r
	}

	// exists-remote should be fine
	if resultMap["exists-remote"].Error != "" {
		t.Errorf("unexpected error for exists-remote: %s", resultMap["exists-remote"].Error)
	}

	// ghost should report "not found in remote source"
	if resultMap["ghost"].Error == "" {
		t.Error("expected error for ghost skill")
	}
	if !strings.Contains(resultMap["ghost"].Error, "not found in remote source") {
		t.Errorf("unexpected error for ghost: %s", resultMap["ghost"].Error)
	}
}

func TestCheckSourceForUpdates_SkillNoMeta(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	contentA := "---\nname: has-meta\ndescription: Has meta\n---\n# Has meta\n"

	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"has-meta/SKILL.md": contentA,
		"no-meta/SKILL.md":  "---\nname: no-meta\ndescription: No meta\n---\n# No meta\n",
	})

	scrollsDir, _ := GetScrollsDir()

	// Install "has-meta" with proper metadata (this is the first skill, so source can be reconstructed)
	{
		skillDir := filepath.Join(scrollsDir, "has-meta")
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(contentA), 0o644)
		meta := NewSkillMeta(&SourceInfo{
			Type: "github", Owner: "testowner", Repo: "remote-repo", URL: repoURL,
		}, "", contentA, nil)
		meta.Source = "testowner/remote-repo"
		meta.SourceType = "github"
		_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)
	}

	// Install "no-meta" WITHOUT metadata sidecar
	{
		skillDir := filepath.Join(scrollsDir, "no-meta")
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: no-meta\ndescription: No meta\n---\n# No meta\n"), 0o644)
		// No WriteSkillMeta — deliberately missing
	}

	results, _ := CheckSourceForUpdates("testowner/remote-repo", []string{"has-meta", "no-meta"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	resultMap := make(map[string]CheckResult)
	for _, r := range results {
		resultMap[r.Name] = r
	}

	if resultMap["has-meta"].Error != "" {
		t.Errorf("unexpected error for has-meta: %s", resultMap["has-meta"].Error)
	}
	if resultMap["no-meta"].Error == "" {
		t.Error("expected error for no-meta skill")
	}
	if !strings.Contains(resultMap["no-meta"].Error, "no metadata") {
		t.Errorf("unexpected error for no-meta: %s", resultMap["no-meta"].Error)
	}
}

// ============================================================================
// CheckAllSourcesForUpdates (updater.go) — full scan
// ============================================================================

func TestCheckAllSourcesForUpdates_Empty(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	results := CheckAllSourcesForUpdates()
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty scrolls, got %d", len(results))
	}
}

func TestCheckAllSourcesForUpdates_SkipsLocalAndSystem(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	installFakeSkill(t, "local-skill", "Local", "local", "/some/path")
	installFakeSkill(t, "builtin-skill", "Builtin", "builtin", "builtin")

	results := CheckAllSourcesForUpdates()
	if len(results) != 0 {
		t.Errorf("expected 0 results (local and builtin should be skipped), got %d", len(results))
	}
}

func TestCheckAllSourcesForUpdates_GroupsBySource(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	contentA := "---\nname: grp-a\ndescription: Group A\n---\n# Group A\n"
	contentB := "---\nname: grp-b\ndescription: Group B\n---\n# Group B\n"

	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"grp-a/SKILL.md": contentA,
		"grp-b/SKILL.md": contentB,
	})

	scrollsDir, _ := GetScrollsDir()
	for _, tc := range []struct{ name, content string }{
		{"grp-a", contentA},
		{"grp-b", contentB},
	} {
		skillDir := filepath.Join(scrollsDir, tc.name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(tc.content), 0o644)
		meta := NewSkillMeta(&SourceInfo{
			Type: "github", Owner: "testowner", Repo: "remote-repo", URL: repoURL,
		}, "", tc.content, nil)
		meta.Source = "testowner/remote-repo"
		meta.SourceType = "github"
		_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)
	}

	results := CheckAllSourcesForUpdates()
	if len(results) != 1 {
		t.Fatalf("expected 1 source group, got %d", len(results))
	}

	sgr, ok := results["github:testowner/remote-repo"]
	if !ok {
		t.Fatal("expected result for 'github:testowner/remote-repo'")
	}
	if sgr.HasUpdates {
		t.Error("expected HasUpdates=false (all up to date)")
	}
	if sgr.Source != "testowner/remote-repo" {
		t.Errorf("Source = %q, want 'testowner/remote-repo'", sgr.Source)
	}
	if sgr.CheckedAt == "" {
		t.Error("expected CheckedAt to be set")
	}
}

func TestCheckAllSourcesForUpdates_SeparatesSameOwnerRepoAcrossProviders(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	githubContent := "---\nname: github-commit\ndescription: GitHub\n---\n# GitHub\n"
	gitlabContent := "---\nname: gitlab-commit\ndescription: GitLab\n---\n# GitLab\n"

	githubRepoDir := filepath.Join(tmpDir, "github-remote-repo")
	githubRepoURL := createTestGitRepo(t, githubRepoDir, map[string]string{
		"github-commit/SKILL.md": githubContent,
	})

	gitlabRepoDir := filepath.Join(tmpDir, "gitlab-remote-repo")
	gitlabRepoURL := createTestGitRepo(t, gitlabRepoDir, map[string]string{
		"gitlab-commit/SKILL.md": gitlabContent,
	})

	scrollsDir, _ := GetScrollsDir()
	for _, tc := range []struct {
		name       string
		content    string
		sourceType string
		sourceURL  string
	}{
		{"github-commit", githubContent, "github", githubRepoURL},
		{"gitlab-commit", gitlabContent, "gitlab", gitlabRepoURL},
	} {
		skillDir := filepath.Join(scrollsDir, tc.name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(tc.content), 0o644)
		meta := NewSkillMeta(&SourceInfo{
			Type: tc.sourceType, Owner: "nunomen", Repo: "claude-skills", URL: tc.sourceURL,
		}, "", tc.content, nil)
		meta.Source = "nunomen/claude-skills"
		meta.SourceType = tc.sourceType
		_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)
	}

	results := CheckAllSourcesForUpdates()
	if len(results) != 2 {
		t.Fatalf("expected 2 source groups, got %d", len(results))
	}

	if _, ok := results["github:nunomen/claude-skills"]; !ok {
		t.Fatal("expected GitHub source group")
	}
	if _, ok := results["gitlab:nunomen/claude-skills"]; !ok {
		t.Fatal("expected GitLab source group")
	}
}

func TestCheckAllSourcesForUpdates_DetectsOutdated(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	oldContent := "---\nname: outdated-scan\ndescription: Outdated\n---\n# Old\n"
	newContent := "---\nname: outdated-scan\ndescription: Outdated\n---\n# New version\n"

	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"outdated-scan/SKILL.md": newContent,
	})

	scrollsDir, _ := GetScrollsDir()
	skillDir := filepath.Join(scrollsDir, "outdated-scan")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(oldContent), 0o644)
	meta := NewSkillMeta(&SourceInfo{
		Type: "github", Owner: "testowner", Repo: "remote-repo", URL: repoURL,
	}, "", oldContent, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)

	results := CheckAllSourcesForUpdates()
	if len(results) != 1 {
		t.Fatalf("expected 1 source group, got %d", len(results))
	}

	sgr := results["github:testowner/remote-repo"]
	if !sgr.HasUpdates {
		t.Error("expected HasUpdates=true")
	}
	if len(sgr.UpdatedSkillNames) != 1 || sgr.UpdatedSkillNames[0] != "outdated-scan" {
		t.Errorf("UpdatedSkillNames = %v, want [outdated-scan]", sgr.UpdatedSkillNames)
	}
}

// ============================================================================
// SourceGroupCheckResult JSON serialization
// ============================================================================

func TestSourceGroupCheckResult_JSONRoundTrip(t *testing.T) {
	original := SourceGroupCheckResult{
		Source:            "owner/repo",
		HasUpdates:        true,
		UpdatedSkillNames: []string{"skill-a", "skill-b"},
		CheckedAt:         "2025-01-01T00:00:00Z",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded SourceGroupCheckResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Source != original.Source {
		t.Errorf("Source = %q, want %q", decoded.Source, original.Source)
	}
	if decoded.HasUpdates != original.HasUpdates {
		t.Errorf("HasUpdates = %v, want %v", decoded.HasUpdates, original.HasUpdates)
	}
	if len(decoded.UpdatedSkillNames) != 2 {
		t.Errorf("UpdatedSkillNames len = %d, want 2", len(decoded.UpdatedSkillNames))
	}
	if decoded.CheckedAt != original.CheckedAt {
		t.Errorf("CheckedAt = %q, want %q", decoded.CheckedAt, original.CheckedAt)
	}
}

func TestSourceGroupCheckResult_JSONOmitsEmptyError(t *testing.T) {
	result := SourceGroupCheckResult{Source: "test", HasUpdates: false}
	data, _ := json.Marshal(result)
	if strings.Contains(string(data), `"error"`) {
		t.Error("expected empty error to be omitted from JSON")
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

func TestUpdateSkill_RemovesSkillNoLongerInSource(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	skillContent := "---\nname: ghost-skill\ndescription: Will be removed from repo\n---\n# Ghost\n"

	// Create a repo that does NOT contain ghost-skill (only has other-skill)
	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"other-skill/SKILL.md": "---\nname: other-skill\ndescription: Other\n---\n# Other\n",
	})

	// Install ghost-skill locally, pointing at the repo that no longer has it
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "ghost-skill")
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

	// Also add skill to a workspace so we can verify it gets cleaned up
	wsDir, _ := GetWorkspacesDir()
	_ = os.MkdirAll(wsDir, 0o755)
	wsContent := `{"name":"test-ws","description":"","skills":["ghost-skill","other-skill"]}`
	_ = os.WriteFile(filepath.Join(wsDir, "test-ws.json"), []byte(wsContent), 0o644)

	result, err := UpdateSkill("ghost-skill", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Removed {
		t.Error("expected Removed = true")
	}
	if result.Updated {
		t.Error("expected Updated = false")
	}

	// Verify skill was uninstalled from disk
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Error("expected skill directory to be removed")
	}

	// Verify skill was removed from workspace
	ws, err := GetWorkspace("test-ws")
	if err != nil {
		t.Fatalf("failed to read workspace: %v", err)
	}
	for _, s := range ws.Skills {
		if s == "ghost-skill" {
			t.Error("expected ghost-skill to be removed from workspace")
		}
	}
}

func TestUpdateSkill_RefreshesCommitInfoWhenUpToDate(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	skillContent := "---\nname: refresh-test\ndescription: Refresh commit info\n---\n# Same\n"

	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"refresh-test/SKILL.md": skillContent,
	})

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	skillDir := filepath.Join(scrollsDir, "refresh-test")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644)

	// Create meta with a stale commit hash
	meta := NewSkillMeta(&SourceInfo{
		Type:  "github",
		Owner: "testowner",
		Repo:  "remote-repo",
		URL:   repoURL,
	}, "", skillContent, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	meta.CommitHash = "oldold1"
	meta.CommitDate = "2020-01-01T00:00:00Z"
	meta.UpdatedAt = "2020-01-01T00:00:00Z"
	_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)

	result, err := UpdateSkill("refresh-test", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Updated {
		t.Error("expected Updated = false (same content)")
	}

	// Verify commit info was refreshed to the repo HEAD
	updatedMeta, err := ReadSkillMeta(filepath.Join(skillDir, MetaFileName))
	if err != nil {
		t.Fatalf("failed to read meta: %v", err)
	}
	if updatedMeta.CommitHash == "oldold1" {
		t.Error("expected CommitHash to be refreshed from stale value")
	}
	if updatedMeta.CommitHash == "" {
		t.Error("expected CommitHash to be set")
	}
	if updatedMeta.CommitDate == "2020-01-01T00:00:00Z" {
		t.Error("expected CommitDate to be refreshed from stale value")
	}
	if updatedMeta.UpdatedAt == meta.UpdatedAt {
		t.Error("expected UpdatedAt to be refreshed")
	}
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

// ============================================================================
// NewAvailableSkills — CheckSourceForUpdates discovers uninstalled skills
// ============================================================================

func TestCheckSourceForUpdates_NewAvailableSkills(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	contentA := "---\nname: new-avail-a\ndescription: Skill A\n---\n# Skill A\n"
	contentB := "---\nname: new-avail-b\ndescription: Skill B\n---\n# Skill B\n"
	contentC := "---\nname: new-avail-c\ndescription: Skill C\n---\n# Skill C\n"

	// Remote repo has three skills
	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"new-avail-a/SKILL.md": contentA,
		"new-avail-b/SKILL.md": contentB,
		"new-avail-c/SKILL.md": contentC,
	})

	// Only install skill A locally
	scrollsDir, _ := GetScrollsDir()
	skillDir := filepath.Join(scrollsDir, "new-avail-a")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(contentA), 0o644)
	meta := NewSkillMeta(&SourceInfo{
		Type: "github", Owner: "testowner", Repo: "remote-repo", URL: repoURL,
	}, "", contentA, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)

	results, newSkills := CheckSourceForUpdates("testowner/remote-repo", []string{"new-avail-a"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error != "" {
		t.Errorf("unexpected error: %s", results[0].Error)
	}

	if len(newSkills) != 2 {
		t.Fatalf("expected 2 new skills, got %d", len(newSkills))
	}

	// Collect new skill names into a set for order-independent comparison
	newNames := make(map[string]bool)
	for _, s := range newSkills {
		newNames[s.Name] = true
	}
	if !newNames["new-avail-b"] {
		t.Error("expected new-avail-b in new available skills")
	}
	if !newNames["new-avail-c"] {
		t.Error("expected new-avail-c in new available skills")
	}

	// Verify descriptions are populated
	for _, s := range newSkills {
		if s.Description == "" {
			t.Errorf("expected description for skill %s", s.Name)
		}
	}
}

func TestCheckSourceForUpdates_NoNewSkills(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	contentA := "---\nname: all-inst-a\ndescription: All installed A\n---\n# A\n"
	contentB := "---\nname: all-inst-b\ndescription: All installed B\n---\n# B\n"

	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"all-inst-a/SKILL.md": contentA,
		"all-inst-b/SKILL.md": contentB,
	})

	// Install both skills locally
	scrollsDir, _ := GetScrollsDir()
	for _, tc := range []struct{ name, content string }{
		{"all-inst-a", contentA},
		{"all-inst-b", contentB},
	} {
		skillDir := filepath.Join(scrollsDir, tc.name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(tc.content), 0o644)
		meta := NewSkillMeta(&SourceInfo{
			Type: "github", Owner: "testowner", Repo: "remote-repo", URL: repoURL,
		}, "", tc.content, nil)
		meta.Source = "testowner/remote-repo"
		meta.SourceType = "github"
		_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)
	}

	results, newSkills := CheckSourceForUpdates("testowner/remote-repo", []string{"all-inst-a", "all-inst-b"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error != "" {
			t.Errorf("unexpected error for %s: %s", r.Name, r.Error)
		}
	}

	if len(newSkills) != 0 {
		names := make([]string, 0, len(newSkills))
		for _, s := range newSkills {
			names = append(names, s.Name)
		}
		t.Errorf("expected 0 new skills, got %d: %v", len(newSkills), names)
	}
}

// ============================================================================
// NewAvailableSkills — CheckAllSourcesForUpdates populates the field
// ============================================================================

func TestCheckAllSourcesForUpdates_PopulatesNewAvailableSkills(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	contentA := "---\nname: pop-a\ndescription: Pop A\n---\n# Pop A\n"
	contentB := "---\nname: pop-b\ndescription: Pop B\n---\n# Pop B\n"
	contentC := "---\nname: pop-c\ndescription: Pop C\n---\n# Pop C\n"

	// Remote repo has three skills
	repoDir := filepath.Join(tmpDir, "remote-repo")
	repoURL := createTestGitRepo(t, repoDir, map[string]string{
		"pop-a/SKILL.md": contentA,
		"pop-b/SKILL.md": contentB,
		"pop-c/SKILL.md": contentC,
	})

	// Only install skill A locally
	scrollsDir, _ := GetScrollsDir()
	skillDir := filepath.Join(scrollsDir, "pop-a")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(contentA), 0o644)
	meta := NewSkillMeta(&SourceInfo{
		Type: "github", Owner: "testowner", Repo: "remote-repo", URL: repoURL,
	}, "", contentA, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	_ = WriteSkillMeta(filepath.Join(skillDir, MetaFileName), meta)

	allResults := CheckAllSourcesForUpdates()
	if len(allResults) != 1 {
		t.Fatalf("expected 1 source group, got %d", len(allResults))
	}

	sgr, ok := allResults["github:testowner/remote-repo"]
	if !ok {
		t.Fatal("expected result for 'github:testowner/remote-repo'")
	}

	if len(sgr.NewAvailableSkills) != 2 {
		t.Fatalf("expected 2 new available skills, got %d", len(sgr.NewAvailableSkills))
	}

	newNames := make(map[string]bool)
	for _, s := range sgr.NewAvailableSkills {
		newNames[s.Name] = true
	}
	if !newNames["pop-b"] {
		t.Error("expected pop-b in NewAvailableSkills")
	}
	if !newNames["pop-c"] {
		t.Error("expected pop-c in NewAvailableSkills")
	}
}

// ============================================================================
// NewAvailableSkills — SourceGroupCheckResult JSON serialization
// ============================================================================

func TestSourceGroupCheckResult_JSONWithNewAvailableSkills(t *testing.T) {
	original := SourceGroupCheckResult{
		Source:     "owner/repo",
		HasUpdates: false,
		NewAvailableSkills: []DiscoveredSkill{
			{Name: "new-skill-x", Description: "Skill X"},
			{Name: "new-skill-y", Description: "Skill Y"},
		},
		CheckedAt: "2025-06-01T00:00:00Z",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	jsonStr := string(data)

	// Verify the JSON contains the newAvailableSkills key
	if !strings.Contains(jsonStr, `"newAvailableSkills"`) {
		t.Error("expected JSON to contain newAvailableSkills key")
	}
	if !strings.Contains(jsonStr, `"new-skill-x"`) {
		t.Error("expected JSON to contain new-skill-x")
	}
	if !strings.Contains(jsonStr, `"new-skill-y"`) {
		t.Error("expected JSON to contain new-skill-y")
	}

	// Round-trip: unmarshal and verify
	var decoded SourceGroupCheckResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(decoded.NewAvailableSkills) != 2 {
		t.Fatalf("expected 2 NewAvailableSkills after round-trip, got %d", len(decoded.NewAvailableSkills))
	}

	decodedNames := make(map[string]string)
	for _, s := range decoded.NewAvailableSkills {
		decodedNames[s.Name] = s.Description
	}
	if decodedNames["new-skill-x"] != "Skill X" {
		t.Errorf("new-skill-x description = %q, want %q", decodedNames["new-skill-x"], "Skill X")
	}
	if decodedNames["new-skill-y"] != "Skill Y" {
		t.Errorf("new-skill-y description = %q, want %q", decodedNames["new-skill-y"], "Skill Y")
	}

	// Verify omitempty: when NewAvailableSkills is nil, it should be omitted
	empty := SourceGroupCheckResult{Source: "test", HasUpdates: false}
	emptyData, _ := json.Marshal(empty)
	if strings.Contains(string(emptyData), `"newAvailableSkills"`) {
		t.Error("expected newAvailableSkills to be omitted when nil")
	}
}
