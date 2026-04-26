package scribe

import (
	"fmt"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
)

// SanitizeSubpath validates that a subpath does not contain ".." segments
// that could escape the repository root directory.
func SanitizeSubpath(subpath string) (string, error) {
	// Normalize backslashes for consistent handling
	normalized := strings.ReplaceAll(subpath, "\\", "/")

	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return "", fmt.Errorf("unsafe subpath: %q contains path traversal segments", subpath)
		}
	}
	return subpath, nil
}

// IsSubpathSafe validates that a resolved subpath stays within the base directory.
func IsSubpathSafe(basePath, subpath string) bool {
	normalizedBase := filepath.Clean(basePath)
	normalizedTarget := filepath.Clean(filepath.Join(basePath, subpath))

	return normalizedTarget == normalizedBase ||
		strings.HasPrefix(normalizedTarget, normalizedBase+string(os.PathSeparator))
}

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
	arg = normalizeSourceString(arg)
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

	// Check for SSH URL (git@host:owner/repo.git, user@host:owner/repo.git, ssh://...)
	if isSSHURL(arg) {
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
		// Self-hosted git instances: https://host/owner/repo(.git)
		if strings.HasSuffix(arg, ".git") || isLikelyGitURL(arg) {
			return parseGenericGitURL(arg)
		}
		return &SourceInfo{Type: "well-known", URL: arg}, nil
	}

	// Check for host-qualified git shorthand used by tools like:
	// gh repo clone ghe.example.com/owner/repo
	if isHostQualifiedGitShorthand(arg) {
		return parseHostQualifiedGitShorthand(arg)
	}

	// Assume GitHub shorthand: owner/repo or owner/repo#branch or owner/repo/path
	return parseGitHubShorthand(arg)
}

func normalizeSourceString(arg string) string {
	arg = strings.TrimSpace(stripMatchingQuotes(arg))
	if arg == "" {
		return ""
	}

	fields := strings.Fields(arg)
	if len(fields) == 0 {
		return arg
	}

	lower := make([]string, len(fields))
	for i, field := range fields {
		lower[i] = strings.ToLower(field)
	}

	switch {
	case len(fields) >= 4 && lower[0] == "npx" && lower[1] == "skills" && lower[2] == "add":
		return commandSourceArg(fields[3:])
	case len(fields) >= 3 && lower[0] == "skills" && lower[1] == "add":
		return commandSourceArg(fields[2:])
	case len(fields) >= 3 && lower[0] == "scribe" && lower[1] == "install":
		return commandSourceArg(fields[2:])
	case len(fields) >= 3 && lower[0] == "git" && lower[1] == "clone":
		return commandSourceArg(fields[2:])
	case len(fields) >= 4 && lower[0] == "gh" && lower[1] == "repo" && lower[2] == "clone":
		return commandSourceArg(fields[3:])
	default:
		return arg
	}
}

func commandSourceArg(args []string) string {
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if strings.HasPrefix(arg, "-") {
			if commandFlagTakesValue(arg) {
				skipNext = true
			}
			continue
		}
		return stripMatchingQuotes(arg)
	}
	return ""
}

func commandFlagTakesValue(arg string) bool {
	if strings.Contains(arg, "=") {
		return false
	}
	switch arg {
	case "-b", "--branch", "-c", "--config", "--depth", "-o", "--origin",
		"--reference", "--reference-if-able", "--separate-git-dir",
		"--template", "-u", "--upload-pack":
		return true
	default:
		return false
	}
}

func stripMatchingQuotes(s string) string {
	if len(s) < 2 {
		return s
	}
	if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}
	return s
}

func isLikelyGitURL(raw string) bool {
	u, err := neturl.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	return len(pathSegments(u.Path)) >= 2
}

