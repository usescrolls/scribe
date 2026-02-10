package scribe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestCacheKeyForSource(t *testing.T) {
	tests := []struct {
		name     string
		source   *SourceInfo
		expected string
	}{
		{
			name: "github source",
			source: &SourceInfo{
				Type:  "github",
				Owner: "user",
				Repo:  "repo",
			},
			expected: filepath.Join("github.com", "user", "repo"),
		},
		{
			name: "gitlab source",
			source: &SourceInfo{
				Type:  "gitlab",
				Owner: "org",
				Repo:  "project",
			},
			expected: filepath.Join("gitlab.com", "org", "project"),
		},
		{
			name: "local source returns empty",
			source: &SourceInfo{
				Type:      "local",
				LocalPath: "/some/path",
			},
			expected: "",
		},
		{
			name: "zip source returns empty",
			source: &SourceInfo{
				Type: "zip",
				URL:  "https://example.com/archive.zip",
			},
			expected: "",
		},
		{
			name: "unknown type returns empty",
			source: &SourceInfo{
				Type: "well-known",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CacheKeyForSource(tt.source)
			if result != tt.expected {
				t.Errorf("CacheKeyForSource() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCacheKeyForSource_GitLab(t *testing.T) {
	source := &SourceInfo{Type: "gitlab", Owner: "org", Repo: "project"}
	key := CacheKeyForSource(source)
	expected := filepath.Join("gitlab.com", "org", "project")
	if key != expected {
		t.Errorf("CacheKeyForSource(gitlab) = %q, want %q", key, expected)
	}
}

func TestBuildCloneURL(t *testing.T) {
	tests := []struct {
		name     string
		source   *SourceInfo
		expected string
	}{
		{
			name:     "URL without .git suffix",
			source:   &SourceInfo{URL: "https://github.com/user/repo"},
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "URL already has .git suffix",
			source:   &SourceInfo{URL: "https://github.com/user/repo.git"},
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "gitlab URL",
			source:   &SourceInfo{URL: "https://gitlab.com/org/project"},
			expected: "https://gitlab.com/org/project.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCloneURL(tt.source)
			if result != tt.expected {
				t.Errorf("buildCloneURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildCloneURL_SSHWithGitSuffix(t *testing.T) {
	source := &SourceInfo{URL: "git@github.com:user/repo.git"}
	got := buildCloneURL(source)
	if got != "git@github.com:user/repo.git" {
		t.Errorf("buildCloneURL = %q, want 'git@github.com:user/repo.git'", got)
	}
}

func TestBuildCloneURL_SSHWithoutGitSuffix(t *testing.T) {
	source := &SourceInfo{URL: "git@github.com:user/repo"}
	got := buildCloneURL(source)
	if got != "git@github.com:user/repo.git" {
		t.Errorf("buildCloneURL = %q, want 'git@github.com:user/repo.git'", got)
	}
}

func TestBuildCloneURL_AlreadyHasGit(t *testing.T) {
	source := &SourceInfo{URL: "https://github.com/u/r.git"}
	got := buildCloneURL(source)
	if got != "https://github.com/u/r.git" {
		t.Errorf("buildCloneURL = %q, want 'https://github.com/u/r.git'", got)
	}
}

func TestBuildCloneURL_NeedsGitSuffix(t *testing.T) {
	source := &SourceInfo{URL: "https://github.com/u/r"}
	got := buildCloneURL(source)
	if got != "https://github.com/u/r.git" {
		t.Errorf("buildCloneURL = %q, want 'https://github.com/u/r.git'", got)
	}
}

func TestGetCacheDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	dir, err := GetCacheDir()
	if err != nil {
		t.Fatalf("GetCacheDir() error: %v", err)
	}

	expected := filepath.Join(home, ".scribe", "cache")
	if dir != expected {
		t.Errorf("GetCacheDir() = %q, want %q", dir, expected)
	}
}

func TestGetCacheDir_IsolatedHome(t *testing.T) {
	tmpDir := setupTempHome(t)
	dir, err := GetCacheDir()
	if err != nil {
		t.Fatalf("GetCacheDir error: %v", err)
	}
	expected := filepath.Join(tmpDir, ".scribe", "cache")
	if dir != expected {
		t.Errorf("GetCacheDir = %q, want %q", dir, expected)
	}
}

func TestClearCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Create cache dir with some content
	cacheDir := filepath.Join(tmpDir, ".scribe", "cache")
	fakeRepo := filepath.Join(cacheDir, "github.com", "user", "repo")
	if err := os.MkdirAll(fakeRepo, 0o755); err != nil {
		t.Fatalf("failed to create fake repo dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fakeRepo, "dummy"), []byte("data"), 0o644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}

	// Clear cache
	if err := ClearCache(); err != nil {
		t.Fatalf("ClearCache() error: %v", err)
	}

	// Cache dir should still exist but be empty
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatalf("failed to read cache dir after clear: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected cache dir to be empty, got %d entries", len(entries))
	}
}

func TestClearCacheForSource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Create two cached repos
	cacheDir := filepath.Join(tmpDir, ".scribe", "cache")
	repo1 := filepath.Join(cacheDir, "github.com", "user", "repo1")
	repo2 := filepath.Join(cacheDir, "github.com", "user", "repo2")
	if err := os.MkdirAll(repo1, 0o755); err != nil {
		t.Fatalf("failed to create repo1: %v", err)
	}
	if err := os.MkdirAll(repo2, 0o755); err != nil {
		t.Fatalf("failed to create repo2: %v", err)
	}

	// Clear only repo1
	source := &SourceInfo{Type: "github", Owner: "user", Repo: "repo1"}
	if err := ClearCacheForSource(source); err != nil {
		t.Fatalf("ClearCacheForSource() error: %v", err)
	}

	// repo1 should be gone
	if _, err := os.Stat(repo1); !os.IsNotExist(err) {
		t.Error("expected repo1 to be removed")
	}
	// repo2 should still exist
	if _, err := os.Stat(repo2); err != nil {
		t.Error("expected repo2 to still exist")
	}
}

func TestClearCacheForSource_NonCacheable(t *testing.T) {
	// Non-cacheable source should be a no-op
	source := &SourceInfo{Type: "local", LocalPath: "/some/path"}
	if err := ClearCacheForSource(source); err != nil {
		t.Fatalf("ClearCacheForSource() for local source should not error: %v", err)
	}
}

func TestCloneOrUpdateRepo_GitHubSource_CloneToCache(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create a local git repo to act as the "remote"
	remoteDir := filepath.Join(tmpDir, "remote-repo")
	createTestGitRepo(t, remoteDir, map[string]string{
		"SKILL.md": "---\nname: cached-skill\ndescription: Cached\n---\n# Cached\n",
	})

	source := &SourceInfo{
		Type:  "github",
		Owner: "testuser",
		Repo:  "testrepo",
		URL:   remoteDir, // local path as URL for testing
	}

	// First clone
	repoDir, isCached, err := CloneOrUpdateRepo(source)
	if err != nil {
		t.Fatalf("CloneOrUpdateRepo() first call error: %v", err)
	}
	if !isCached {
		t.Error("expected isCached=true for github source")
	}
	if repoDir == "" {
		t.Fatal("repoDir is empty")
	}

	// Verify SKILL.md exists in the cached dir
	skillPath := filepath.Join(repoDir, "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("SKILL.md not found in cloned repo: %v", err)
	}

	// Second call: should fetch/update existing cache
	repoDir2, isCached2, err := CloneOrUpdateRepo(source)
	if err != nil {
		t.Fatalf("CloneOrUpdateRepo() second call error: %v", err)
	}
	if !isCached2 {
		t.Error("expected isCached=true on second call")
	}
	if repoDir2 != repoDir {
		t.Errorf("expected same cached path, got %q vs %q", repoDir2, repoDir)
	}
}

func TestCloneOrUpdateRepo_NonCacheable_CloneToTempDir(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)

	// Create a local git repo
	remoteDir := filepath.Join(tmpDir, "remote-zip-repo")
	createTestGitRepo(t, remoteDir, map[string]string{
		"SKILL.md": "---\nname: temp-skill\ndescription: Temp\n---\n# Temp\n",
	})

	// zip type is not cacheable, so it goes to cloneToTempDir
	source := &SourceInfo{
		Type: "zip",
		URL:  remoteDir, // local path
	}

	repoDir, isCached, err := CloneOrUpdateRepo(source)
	if err != nil {
		t.Fatalf("CloneOrUpdateRepo(non-cacheable) error: %v", err)
	}
	if isCached {
		t.Error("expected isCached=false for non-cacheable source")
	}
	defer func() { _ = os.RemoveAll(repoDir) }()

	if _, err := os.Stat(filepath.Join(repoDir, "SKILL.md")); err != nil {
		t.Error("SKILL.md not found in temp clone")
	}
}

func TestCloneToCache_InvalidURL(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	source := &SourceInfo{
		Type:  "github",
		Owner: "testuser",
		Repo:  "nonexistent-repo-xyz",
		URL:   "https://invalid-host-that-does-not-exist.example.com/test/repo",
	}

	_, _, err := CloneOrUpdateRepo(source)
	if err == nil {
		t.Error("expected error for invalid clone URL")
	}
	// Verify the cache dir was cleaned up
	cacheDir := filepath.Join(tmpDir, ".scribe", "cache", "github.com", "testuser", "nonexistent-repo-xyz")
	if _, statErr := os.Stat(cacheDir); !os.IsNotExist(statErr) {
		t.Error("expected cache dir to be cleaned up after clone failure")
	}
}

func TestCloneToCache_WithRef(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create a local git repo with a branch
	remoteDir := filepath.Join(tmpDir, "remote-branch-repo")
	createTestGitRepo(t, remoteDir, map[string]string{
		"SKILL.md": "---\nname: branch-skill\ndescription: Branch\n---\n# Branch\n",
	})

	// Create a new branch
	repo, err := git.PlainOpen(remoteDir)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	headRef, err := repo.Head()
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	branchRefName := plumbing.NewBranchReferenceName("test-branch")
	ref := plumbing.NewHashReference(branchRefName, headRef.Hash())
	err = repo.Storer.SetReference(ref)
	if err != nil {
		t.Fatalf("set branch ref: %v", err)
	}

	source := &SourceInfo{
		Type:  "github",
		Owner: "testuser",
		Repo:  "branch-repo",
		URL:   remoteDir,
		Ref:   "test-branch",
	}

	repoDir, isCached, err := CloneOrUpdateRepo(source)
	if err != nil {
		t.Fatalf("CloneOrUpdateRepo with ref error: %v", err)
	}
	if !isCached {
		t.Error("expected isCached=true")
	}
	if _, err := os.Stat(filepath.Join(repoDir, "SKILL.md")); err != nil {
		t.Error("SKILL.md not found in branch clone")
	}
}

func TestCloneToCache_RefFailsFallbackToTag(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	remoteDir := filepath.Join(tmpDir, "tag-only-repo")
	createTestGitRepo(t, remoteDir, map[string]string{
		"SKILL.md": "---\nname: tag-only\ndescription: Tag\n---\n# Tag\n",
	})

	// Create a tag but no branch with this name
	repo, _ := git.PlainOpen(remoteDir)
	headRef, _ := repo.Head()
	tagRefName := plumbing.NewTagReferenceName("v2.0.0")
	tagRef := plumbing.NewHashReference(tagRefName, headRef.Hash())
	_ = repo.Storer.SetReference(tagRef)

	source := &SourceInfo{
		Type:  "github",
		Owner: "test",
		Repo:  "tag-only-cache",
		URL:   remoteDir,
		Ref:   "v2.0.0",
	}

	repoDir, isCached, err := CloneOrUpdateRepo(source)
	if err != nil {
		t.Fatalf("CloneOrUpdateRepo with tag ref error: %v", err)
	}
	if !isCached {
		t.Error("expected isCached=true")
	}
	if _, err := os.Stat(filepath.Join(repoDir, "SKILL.md")); err != nil {
		t.Error("SKILL.md not found in tag clone")
	}
}

func TestCloneToTempDir_WithRef_FallbackToTag(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)

	// Create a local git repo with a tag
	remoteDir := filepath.Join(tmpDir, "remote-tag-repo")
	createTestGitRepo(t, remoteDir, map[string]string{
		"SKILL.md": "---\nname: tag-skill\ndescription: Tag\n---\n# Tag\n",
	})

	repo, err := git.PlainOpen(remoteDir)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	headRef, err := repo.Head()
	if err != nil {
		t.Fatalf("head: %v", err)
	}

	// Create a tag
	tagRefName := plumbing.NewTagReferenceName("v1.0.0")
	tagRef := plumbing.NewHashReference(tagRefName, headRef.Hash())
	err = repo.Storer.SetReference(tagRef)
	if err != nil {
		t.Fatalf("set tag ref: %v", err)
	}

	// Use a non-cacheable type to force cloneToTempDir
	source := &SourceInfo{
		Type: "zip",
		URL:  remoteDir,
		Ref:  "v1.0.0",
	}

	repoDir, isCached, err := CloneOrUpdateRepo(source)
	if err != nil {
		t.Fatalf("CloneOrUpdateRepo with tag error: %v", err)
	}
	if isCached {
		t.Error("expected isCached=false for non-cacheable source")
	}
	defer func() { _ = os.RemoveAll(repoDir) }()

	if _, err := os.Stat(filepath.Join(repoDir, "SKILL.md")); err != nil {
		t.Error("SKILL.md not found in tag clone")
	}
}

func TestCloneToTempDir_InvalidURL(t *testing.T) {
	InitLoggerCLI(false)

	source := &SourceInfo{
		Type: "zip",
		URL:  "https://invalid-host-xyz.example.com/nonexistent",
	}

	_, _, err := CloneOrUpdateRepo(source)
	if err == nil {
		t.Error("expected error for invalid clone URL in cloneToTempDir")
	}
}

func TestCloneToTempDir_InvalidRef(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)

	remoteDir := filepath.Join(tmpDir, "invalid-ref-repo")
	createTestGitRepo(t, remoteDir, map[string]string{
		"file.txt": "content",
	})

	source := &SourceInfo{
		Type: "zip", // non-cacheable, forces cloneToTempDir
		URL:  remoteDir,
		Ref:  "nonexistent-ref-xyz",
	}

	_, _, err := CloneOrUpdateRepo(source)
	if err == nil {
		t.Error("expected error for nonexistent ref in cloneToTempDir")
	}
}

func TestFetchRepo_AlreadyUpToDate(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)

	// Create a remote repo
	remoteDir := filepath.Join(tmpDir, "remote")
	createTestGitRepo(t, remoteDir, map[string]string{"file.txt": "hello"})

	// Clone it
	cloneDir := filepath.Join(tmpDir, "clone")
	repo, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
		URL:   remoteDir,
		Depth: 1,
	})
	if err != nil {
		t.Fatalf("clone: %v", err)
	}

	// Fetch again (should be already up to date = no error)
	source := &SourceInfo{Type: "local", URL: remoteDir}
	err = fetchRepo(repo, source)
	if err != nil {
		t.Errorf("fetchRepo() after fresh clone should return nil, got: %v", err)
	}
}

func TestResetToRemote_NoRef(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)

	remoteDir := filepath.Join(tmpDir, "remote")
	createTestGitRepo(t, remoteDir, map[string]string{"file.txt": "content"})

	cloneDir := filepath.Join(tmpDir, "clone")
	repo, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
		URL:   remoteDir,
		Depth: 1,
	})
	if err != nil {
		t.Fatalf("clone: %v", err)
	}

	// resetToRemote with empty ref (should use HEAD)
	err = resetToRemote(repo, "")
	if err != nil {
		t.Errorf("resetToRemote('') error: %v", err)
	}

	// Verify the working tree is intact
	content, err := os.ReadFile(filepath.Join(cloneDir, "file.txt"))
	if err != nil {
		t.Fatalf("read file after reset: %v", err)
	}
	if string(content) != "content" {
		t.Errorf("file content = %q, want 'content'", string(content))
	}
}

