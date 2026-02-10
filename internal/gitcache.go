package scribe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
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
	case "git":
		host := hostFromURL(source.URL)
		if host == "" {
			return ""
		}
		return filepath.Join(host, source.Owner, source.Repo)
	default:
		return ""
	}
}

// CloneOrUpdateRepo ensures a repo is present and up-to-date in the cache.
// Returns the path to the repository worktree, whether it's a cached directory,
// and whether authentication was required (indicating a private repository).
// If isCached is true, the caller must NOT delete the returned directory.
// If isCached is false, the caller is responsible for cleanup.
func CloneOrUpdateRepo(source *SourceInfo) (repoDir string, isCached, authRequired bool, err error) {
	cacheKey := CacheKeyForSource(source)
	if cacheKey == "" {
		// Not cacheable -- fall back to temp dir clone
		return cloneToTempDir(source)
	}

	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", false, false, fmt.Errorf("failed to get cache dir: %w", err)
	}

	repoPath := filepath.Join(cacheDir, cacheKey)

	// Try to open existing cached repo
	repo, err := git.PlainOpen(repoPath)
	if err == nil {
		// Cached repo exists -- fetch updates
		authNeeded, fetchErr := fetchRepo(repo, source)
		if fetchErr != nil {
			Logger.Warn("fetch failed, re-cloning", "path", repoPath, "error", fetchErr)
			_ = os.RemoveAll(repoPath)
			return cloneToCache(repoPath, source)
		}
		// Reset worktree to match the remote
		if err := resetToRemote(repo, source.Ref); err != nil {
			Logger.Warn("reset failed, re-cloning", "path", repoPath, "error", err)
			_ = os.RemoveAll(repoPath)
			return cloneToCache(repoPath, source)
		}
		return repoPath, true, authNeeded, nil
	}

	// Not cached or corrupted -- remove any remnants and clone fresh
	_ = os.RemoveAll(repoPath)
	return cloneToCache(repoPath, source)
}

// cloneToCache clones into the cache directory.
// For HTTPS URLs, it tries without auth first to detect whether the repo is public.
func cloneToCache(repoPath string, source *SourceInfo) (repoDir string, isCached, authRequired bool, err error) {
	if err = os.MkdirAll(filepath.Dir(repoPath), 0o755); err != nil {
		return "", false, false, fmt.Errorf("failed to create cache dir: %w", err)
	}

	cloneURL := buildCloneURL(source)
	Logger.Info("cloning repository", "url", cloneURL, "path", repoPath)

	// For SSH URLs, always use auth
	if isSSHURL(source.URL) {
		if err := cloneWithRefFallback(repoPath, cloneURL, source, authForSource(source)); err != nil {
			return "", false, false, fmt.Errorf("git clone failed: %w", err)
		}
		return repoPath, true, true, nil
	}

	// For HTTPS URLs, try without auth first to detect public repos
	if err := cloneWithRefFallback(repoPath, cloneURL, source, nil); err == nil {
		return repoPath, true, false, nil
	} else if !IsAuthError(err) {
		return "", false, false, fmt.Errorf("git clone failed: %w", err)
	}

	// Auth error -- retry with credentials
	Logger.Info("retrying clone with authentication", "url", cloneURL)
	_ = os.RemoveAll(repoPath)
	auth := authForSource(source)
	if auth == nil {
		return "", false, false, fmt.Errorf("git clone failed: authentication required but no credentials available")
	}
	if err := cloneWithRefFallback(repoPath, cloneURL, source, auth); err != nil {
		return "", false, false, fmt.Errorf("git clone failed: %w", err)
	}
	return repoPath, true, true, nil
}

// cloneWithRefFallback attempts a clone, falling back from branch to tag ref if needed.
// On failure, it cleans up the repoPath directory.
func cloneWithRefFallback(repoPath, cloneURL string, source *SourceInfo, auth transport.AuthMethod) error {
	opts := &git.CloneOptions{
		URL:   cloneURL,
		Depth: 1,
		Auth:  auth,
	}
	if source.Ref != "" {
		opts.SingleBranch = true
		opts.ReferenceName = plumbing.NewBranchReferenceName(source.Ref)
	}

	_, err := git.PlainClone(repoPath, false, opts)
	if err != nil && source.Ref != "" {
		// Branch ref failed, retry as tag
		_ = os.RemoveAll(repoPath)
		opts.ReferenceName = plumbing.NewTagReferenceName(source.Ref)
		_, err = git.PlainClone(repoPath, false, opts)
	}
	if err != nil {
		_ = os.RemoveAll(repoPath)
	}
	return err
}