func isHostQualifiedGitShorthand(arg string) bool {
	if strings.Contains(arg, "://") || isSSHURL(arg) {
		return false
	}
	base := arg
	if before, _, ok := strings.Cut(base, "#"); ok {
		base = before
	}
	parts := strings.Split(strings.Trim(base, "/"), "/")
	return len(parts) >= 3 && strings.Contains(parts[0], ".")
}

func parseHostQualifiedGitShorthand(arg string) (*SourceInfo, error) {
	host, rest, _ := strings.Cut(arg, "/")
	rawURL := "https://" + host + "/" + rest
	switch hostNameOnly(host) {
	case "github.com":
		return parseGitHubURL(rawURL)
	case "gitlab.com":
		return parseGitLabURL(rawURL)
	case "bitbucket.org":
		return parseBitbucketURL(rawURL)
	default:
		return parseGenericGitURL(rawURL)
	}
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
		sp, err := SanitizeSubpath(strings.Join(parts[2:], "/"))
		if err != nil {
			return nil, err
		}
		source.Subpath = sp
	}

	source.URL = fmt.Sprintf("https://github.com/%s/%s", source.Owner, source.Repo)
	return source, nil
}

func parseGitHubURL(url string) (*SourceInfo, error) {
	source := &SourceInfo{Type: "github"}

	path := strings.TrimPrefix(url, "https://github.com/")

	// Check for branch in URL (tree/branch/...)
	if before, after, ok := strings.Cut(path, "/tree/"); ok {
		beforeTree := before
		afterTree := after // Skip "/tree/"

		parts := strings.Split(beforeTree, "/")
		if len(parts) >= 2 {
			source.Owner = parts[0]
			source.Repo = strings.TrimSuffix(parts[1], ".git")
		}

		afterParts := strings.SplitN(afterTree, "/", 2)
		source.Ref = afterParts[0]
		if len(afterParts) > 1 {
			sp, err := SanitizeSubpath(afterParts[1])
			if err != nil {
				return nil, err
			}
			source.Subpath = sp
		}
	} else {
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			source.Owner = parts[0]
			source.Repo = strings.TrimSuffix(parts[1], ".git")
		}
		if len(parts) > 2 {
			sp, err := SanitizeSubpath(strings.Join(parts[2:], "/"))
			if err != nil {
				return nil, err
			}
			source.Subpath = sp
		}
	}

	if source.Owner != "" && source.Repo != "" {
		source.URL = repoHTTPURL("https", "github.com", source.Owner, source.Repo)
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

// parseSSHURL parses SSH git URLs.
// Supported: git@host:owner/repo.git, user@host:owner/repo.git, ssh://git@host/owner/repo.git.
func parseSSHURL(arg string) (*SourceInfo, error) {
	user, host, path, ok := splitSSHURL(arg)
	if !ok {
		return nil, fmt.Errorf("invalid SSH URL: expected user@host:owner/repo format")
	}

	ref := ""
	if before, after, ok := strings.Cut(path, "#"); ok {
		path = before
		ref = after
	}

	hadGitSuffix := strings.HasSuffix(path, ".git")
	path = strings.TrimSuffix(path, ".git")
	path = strings.Trim(path, "/")

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid SSH URL: expected user@host:owner/repo format")
	}

	source := &SourceInfo{Ref: ref}

	// Determine source type from host
	switch hostNameOnly(host) {
	case "github.com":
		source.Type = "github"
		source.Owner = parts[0]
		source.Repo = parts[1]
		source.URL = sshRepoURL(arg, user, host, parts[:2], hadGitSuffix)
		if len(parts) > 2 {
			sp, err := SanitizeSubpath(strings.Join(parts[2:], "/"))
			if err != nil {
				return nil, err
			}
			source.Subpath = sp
		}
	case "gitlab.com":
		// GitLab supports nested groups: group/subgroup/project
		source.Type = "gitlab"
		source.Repo = parts[len(parts)-1]
		source.Owner = strings.Join(parts[:len(parts)-1], "/")
		source.URL = sshRepoURL(arg, user, host, parts, hadGitSuffix)
	case "bitbucket.org":
		source.Type = "bitbucket"
		source.Owner = parts[0]
		source.Repo = parts[1]
		source.URL = sshRepoURL(arg, user, host, parts[:2], hadGitSuffix)
		if len(parts) > 2 {
			sp, err := SanitizeSubpath(strings.Join(parts[2:], "/"))
			if err != nil {
				return nil, err
			}
			source.Subpath = sp
		}
	default:
		source.Type = "git"
		if isGitHubEnterpriseHost(host) {
			source.Owner = parts[0]
			source.Repo = parts[1]
			source.URL = sshRepoURL(arg, user, host, parts[:2], hadGitSuffix)
			if len(parts) > 2 {
				sp, err := SanitizeSubpath(strings.Join(parts[2:], "/"))
				if err != nil {
					return nil, err
				}
				source.Subpath = sp
			}
		} else {
			// Self-hosted git instance — last component is repo, rest is owner/group.
			source.Repo = parts[len(parts)-1]
			source.Owner = strings.Join(parts[:len(parts)-1], "/")
			source.URL = sshRepoURL(arg, user, host, parts, hadGitSuffix)
		}
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

// parseGenericGitURL parses HTTP(S) git URLs from self-hosted instances.
// Formats: https://host/owner/repo(.git), https://host:port/group/subgroup/repo(.git)
func parseGenericGitURL(rawURL string) (*SourceInfo, error) {
	u, err := neturl.Parse(rawURL)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("invalid git URL: %s", rawURL)
	}

	parts := pathSegments(u.Path)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid git URL: no path after host")
	}

	source := &SourceInfo{Type: "git"}
	if u.Fragment != "" {
		source.Ref = u.Fragment
	}

	if isGitHubEnterpriseHost(u.Host) {
		source.Owner = parts[0]
		source.Repo = strings.TrimSuffix(parts[1], ".git")
		source.URL = repoHTTPURL(u.Scheme, u.Host, source.Owner, source.Repo)
		if len(parts) > 2 {
			if parts[2] == "tree" && len(parts) >= 4 {
				source.Ref = parts[3]
				if len(parts) > 4 {
					sp, err := SanitizeSubpath(strings.Join(parts[4:], "/"))
					if err != nil {
						return nil, err
					}
					source.Subpath = sp
				}
			} else {
				sp, err := SanitizeSubpath(strings.Join(parts[2:], "/"))
				if err != nil {
					return nil, err
				}
				source.Subpath = sp
			}
		}
		return source, nil
	}

	// Last component is repo (strip .git), everything before is owner/group.
	source.Repo = strings.TrimSuffix(parts[len(parts)-1], ".git")
	source.Owner = strings.Join(parts[:len(parts)-1], "/")
	source.URL = repoHTTPURL(u.Scheme, u.Host, append(parts[:len(parts)-1], source.Repo)...)

	return source, nil
}

func pathSegments(path string) []string {
	raw := strings.Split(strings.Trim(path, "/"), "/")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func repoHTTPURL(scheme, host string, parts ...string) string {
	if scheme == "" {
		scheme = "https"
	}
	return scheme + "://" + host + "/" + strings.Join(parts, "/")
}

func sshRepoURL(original, user, host string, parts []string, hadGitSuffix bool) string {
	path := strings.Join(parts, "/")
	if hadGitSuffix {
		path += ".git"
	}
	if strings.HasPrefix(original, "ssh://") {
		if user != "" {
			return "ssh://" + user + "@" + host + "/" + path
		}
		return "ssh://" + host + "/" + path
	}
	return user + "@" + host + ":" + path
}

func isGitHubEnterpriseHost(host string) bool {
	h := strings.ToLower(hostNameOnly(host))
	return strings.Contains(h, "github") || strings.Contains(h, "ghe")
}

func hostNameOnly(host string) string {
	if before, _, ok := strings.Cut(host, ":"); ok {
		return before
	}
	return host
}