func TestResetToRemote_RefNotFound(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)

	remoteDir := filepath.Join(tmpDir, "remote")
	createTestGitRepo(t, remoteDir, map[string]string{"file.txt": "content"})

	cloneDir := filepath.Join(tmpDir, "clone")
	repo, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
		URL:   remoteDir,
		Depth: 1,
	})
	if err != nil {
		t.Fatalf("clone: %v", err)
	}

	// resetToRemote with nonexistent ref
	err = resetToRemote(repo, "nonexistent-branch-xyz")
	if err == nil {
		t.Error("expected error for nonexistent ref")
	}
	if !strings.Contains(err.Error(), "ref not found") {
		t.Errorf("error = %q, want it to contain 'ref not found'", err.Error())
	}
}

func TestCloneOrUpdateRepo_CorruptedCache(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create a real remote
	remoteDir := filepath.Join(tmpDir, "remote")
	createTestGitRepo(t, remoteDir, map[string]string{
		"SKILL.md": "---\nname: corrupt-test\ndescription: Test\n---\n# Test\n",
	})

	// Create a corrupted cache entry (exists but not a valid git repo)
	cacheDir := filepath.Join(tmpDir, ".scribe", "cache", "github.com", "testuser", "corrupt-repo")
	_ = os.MkdirAll(cacheDir, 0o755)
	_ = os.WriteFile(filepath.Join(cacheDir, "garbage"), []byte("not a repo"), 0o644)

	source := &SourceInfo{
		Type:  "github",
		Owner: "testuser",
		Repo:  "corrupt-repo",
		URL:   remoteDir,
	}

	// Should detect corruption and re-clone
	repoDir, isCached, err := CloneOrUpdateRepo(source)
	if err != nil {
		t.Fatalf("CloneOrUpdateRepo(corrupted) error: %v", err)
	}
	if !isCached {
		t.Error("expected isCached=true after re-clone")
	}
	if _, err := os.Stat(filepath.Join(repoDir, "SKILL.md")); err != nil {
		t.Error("SKILL.md not found after re-clone")
	}
}

