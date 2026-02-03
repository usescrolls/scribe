package cli

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	scribe "github.com/usescrolls/scribe/internal"
)

var (
	// Install flags
	installAgents   string
	installSkills   string
	installListOnly bool
	installYes      bool
	installAll      bool

	installCmd = &cobra.Command{
		Use:   "install <source>",
		Short: "Install skills from a source",
		Long: `Install skills from various sources.

All skills are installed globally and managed by Scribe across all detected
coding agents.

Sources can be:
  owner/repo                    GitHub shorthand
  https://github.com/owner/repo Full GitHub URL
  ./local/path                  Local directory
  https://example.com           Well-known endpoint

Examples:
  scribe install vercel-labs/agent-skills
  scribe install https://github.com/owner/repo
  scribe install ./my-skills
  scribe install owner/repo --agent claude-code,cursor
  scribe install owner/repo --list`,
		Args: cobra.ExactArgs(1),
		RunE: runInstall,
	}
)

func init() {
	installCmd.Flags().StringVarP(&installAgents, "agent", "a", "", "Target specific agents (comma-separated)")
	installCmd.Flags().StringVarP(&installSkills, "skill", "s", "", "Select specific skills to install (comma-separated)")
	installCmd.Flags().BoolVarP(&installListOnly, "list", "l", false, "List available skills without installing")
	installCmd.Flags().BoolVarP(&installYes, "yes", "y", false, "Skip interactive prompts")
	installCmd.Flags().BoolVar(&installAll, "all", false, "Install all skills to all detected agents")
}

