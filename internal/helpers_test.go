package scribe

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// createTestGitRepo initializes a non-bare git repo at dir with the given
// files committed, and returns a URL string suitable for use in source.URL.
//
// Because buildCloneURL always appends ".git", we create a symlink
// "dir.git" -> "dir" so that the final clone URL resolves correctly.
func createTestGitRepo(t *testing.T, dir string, files map[string]string) string {
	t.Helper()

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("git init: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	for name, content := range files {
		fullPath := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		if _, err := wt.Add(name); err != nil {
			t.Fatalf("git add: %v", err)
		}
	}
	_, err = wt.Commit("initial", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@test.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("git commit: %v", err)
	}

	// buildCloneURL appends ".git" to the URL. Create a symlink
	// "dir.git" -> "dir" so go-git can resolve both paths.
	gitPath := dir + ".git"
	_ = os.Symlink(dir, gitPath)

	// Return the path without ".git" suffix; buildCloneURL adds it.
	return dir
}

// setupTempHome creates a temp dir, sets HOME to it, and returns cleanup func.
func setupTempHome(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "scribe-boost-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })
	return tmpDir
}
