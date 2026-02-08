package scribe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	// CacheDirName is the subdirectory for cached git clones
	CacheDirName = "cache"
)

// GetCacheDir returns the cache directory (~/.scribe/cache/)
func GetCacheDir() (string, error) {
	scribeDir, err := GetScribeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(scribeDir, CacheDirName), nil
}

// CacheKeyForSource returns the relative cache path for a SourceInfo,
// or empty string if the source type is not cacheable.
func CacheKeyForSource(source *SourceInfo) string {
	switch source.Type {
	case "github":
		return filepath.Join("github.com", source.Owner, source.Repo)
	case "gitlab":
		return filepath.Join("gitlab.com", source.Owner, source.Repo)
	case "bitbucket":
		return filepath.Join("bitbucket.org", source.Owner, source.Repo)
	default:
		return ""
	}
}

// CloneOrUpdateRepo ensures a repo is present and up-to-date in the cache.
// Returns the path to the repository worktree and whether it's a cached directory.
// If isCached is true, the caller must NOT delete the returned directory.
// If isCached is false, the caller is responsible for cleanup.
func CloneOrUpdateRepo(source *SourceInfo) (repoDir string, isCached bool, err error) {
	cacheKey := CacheKeyForSource(source)
	if cacheKey == "" {
		// Not cacheable -- fall back to temp dir clone
		return cloneToTempDir(source)
	}

	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", false, fmt.Errorf("failed to get cache dir: %w", err)
	}

	repoPath := filepath.Join(cacheDir, cacheKey)

	// Try to open existing cached repo
	repo, err := git.PlainOpen(repoPath)
	if err == nil {
		// Cached repo exists -- fetch updates
		if err := fetchRepo(repo); err != nil {
			Logger.Warn("fetch failed, re-cloning", "path", repoPath, "error", err)
			_ = os.RemoveAll(repoPath)
			return cloneToCache(repoPath, source)
		}
		// Reset worktree to match the remote
		if err := resetToRemote(repo, source.Ref); err != nil {
			Logger.Warn("reset failed, re-cloning", "path", repoPath, "error", err)
			_ = os.RemoveAll(repoPath)
			return cloneToCache(repoPath, source)
		}
		return repoPath, true, nil
	}

	// Not cached or corrupted -- remove any remnants and clone fresh
	_ = os.RemoveAll(repoPath)
	return cloneToCache(repoPath, source)
}

// cloneToCache clones into the cache directory.
func cloneToCache(repoPath string, source *SourceInfo) (repoDir string, isCached bool, err error) {
	if err = os.MkdirAll(filepath.Dir(repoPath), 0o755); err != nil {
		return "", false, fmt.Errorf("failed to create cache dir: %w", err)
	}

	cloneURL := buildCloneURL(source)

	opts := &git.CloneOptions{
		URL:   cloneURL,
		Depth: 1,
	}
	if source.Ref != "" {
		opts.SingleBranch = true
		opts.ReferenceName = plumbing.NewBranchReferenceName(source.Ref)
	}

	Logger.Info("cloning repository", "url", cloneURL, "path", repoPath)

	_, err = git.PlainClone(repoPath, false, opts)
	if err != nil && source.Ref != "" {
		// Branch ref failed, retry as tag
		_ = os.RemoveAll(repoPath)
		opts.ReferenceName = plumbing.NewTagReferenceName(source.Ref)
		_, err = git.PlainClone(repoPath, false, opts)
	}
	if err != nil {
		_ = os.RemoveAll(repoPath)
		return "", false, fmt.Errorf("git clone failed: %w", err)
	}

	return repoPath, true, nil
}

// cloneToTempDir clones to a temp directory for non-cacheable sources.
// Caller must clean up the returned directory.
func cloneToTempDir(source *SourceInfo) (tempDir string, isCached bool, err error) {
	tempDir, err = os.MkdirTemp("", "scribe-clone-*")
	if err != nil {
		return "", false, err
	}

	cloneURL := buildCloneURL(source)

	opts := &git.CloneOptions{
		URL:   cloneURL,
		Depth: 1,
	}
	if source.Ref != "" {
		opts.SingleBranch = true
		opts.ReferenceName = plumbing.NewBranchReferenceName(source.Ref)
	}

	_, err = git.PlainClone(tempDir, false, opts)
	if err != nil && source.Ref != "" {
		// Branch ref failed, retry as tag
		opts.ReferenceName = plumbing.NewTagReferenceName(source.Ref)
		_, err = git.PlainClone(tempDir, false, opts)
	}
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return "", false, fmt.Errorf("git clone failed: %w", err)
	}

	return tempDir, false, nil
}

// fetchRepo fetches the latest changes for a cached repo.
func fetchRepo(repo *git.Repository) error {
	err := repo.Fetch(&git.FetchOptions{
		Depth: 1,
		Force: true,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

// resetToRemote resets the worktree to match the fetched remote state.
func resetToRemote(repo *git.Repository, ref string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Get the remote HEAD reference
	remoteRefs, err := repo.References()
	if err != nil {
		return err
	}

	var targetHash plumbing.Hash
	var found bool

	err = remoteRefs.ForEach(func(r *plumbing.Reference) error {
		refName := r.Name().String()

		if ref == "" {
			// No specific ref requested -- use origin's HEAD
			if r.Name() == plumbing.HEAD || strings.HasPrefix(refName, "refs/remotes/origin/") {
				targetHash = r.Hash()
				found = true
			}
		} else {
			// Look for the specific ref in remotes
			if strings.HasSuffix(refName, "/"+ref) {
				targetHash = r.Hash()
				found = true
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("ref not found: %s", ref)
	}

	return wt.Reset(&git.ResetOptions{
		Commit: targetHash,
		Mode:   git.HardReset,
	})
}

// buildCloneURL constructs the clone URL from a SourceInfo.
func buildCloneURL(source *SourceInfo) string {
	cloneURL := source.URL
	if !strings.HasSuffix(cloneURL, ".git") {
		cloneURL += ".git"
	}
	return cloneURL
}

// ClearCache removes the entire cache directory.
func ClearCache() error {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(cacheDir); err != nil {
		return err
	}
	// Recreate the empty cache directory
	return EnsureDir(cacheDir)
}

// ClearCacheForSource removes the cached repo for a specific source.
func ClearCacheForSource(source *SourceInfo) error {
	cacheKey := CacheKeyForSource(source)
	if cacheKey == "" {
		return nil
	}
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(cacheDir, cacheKey))
}