// ============================================================================
// GetHeadCommitInfo (gitcache.go)
// ============================================================================

func TestGetHeadCommitInfo_ValidRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-head-info-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	repoDir := filepath.Join(tmpDir, "repo")
	createTestGitRepo(t, repoDir, map[string]string{
		"SKILL.md": "---\nname: test\ndescription: test\n---\n# Test\n",
	})

	info := GetHeadCommitInfo(repoDir)
	if info == nil {
		t.Fatal("GetHeadCommitInfo returned nil for valid git repo")
	}
	if len(info.Hash) != 7 {
		t.Errorf("Hash length = %d, want 7 (short hash)", len(info.Hash))
	}
	if info.Date == "" {
		t.Error("Date is empty")
	}
	// Verify date is valid RFC3339
	if !strings.Contains(info.Date, "T") || !strings.Contains(info.Date, "Z") {
		t.Errorf("Date = %q, does not look like RFC3339", info.Date)
	}
}

func TestGetHeadCommitInfo_NonGitDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-head-info-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	info := GetHeadCommitInfo(tmpDir)
	if info != nil {
		t.Errorf("GetHeadCommitInfo(non-git dir) = %+v, want nil", info)
	}
}

func TestGetHeadCommitInfo_EmptyString(t *testing.T) {
	info := GetHeadCommitInfo("")
	if info != nil {
		t.Errorf("GetHeadCommitInfo('') = %+v, want nil", info)
	}
}