// cloneToTempDir clones to a temp directory for non-cacheable sources.
// Caller must clean up the returned directory.
func cloneToTempDir(source *SourceInfo) (tempDir string, isCached, authRequired bool, err error) {
	tempDir, err = os.MkdirTemp("", "scribe-clone-*")
	if err != nil {
		return "", false, false, err
	}

	cloneURL := buildCloneURL(source)

	// For SSH URLs, always use auth
	if isSSHURL(source.URL) {
		if err := cloneWithRefFallback(tempDir, cloneURL, source, authForSource(source)); err != nil {
			return "", false, false, fmt.Errorf("git clone failed: %w", err)
		}
		return tempDir, false, true, nil
	}

	// For HTTPS URLs, try without auth first to detect public repos
	if err := cloneWithRefFallback(tempDir, cloneURL, source, nil); err == nil {
		return tempDir, false, false, nil
	} else if !IsAuthError(err) {
		return "", false, false, fmt.Errorf("git clone failed: %w", err)
	}

	// Auth error -- retry with credentials
	// cloneWithRefFallback already cleaned up, recreate temp dir
	tempDir, err = os.MkdirTemp("", "scribe-clone-*")
	if err != nil {
		return "", false, false, err
	}
	auth := authForSource(source)
	if auth == nil {
		_ = os.RemoveAll(tempDir)
		return "", false, false, fmt.Errorf("git clone failed: authentication required but no credentials available")
	}
	if err := cloneWithRefFallback(tempDir, cloneURL, source, auth); err != nil {
		return "", false, false, fmt.Errorf("git clone failed: %w", err)
	}
	return tempDir, false, true, nil
}

// fetchRepo fetches the latest changes for a cached repo.
// Returns whether authentication was required.
func fetchRepo(repo *git.Repository, source *SourceInfo) (authRequired bool, err error) {
	// For SSH URLs, always use auth
	if isSSHURL(source.URL) {
		err := repo.Fetch(&git.FetchOptions{
			Depth: 1,
			Force: true,
			Auth:  authForSource(source),
		})
		if err == git.NoErrAlreadyUpToDate {
			return true, nil
		}
		return true, err
	}

	// For HTTPS URLs, try without auth first
	err = repo.Fetch(&git.FetchOptions{
		Depth: 1,
		Force: true,
	})
	if err == nil || err == git.NoErrAlreadyUpToDate {
		return false, nil
	}
	if !IsAuthError(err) {
		return false, err
	}

	// Auth error -- retry with credentials
	auth := authForSource(source)
	if auth == nil {
		return false, fmt.Errorf("fetch failed: authentication required but no credentials available")
	}
	err = repo.Fetch(&git.FetchOptions{
		Depth: 1,
		Force: true,
		Auth:  auth,
	})
	if err == git.NoErrAlreadyUpToDate {
		return true, nil
	}
	return true, err
}

// resetToRemote resets the worktree to match the fetched remote state.
func resetToRemote(repo *git.Repository, ref string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	var targetHash plumbing.Hash
	var found bool

	if ref != "" {
		// Look for the specific ref as a remote branch, then as a tag
		remoteRef, err := repo.Reference(
			plumbing.NewRemoteReferenceName("origin", ref), true)
		if err == nil {
			targetHash = remoteRef.Hash()
			found = true
		} else {
			tagRef, err := repo.Reference(
				plumbing.NewTagReferenceName(ref), true)
			if err == nil {
				targetHash = tagRef.Hash()
				found = true
			}
		}
	} else {
		// No specific ref -- find the remote tracking branch for the local HEAD.
		// The local HEAD branch name (e.g. "main") tells us which remote ref to use.
		headRef, err := repo.Head()
		if err == nil {
			branchName := headRef.Name().Short()
			remoteRef, err := repo.Reference(
				plumbing.NewRemoteReferenceName("origin", branchName), true)
			if err == nil {
				targetHash = remoteRef.Hash()
				found = true
			}
		}
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

	if isSSHURL(cloneURL) {
		// SSH URLs: git@host:owner/repo.git
		if !strings.HasSuffix(cloneURL, ".git") {
			cloneURL += ".git"
		}
		return cloneURL
	}

	// HTTPS URLs
	if !strings.HasSuffix(cloneURL, ".git") {
		cloneURL += ".git"
	}
	return cloneURL
}

// GetHeadCommitInfo extracts the HEAD commit's short hash and date from a git repo directory.
// Returns nil if the directory is not a git repo or the commit can't be read.
func GetHeadCommitInfo(repoDir string) *GitCommitInfo {
	if repoDir == "" {
		return nil
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return nil
	}

	head, err := repo.Head()
	if err != nil {
		return nil
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil
	}

	return &GitCommitInfo{
		Hash: commit.Hash.String()[:7],
		Date: commit.Author.When.UTC().Format(time.RFC3339),
	}
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
