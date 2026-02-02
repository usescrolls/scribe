package scribe

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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
	skills, tempDir, err := FetchAndDiscoverSkills(source)
	if tempDir != "" {
		defer os.RemoveAll(tempDir)
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

// FetchAndDiscoverSkills fetches content from a source and discovers skills
// Returns the skills and a temp directory that should be cleaned up by caller
func FetchAndDiscoverSkills(source *SourceInfo) ([]*Skill, string, error) {
	var skillsDir string
	var tempDir string

	switch source.Type {
	case "local":
		skillsDir = source.LocalPath
		if source.Subpath != "" {
			skillsDir = filepath.Join(skillsDir, source.Subpath)
		}

	case "github", "gitlab":
		var err error
		tempDir, err = cloneRepository(source)
		if err != nil {
			return nil, "", err
		}
		skillsDir = tempDir
		if source.Subpath != "" {
			skillsDir = filepath.Join(tempDir, source.Subpath)
		}

	case "zip":
		var err error
		tempDir, err = downloadAndExtractZip(source.URL)
		if err != nil {
			return nil, "", err
		}
		skillsDir = tempDir

	case "well-known":
		return nil, "", fmt.Errorf("well-known sources not yet implemented")

	default:
		return nil, "", fmt.Errorf("unsupported source type: %s", source.Type)
	}

	// Discover skills in the directory
	skills, err := DiscoverSkills(skillsDir)
	if err != nil {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
		return nil, "", err
	}

	return skills, tempDir, nil
}

// cloneRepository clones a git repository to a temp directory
func cloneRepository(source *SourceInfo) (string, error) {
	tempDir, err := os.MkdirTemp("", "scribe-clone-*")
	if err != nil {
		return "", err
	}

	// Build clone URL
	cloneURL := source.URL
	if !strings.HasSuffix(cloneURL, ".git") {
		cloneURL += ".git"
	}

	args := []string{"clone", "--depth", "1"}
	if source.Ref != "" {
		args = append(args, "--branch", source.Ref)
	}
	args = append(args, cloneURL, tempDir)

	cmd := exec.Command("git", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("git clone failed: %w\n%s", err, string(output))
	}

	return tempDir, nil
}

// downloadAndExtractZip downloads and extracts a zip file to a temp directory
func downloadAndExtractZip(zipURL string) (string, error) {
	tempDir, err := os.MkdirTemp("", "scribe-zip-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	resp, err := http.Get(zipURL)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to download zip: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to download zip: status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "scribe-download-*.zip")
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to save zip: %w", err)
	}
	tmpFile.Close()

	zipReader, err := zip.OpenReader(tmpPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer zipReader.Close()

	commonRoot := findCommonRoot(zipReader.File)

	for _, file := range zipReader.File {
		filePath := file.Name

		if commonRoot != "" {
			filePath = strings.TrimPrefix(filePath, commonRoot)
			if filePath == "" {
				continue
			}
		}

		destPath := filepath.Join(tempDir, filePath)

		// Check for zip slip vulnerability
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(tempDir)+string(os.PathSeparator)) {
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("invalid file path in zip: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, file.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		if err := extractZipFile(file, destPath); err != nil {
			os.RemoveAll(tempDir)
			return "", err
		}
	}

	return tempDir, nil
}

// findCommonRoot checks if all files in a zip share a common root directory
func findCommonRoot(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}

	var commonRoot string
	for _, file := range files {
		parts := strings.SplitN(file.Name, "/", 2)
		if len(parts) > 1 {
			if commonRoot == "" {
				commonRoot = parts[0] + "/"
			} else if parts[0]+"/" != commonRoot {
				return ""
			}
		} else if !file.FileInfo().IsDir() {
			return ""
		}
	}

	return commonRoot
}

// extractZipFile extracts a single file from a zip archive
func extractZipFile(file *zip.File, destPath string) error {
	srcFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	return nil
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
