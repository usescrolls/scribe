package scribe

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ParseSourceString parses a source string into a SourceInfo.
// Supported formats:
//   - Local path: ./path, /absolute/path, ~/path
//   - GitHub URL: https://github.com/owner/repo
//   - GitLab URL: https://gitlab.com/owner/repo
//   - Bitbucket URL: https://bitbucket.org/owner/repo
//   - SSH URL: git@github.com:owner/repo, git@gitlab.com:owner/repo
//   - GitHub shorthand: owner/repo, owner/repo#branch, owner/repo/subpath
//   - Zip URL: https://example.com/skills.zip
func ParseSourceString(arg string) (*SourceInfo, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return nil, fmt.Errorf("empty source string")
	}

	// Check for local path
	if strings.HasPrefix(arg, "./") || strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, "~") {
		absPath, err := filepath.Abs(arg)
		if err != nil {
			return nil, err
		}
		return &SourceInfo{Type: "local", LocalPath: absPath}, nil
	}

	// Check for SSH URL (git@host:owner/repo.git)
	if strings.HasPrefix(arg, "git@") {
		return parseSSHURL(arg)
	}

	// Check for GitHub URL
	if strings.HasPrefix(arg, "https://github.com/") {
		return parseGitHubURL(arg)
	}

	// Check for GitLab URL
	if strings.HasPrefix(arg, "https://gitlab.com/") {
		return parseGitLabURL(arg)
	}

	// Check for Bitbucket URL
	if strings.HasPrefix(arg, "https://bitbucket.org/") {
		return parseBitbucketURL(arg)
	}

	// Check for generic URL
	if strings.HasPrefix(arg, "https://") || strings.HasPrefix(arg, "http://") {
		if strings.HasSuffix(arg, ".zip") {
			return &SourceInfo{Type: "zip", URL: arg}, nil
		}
		// Self-hosted git instances: https://host/owner/repo.git
		if strings.HasSuffix(arg, ".git") {
			return parseGenericGitURL(arg)
		}
		return &SourceInfo{Type: "well-known", URL: arg}, nil
	}

	// Assume GitHub shorthand: owner/repo or owner/repo#branch or owner/repo/path
	return parseGitHubShorthand(arg)
}

func parseGitHubShorthand(arg string) (*SourceInfo, error) {
	source := &SourceInfo{Type: "github"}

	// Check for branch/ref
	if idx := strings.Index(arg, "#"); idx != -1 {
		source.Ref = arg[idx+1:]
		arg = arg[:idx]
	}

	parts := strings.Split(arg, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid source: expected owner/repo format")
	}

	source.Owner = parts[0]
	source.Repo = parts[1]

	if len(parts) > 2 {
		source.Subpath = strings.Join(parts[2:], "/")
	}

	source.URL = fmt.Sprintf("https://github.com/%s/%s", source.Owner, source.Repo)
	return source, nil
}

func parseGitHubURL(url string) (*SourceInfo, error) {
	source := &SourceInfo{Type: "github", URL: url}

	path := strings.TrimPrefix(url, "https://github.com/")

	// Check for branch in URL (tree/branch/...)
	if idx := strings.Index(path, "/tree/"); idx != -1 {
		beforeTree := path[:idx]
		afterTree := path[idx+6:] // Skip "/tree/"

		parts := strings.Split(beforeTree, "/")
		if len(parts) >= 2 {
			source.Owner = parts[0]
			source.Repo = parts[1]
		}

		afterParts := strings.SplitN(afterTree, "/", 2)
		source.Ref = afterParts[0]
		if len(afterParts) > 1 {
			source.Subpath = afterParts[1]
		}
	} else {
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			source.Owner = parts[0]
			source.Repo = strings.TrimSuffix(parts[1], ".git")
		}
		if len(parts) > 2 {
			source.Subpath = strings.Join(parts[2:], "/")
		}
	}

	return source, nil
}

func parseGitLabURL(url string) (*SourceInfo, error) {
	source := &SourceInfo{Type: "gitlab", URL: url}

	path := strings.TrimPrefix(url, "https://gitlab.com/")
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		// GitLab supports nested groups: group/subgroup/project
		// Last component is always the repo, everything before is the owner/group path
		source.Repo = strings.TrimSuffix(parts[len(parts)-1], ".git")
		source.Owner = strings.Join(parts[:len(parts)-1], "/")
	}

	return source, nil
}

