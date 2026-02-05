package scribe

import (
	"fmt"
	"net/url"
	"strings"
)

// InstallResult contains the result of a URL scheme installation
type InstallResult struct {
	Success      bool
	SkillsCount  int
	SkillNames   []string
	ErrorMessage string
}

// HandleInstallURL parses an agenthub:// URL and installs the skills
// URL format: agenthub://install?name=skill&source=github&repo=owner/repo
func HandleInstallURL(urlString string) *InstallResult {
	result := &InstallResult{}

	// Parse the URL
	source, skillFilter, err := ParseInstallURL(urlString)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to parse URL: %v", err)
		return result
	}

	Logger.Info("installing from URL scheme", "source", source.Type, "repo", source.Owner+"/"+source.Repo)

	// Fetch and discover skills
	skills, fetchResult, err := FetchAndDiscoverSkills(source)
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to fetch skills: %v", err)
		return result
	}

	if len(skills) == 0 {
		result.ErrorMessage = "No skills found in source"
		return result
	}

	// Filter skills if a specific skill was requested
	if skillFilter != "" {
		skills = filterSkillsByName(skills, skillFilter)
		if len(skills) == 0 {
			result.ErrorMessage = fmt.Sprintf("Skill '%s' not found in source", skillFilter)
			return result
		}
	}

	// Ensure directories exist
	if err := EnsureScribeDirs(); err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to create directories: %v", err)
		return result
	}

	// Ensure default workspace exists
	if err := EnsureDefaultWorkspace(); err != nil {
		Logger.Warn("failed to ensure default workspace", "error", err)
	}

	// Install each skill
	opts := InstallOptions{Yes: true} // Auto-confirm for URL scheme installs
	for _, skill := range skills {
		Logger.Info("installing skill", "name", skill.Name)

		if err := InstallSkill(skill, source, opts); err != nil {
			Logger.Error("failed to install skill", "name", skill.Name, "error", err)
			continue
		}

		// Add to workspaces
		if err := AddSkillToActiveAndDefaultWorkspace(skill.Name); err != nil {
			Logger.Warn("failed to add to workspace", "skill", skill.Name, "error", err)
		}

		result.SkillNames = append(result.SkillNames, skill.Name)
	}

	result.SkillsCount = len(result.SkillNames)
	result.Success = result.SkillsCount > 0

	if !result.Success && result.ErrorMessage == "" {
		result.ErrorMessage = "Failed to install any skills"
	}

	return result
}

// ParseInstallURL parses an agenthub:// URL into a SourceInfo
// Returns the source info and an optional skill name filter
func ParseInstallURL(urlString string) (*SourceInfo, string, error) {
	// Parse the URL
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, "", fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "agenthub" {
		return nil, "", fmt.Errorf("expected agenthub:// scheme, got %s://", u.Scheme)
	}

	// Parse query parameters
	query := u.Query()
	sourceType := query.Get("source")
	repo := query.Get("repo")
	skillName := query.Get("name")

	// Validate required parameters
	if repo == "" {
		return nil, "", fmt.Errorf("missing 'repo' parameter")
	}

	source := &SourceInfo{}

	// Determine source type
	switch sourceType {
	case "github", "":
		source.Type = "github"
		parts := strings.Split(repo, "/")
		if len(parts) < 2 {
			return nil, "", fmt.Errorf("invalid repo format, expected owner/repo")
		}
		source.Owner = parts[0]
		source.Repo = parts[1]
		if len(parts) > 2 {
			source.Subpath = strings.Join(parts[2:], "/")
		}
		source.URL = fmt.Sprintf("https://github.com/%s/%s", source.Owner, source.Repo)

	case "gitlab":
		source.Type = "gitlab"
		parts := strings.Split(repo, "/")
		if len(parts) < 2 {
			return nil, "", fmt.Errorf("invalid repo format, expected owner/repo")
		}
		source.Owner = parts[0]
		source.Repo = parts[1]
		source.URL = fmt.Sprintf("https://gitlab.com/%s/%s", source.Owner, source.Repo)

	case "url", "zip":
		source.Type = "zip"
		source.URL = repo // In this case, repo contains the full URL

	default:
		return nil, "", fmt.Errorf("unsupported source type: %s", sourceType)
	}

	// Check for branch/ref parameter
	if ref := query.Get("ref"); ref != "" {
		source.Ref = ref
	}

	return source, skillName, nil
}

// filterSkillsByName filters skills to only include the specified name
func filterSkillsByName(skills []*Skill, name string) []*Skill {
	var filtered []*Skill
	for _, skill := range skills {
		if skill.Name == name {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

// FormatSource formats a SourceInfo for display
func FormatSource(source *SourceInfo) string {
	switch source.Type {
	case "github":
		s := source.Owner + "/" + source.Repo
		if source.Ref != "" {
			s += "#" + source.Ref
		}
		if source.Subpath != "" {
			s += "/" + source.Subpath
		}
		return "github:" + s
	case "gitlab":
		return "gitlab:" + source.Owner + "/" + source.Repo
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
