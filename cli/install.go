package cli

import (
	"fmt"
	"os"
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
	skills, fetchResult, err := scribe.FetchAndDiscoverSkills(source)
	if err != nil {
		if scribe.IsAuthError(err) {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", scribe.AuthHintMessage())
		}
		return fmt.Errorf("failed to fetch skills: %w", err)
	}
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}

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

	// Extract git commit info from fetched repo
	gitInfo := scribe.GetHeadCommitInfo(fetchResult.ContentDir)

	// Install each skill
	opts := scribe.InstallOptions{
		Agents:    targetAgents,
		Yes:       installYes,
		IsPrivate: fetchResult.IsPrivate,
	}

	successCount := 0
	for _, skill := range skills {
		if !quiet {
			fmt.Printf("Installing %s...\n", skill.Name)
		}

		if err := scribe.InstallSkill(skill, source, opts, gitInfo); err != nil {
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

	// Check for SSH URL (git@host:owner/repo.git)
	if strings.HasPrefix(arg, "git@") {
		return scribe.ParseSourceString(arg)
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