// ReconstructSource creates a SourceInfo from stored SkillMeta.
// This is the inverse of formatSource — it rebuilds the full source reference
// so the skill can be fetched again for updates or checks.
func ReconstructSource(meta *SkillMeta) *SourceInfo {
	source := &SourceInfo{
		Type: meta.SourceType,
		URL:  meta.SourceURL,
	}

	switch meta.SourceType {
	case "github":
		srcStr := meta.Source
		parts := strings.Split(srcStr, "/")
		if len(parts) >= 2 {
			source.Owner = parts[0]
			repoAndRef := strings.SplitN(parts[1], "#", 2)
			source.Repo = repoAndRef[0]
			if len(repoAndRef) > 1 {
				source.Ref = repoAndRef[1]
			}
			if len(parts) > 2 {
				source.Subpath = strings.Join(parts[2:], "/")
			}
		}
		if source.URL == "" {
			source.URL = "https://github.com/" + source.Owner + "/" + source.Repo
		}

	case "gitlab":
		srcStr := meta.Source
		parts := strings.Split(srcStr, "/")
		if len(parts) >= 2 {
			source.Owner = parts[0]
			source.Repo = parts[1]
		}
		if source.URL == "" {
			source.URL = "https://gitlab.com/" + source.Owner + "/" + source.Repo
		}

	case "bitbucket":
		srcStr := meta.Source
		parts := strings.Split(srcStr, "/")
		if len(parts) >= 2 {
			source.Owner = parts[0]
			source.Repo = parts[1]
		}
		if source.URL == "" {
			source.URL = "https://bitbucket.org/" + source.Owner + "/" + source.Repo
		}

	case "git":
		srcStr := meta.Source
		parts := strings.Split(srcStr, "/")
		if len(parts) >= 2 {
			source.Repo = parts[len(parts)-1]
			source.Owner = strings.Join(parts[:len(parts)-1], "/")
		}
		// URL must be preserved from meta for git type (no implicit host)

	case "zip":
		if source.URL == "" {
			source.URL = meta.Source
		}

	case "url", "well-known":
		if source.URL == "" {
			source.URL = meta.Source
		}

	case "local":
		source.LocalPath = meta.Source
	}

	if meta.SkillPath != "" {
		source.Subpath = meta.SkillPath
	}

	return source
}

// parseSSHURL parses git@host:owner/repo.git format.
// Supported: git@github.com:owner/repo.git, git@gitlab.com:owner/repo, git@bitbucket.org:owner/repo.git
func parseSSHURL(arg string) (*SourceInfo, error) {
	// Format: git@host:owner/repo.git
	// Strip "git@"
	rest := strings.TrimPrefix(arg, "git@")

	// Split on ":"
	colonIdx := strings.IndexByte(rest, ':')
	if colonIdx == -1 {
		return nil, fmt.Errorf("invalid SSH URL: expected git@host:owner/repo format")
	}

	host := rest[:colonIdx]
	path := rest[colonIdx+1:]
	path = strings.TrimSuffix(path, ".git")

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid SSH URL: expected git@host:owner/repo format")
	}

	source := &SourceInfo{
		URL: arg, // Keep the SSH URL as-is for cloning
	}

	// Determine source type from host
	switch host {
	case "github.com":
		source.Type = "github"
		source.Owner = parts[0]
		source.Repo = parts[1]
		if len(parts) > 2 {
			source.Subpath = strings.Join(parts[2:], "/")
		}
	case "gitlab.com":
		// GitLab supports nested groups: group/subgroup/project
		source.Type = "gitlab"
		source.Repo = parts[len(parts)-1]
		source.Owner = strings.Join(parts[:len(parts)-1], "/")
	case "bitbucket.org":
		source.Type = "bitbucket"
		source.Owner = parts[0]
		source.Repo = parts[1]
		if len(parts) > 2 {
			source.Subpath = strings.Join(parts[2:], "/")
		}
	default:
		// Self-hosted git instance — last component is repo, rest is owner/group
		source.Type = "git"
		source.Repo = parts[len(parts)-1]
		source.Owner = strings.Join(parts[:len(parts)-1], "/")
	}

	return source, nil
}

func parseBitbucketURL(url string) (*SourceInfo, error) {
	source := &SourceInfo{Type: "bitbucket", URL: url}

	path := strings.TrimPrefix(url, "https://bitbucket.org/")
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		source.Owner = parts[0]
		source.Repo = strings.TrimSuffix(parts[1], ".git")
	}

	return source, nil
}

// parseGenericGitURL parses HTTPS git URLs from self-hosted instances.
// Format: https://host/owner/repo.git or https://host:port/group/subgroup/repo.git
func parseGenericGitURL(url string) (*SourceInfo, error) {
	source := &SourceInfo{Type: "git", URL: url}

	// Strip protocol
	path := strings.TrimPrefix(url, "https://")
	path = strings.TrimPrefix(path, "http://")

	// Split into host and rest
	slashIdx := strings.IndexByte(path, '/')
	if slashIdx == -1 {
		return nil, fmt.Errorf("invalid git URL: no path after host")
	}
	ownerRepo := path[slashIdx+1:]

	parts := strings.Split(ownerRepo, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid git URL: expected at least owner/repo.git")
	}

	// Last component is repo (strip .git), everything before is owner/group
	source.Repo = strings.TrimSuffix(parts[len(parts)-1], ".git")
	source.Owner = strings.Join(parts[:len(parts)-1], "/")

	return source, nil
}
