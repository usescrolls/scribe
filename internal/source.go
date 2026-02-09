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
		source.Owner = parts[0]
		source.Repo = strings.TrimSuffix(parts[1], ".git")
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