func TestGetHeadCommitInfo_NonExistentDir(t *testing.T) {
	info := GetHeadCommitInfo("/tmp/nonexistent-scribe-dir-xyz-12345")
	if info != nil {
		t.Errorf("GetHeadCommitInfo(nonexistent) = %+v, want nil", info)
	}
}

func TestCloneOrUpdateRepo_FetchFails(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create a remote repo
	remoteDir := filepath.Join(tmpDir, "fetch-fail-remote")
	createTestGitRepo(t, remoteDir, map[string]string{
		"SKILL.md": "---\nname: fetch-fail\ndescription: test\n---\n# T\n",
	})

	source := &SourceInfo{
		Type:  "github",
		Owner: "test",
		Repo:  "fetch-fail",
		URL:   remoteDir,
	}

	// First clone
	_, _, err := CloneOrUpdateRepo(source)
	if err != nil {
		t.Fatalf("first clone error: %v", err)
	}

	// Now corrupt the cached repo's remote config to make fetch fail
	cacheDirPath, _ := GetCacheDir()
	repoPath := filepath.Join(cacheDirPath, "github.com", "test", "fetch-fail")

	// Overwrite the git config to break the remote
	gitConfigPath := filepath.Join(repoPath, ".git", "config")
	_ = os.WriteFile(gitConfigPath, []byte("[remote \"origin\"]\n\turl = /nonexistent/path.git\n\tfetch = +refs/heads/*:refs/remotes/origin/*\n"), 0o644)

	// Second call should detect fetch failure and re-clone
	repoDir, _, err := CloneOrUpdateRepo(source)
	if err != nil {
		t.Fatalf("re-clone after fetch fail error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "SKILL.md")); err != nil {
		t.Error("SKILL.md not found after re-clone")
	}
}