func runInstall(cmd *cobra.Command, args []string) error {
	sourceArg := args[0]

	// Parse the source
	source, err := parseSource(sourceArg)
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	if !quiet {
		fmt.Printf("Fetching skills from %s...\n", formatSourceInfo(source))
	}

	// Fetch and discover skills
	skills, tempDir, err := fetchAndDiscoverSkills(source)
	if err != nil {
		return fmt.Errorf("failed to fetch skills: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp directory

	if len(skills) == 0 {
		return fmt.Errorf("no skills found in source")
	}

	// List only mode
	if installListOnly {
		fmt.Printf("\nFound %d skill(s):\n", len(skills))
		for _, skill := range skills {
			fmt.Printf("  - %s - %s\n", skill.Name, skill.Description)
		}
		return nil
	}

	// Filter skills if --skill flag is provided
	if installSkills != "" {
		skillNames := strings.Split(installSkills, ",")
		skills = filterSkills(skills, skillNames)
		if len(skills) == 0 {
			return fmt.Errorf("no matching skills found for: %s", installSkills)
		}
	}

	// Parse agent filter
	var targetAgents []string
	if installAgents != "" {
		targetAgents = strings.Split(installAgents, ",")
	}

	// Detect installed agents
	installedAgents := scribe.DetectInstalledAgents()
	if len(installedAgents) == 0 {
		return fmt.Errorf("no coding agents detected. Please install at least one agent (Claude Code, Cursor, etc.)")
	}

	if !quiet {
		fmt.Printf("\nFound %d skill(s) to install:\n", len(skills))
		for _, skill := range skills {
			fmt.Printf("  - %s - %s\n", skill.Name, skill.Description)
		}

		fmt.Printf("\nDetected %d agent(s):\n", len(installedAgents))
		for _, agent := range installedAgents {
			fmt.Printf("  - %s\n", agent.DisplayName)
		}
		fmt.Println()
	}

	// Ensure Scribe directories exist
	if err := scribe.EnsureScribeDirs(); err != nil {
		return fmt.Errorf("failed to create Scribe directories: %w", err)
	}

	// Ensure default workspace exists
	if err := scribe.EnsureDefaultWorkspace(); err != nil {
		scribe.Logger.Warn("failed to ensure default workspace", "error", err)
	}

	// Install each skill
	opts := scribe.InstallOptions{
		Agents: targetAgents,
		Yes:    installYes,
	}

	successCount := 0
	for _, skill := range skills {
		if !quiet {
			fmt.Printf("Installing %s...\n", skill.Name)
		}

		if err := scribe.InstallSkill(skill, source, opts); err != nil {
			fmt.Fprintf(os.Stderr, "  x Failed to install %s: %v\n", skill.Name, err)
			continue
		}

		// Add to workspaces
		if err := scribe.AddSkillToActiveAndDefaultWorkspace(skill.Name); err != nil {
			scribe.Logger.Warn("failed to add to workspace", "skill", skill.Name, "error", err)
		}

		if !quiet {
			fmt.Printf("  Installed %s\n", skill.Name)
		}
		successCount++
	}

	if !quiet {
		fmt.Printf("\nInstalled %d/%d skill(s)\n", successCount, len(skills))
	}

	return nil
}

// parseSource parses a source argument into a SourceInfo
func parseSource(arg string) (*scribe.SourceInfo, error) {
	source := &scribe.SourceInfo{}

	// Check for local path
	if strings.HasPrefix(arg, "./") || strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, "~") {
		absPath, err := filepath.Abs(arg)
		if err != nil {
			return nil, err
		}
		source.Type = "local"
		source.LocalPath = absPath
		return source, nil
	}

	// Check for GitHub URL
	if strings.HasPrefix(arg, "https://github.com/") {
		return parseGitHubURL(arg)
	}

	// Check for GitLab URL
	if strings.HasPrefix(arg, "https://gitlab.com/") {
		return parseGitLabURL(arg)
	}

	// Check for generic URL
	if strings.HasPrefix(arg, "https://") || strings.HasPrefix(arg, "http://") {
		// Check if it's a zip URL
		if strings.HasSuffix(arg, ".zip") {
			source.Type = "zip"
			source.URL = arg
			return source, nil
		}
		// Treat as well-known URL
		source.Type = "well-known"
		source.URL = arg
		return source, nil
	}

	// Assume GitHub shorthand: owner/repo or owner/repo#branch or owner/repo/path
	return parseGitHubShorthand(arg)
}

// parseGitHubShorthand parses formats like:
// owner/repo
// owner/repo#branch
// owner/repo/path/to/skills
// owner/repo/path#branch
func parseGitHubShorthand(arg string) (*scribe.SourceInfo, error) {
	source := &scribe.SourceInfo{Type: "github"}

	// Check for branch/ref
	if idx := strings.Index(arg, "#"); idx != -1 {
		source.Ref = arg[idx+1:]
		arg = arg[:idx]
	}

	parts := strings.Split(arg, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid GitHub shorthand: expected owner/repo")
	}

	source.Owner = parts[0]
	source.Repo = parts[1]

	// Check for subpath
	if len(parts) > 2 {
		source.Subpath = strings.Join(parts[2:], "/")
	}

	source.URL = fmt.Sprintf("https://github.com/%s/%s", source.Owner, source.Repo)

	return source, nil
}

// parseGitHubURL parses a full GitHub URL
func parseGitHubURL(url string) (*scribe.SourceInfo, error) {
	source := &scribe.SourceInfo{Type: "github", URL: url}

	// Remove https://github.com/ prefix
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

		// afterTree is branch/path or just branch
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

// parseGitLabURL parses a GitLab URL
func parseGitLabURL(url string) (*scribe.SourceInfo, error) {
	source := &scribe.SourceInfo{Type: "gitlab", URL: url}

	path := strings.TrimPrefix(url, "https://gitlab.com/")
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		source.Owner = parts[0]
		source.Repo = strings.TrimSuffix(parts[1], ".git")
	}

	return source, nil
}

// fetchAndDiscoverSkills fetches content from source and discovers skills
// Returns the skills and a temp directory that should be cleaned up by caller
func fetchAndDiscoverSkills(source *scribe.SourceInfo) ([]*scribe.Skill, string, error) {
	var skillsDir string
	var tempDir string

	switch source.Type {
	case "local":
		skillsDir = source.LocalPath
		if source.Subpath != "" {
			skillsDir = filepath.Join(skillsDir, source.Subpath)
		}
	case "github", "gitlab":
		// Clone repository to temp directory
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
	skills, err := scribe.DiscoverSkills(skillsDir)
	if err != nil {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
		return nil, "", err
	}

	return skills, tempDir, nil
}

// cloneRepository clones a git repository to a temp directory
func cloneRepository(source *scribe.SourceInfo) (string, error) {
	tempDir, err := os.MkdirTemp("", "scribe-clone-*")
	if err != nil {
		return "", err
	}

	// Build clone URL
	cloneURL := source.URL
	if !strings.HasSuffix(cloneURL, ".git") {
		cloneURL += ".git"
	}

	if err := cloneWithGit(cloneURL, tempDir, source.Ref); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	return tempDir, nil
}

// downloadAndExtractZip downloads and extracts a zip file to a temp directory
func downloadAndExtractZip(zipURL string) (string, error) {
	// Create temp directory for extraction
	tempDir, err := os.MkdirTemp("", "scribe-zip-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Download the zip file
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

	// Create a temporary file for the zip
	tmpFile, err := os.CreateTemp("", "scribe-download-*.zip")
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Copy the response body to the temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to save zip: %w", err)
	}
	tmpFile.Close()

	// Open the zip file
	zipReader, err := zip.OpenReader(tmpPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer zipReader.Close()

	// Check if all files share a common root directory
	commonRoot := findCommonRoot(zipReader.File)

	// Extract files
	for _, file := range zipReader.File {
		filePath := file.Name

		// Strip common root directory if present
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

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract file
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

	// Find the first directory component of the first file
	var commonRoot string
	for _, file := range files {
		parts := strings.SplitN(file.Name, "/", 2)
		if len(parts) > 1 {
			if commonRoot == "" {
				commonRoot = parts[0] + "/"
			} else if parts[0]+"/" != commonRoot {
				// Files don't share a common root
				return ""
			}
		} else if !file.FileInfo().IsDir() {
			// File at root level without directory
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

// filterSkills filters skills by name
func filterSkills(skills []*scribe.Skill, names []string) []*scribe.Skill {
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[strings.TrimSpace(n)] = true
	}

	var filtered []*scribe.Skill
	for _, skill := range skills {
		if nameSet[skill.Name] {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

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

// cloneWithGit clones a repository using the git command
func cloneWithGit(cloneURL, targetDir, ref string) error {
	args := []string{"clone", "--depth", "1"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, cloneURL, targetDir)

	cmd := exec.Command("git", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w\n%s", err, string(output))
	}
	return nil
}
