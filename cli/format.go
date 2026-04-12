package cli

import (
	"fmt"
	"strings"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

// formatSourceInfo formats a SourceInfo for display
func formatSourceInfo(source *scribe.SourceInfo) string {
	switch source.Type {
	case "github":
		s := source.Owner + "/" + source.Repo
		if source.Ref != "" {
			s += "#" + source.Ref
		}
		if source.Subpath != "" {
			s += "/" + source.Subpath
		}
		if strings.HasPrefix(source.URL, "git@") {
			return "github(ssh):" + s
		}
		return "github:" + s
	case "gitlab":
		s := source.Owner + "/" + source.Repo
		if strings.HasPrefix(source.URL, "git@") {
			return "gitlab(ssh):" + s
		}
		return "gitlab:" + s
	case "bitbucket":
		s := source.Owner + "/" + source.Repo
		if strings.HasPrefix(source.URL, "git@") {
			return "bitbucket(ssh):" + s
		}
		return "bitbucket:" + s
	case "local":
		return "local:" + source.LocalPath
	case "zip":
		return "zip:" + source.URL
	case "well-known":
		return source.URL
	default:
		return source.URL
	}
}

// formatSkillSource returns a human-readable source string for table display
func formatSkillSource(info scribe.SkillInfo) string {
	if info.Source == "" {
		return "-"
	}
	switch info.SourceType {
	case "github":
		return fmt.Sprintf("github:%s", info.Source)
	case "gitlab":
		return fmt.Sprintf("gitlab:%s", info.Source)
	case "local":
		return fmt.Sprintf("local:%s", info.Source)
	case "url", "zip":
		return info.Source
	default:
		return info.Source
	}
}

// truncateHash truncates a hash for display
func truncateHash(hash string) string {
	if hash == "" {
		return "-"
	}
	// Show first 20 chars of hash (sha256:abc123...)
	if len(hash) > 20 {
		return hash[:20] + "..."
	}
	return hash
}

// truncateString truncates a string to maxLen and adds "..." if needed
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatInstalledAt formats the installed timestamp for display
func formatInstalledAt(timestamp string) string {
	if timestamp == "" {
		return "-"
	}
	// Try to extract just the date part (YYYY-MM-DD)
	if len(timestamp) >= 10 {
		return timestamp[:10]
	}
	return timestamp
}
