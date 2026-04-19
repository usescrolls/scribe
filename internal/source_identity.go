package scribe

// sourceIdentity returns a stable key for a source that includes its type.
// This keeps repositories like github.com/owner/repo and gitlab.com/owner/repo
// distinct even when their human-readable source string is the same.
func sourceIdentity(sourceType, source string) string {
	if source == "" {
		return ""
	}
	if sourceType == "" {
		return source
	}
	return sourceType + ":" + source
}

func sourceIdentityFromSourceInfo(source *SourceInfo) string {
	if source == nil {
		return ""
	}
	switch source.Type {
	case "git":
		if source.URL != "" {
			return sourceIdentity(source.Type, source.URL)
		}
	case "local":
		return sourceIdentity(source.Type, source.LocalPath)
	case "url", "well-known", "zip":
		return sourceIdentity(source.Type, source.URL)
	}
	return sourceIdentity(source.Type, formatSource(source))
}

func sourceIdentityFromMeta(meta *SkillMeta) string {
	if meta == nil {
		return ""
	}
	switch meta.SourceType {
	case "git":
		if meta.SourceURL != "" {
			return sourceIdentity(meta.SourceType, meta.SourceURL)
		}
	case "local":
		return sourceIdentity(meta.SourceType, meta.Source)
	case "url", "well-known", "zip":
		if meta.SourceURL != "" {
			return sourceIdentity(meta.SourceType, meta.SourceURL)
		}
	}
	return sourceIdentity(meta.SourceType, meta.Source)
}
