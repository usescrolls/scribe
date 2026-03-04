package scribe

import (
	"net/url"
	"path/filepath"
	"strings"
)

const (
	// QualifiedSeparator separates the source qualifier from the skill name.
	// Safe because SanitizeName collapses "--" in frontmatter names, so this
	// separator can never appear in a regular (unqualified) skill name.
	QualifiedSeparator = "--"
)

// SourceQualifier derives a sanitized source qualifier from a SourceInfo.
// Examples: "alice-skills" from owner=alice, repo=skills.
func SourceQualifier(source *SourceInfo) string {
	switch source.Type {
	case "github", "gitlab", "bitbucket", "git":
		if source.Owner != "" && source.Repo != "" {
			return SanitizeName(source.Owner + "-" + source.Repo)
		}
		return ""
	case "local":
		if source.LocalPath != "" {
			return SanitizeName(filepath.Base(source.LocalPath))
		}
		return ""
	default:
		if source.URL != "" {
			return qualifierFromURL(source.URL)
		}
		return ""
	}
}

// SourceQualifierFromMeta derives a source qualifier from existing skill metadata.
func SourceQualifierFromMeta(meta *SkillMeta) string {
	if meta == nil {
		return ""
	}
	source := ReconstructSource(meta)
	return SourceQualifier(source)
}

// QualifiedName constructs a source-qualified storage name.
// Format: "<qualifier>--<skillName>", e.g. "alice-skills--commit".
func QualifiedName(qualifier, skillName string) string {
	if qualifier == "" {
		return skillName
	}
	return qualifier + QualifiedSeparator + skillName
}

// IsQualifiedName checks whether a storage name contains the qualified separator.
func IsQualifiedName(name string) bool {
	return strings.Contains(name, QualifiedSeparator)
}

// qualifierFromURL extracts a sanitized qualifier from a URL.
func qualifierFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return SanitizeName(rawURL)
	}
	// Use host + first two path segments at most
	host := strings.TrimPrefix(u.Host, "www.")
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	segments := []string{host}
	for i, p := range parts {
		if i >= 2 {
			break
		}
		segments = append(segments, p)
	}
	return SanitizeName(strings.Join(segments, "-"))
}

// FrontmatterNameFromStorage extracts the simple skill name from a possibly
// qualified storage name. For "alice-skills--commit" it returns "commit".
// For "commit" (unqualified) it returns "commit" unchanged.
func FrontmatterNameFromStorage(storageName string) string {
	if idx := strings.LastIndex(storageName, QualifiedSeparator); idx >= 0 {
		return storageName[idx+len(QualifiedSeparator):]
	}
	return storageName
}

// GetFrontmatterName reads the SKILL.md for a stored skill and returns
// its sanitized frontmatter name (which may differ from the storage/directory name).
func GetFrontmatterName(storageName string) (string, error) {
	skill, err := ReadSkill(storageName)
	if err != nil {
		return "", err
	}
	return skill.Name, nil
}
