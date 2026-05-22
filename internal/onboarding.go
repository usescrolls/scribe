package scribe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
)

// DemoSkillContent is the content of the demo skill installed during onboarding
const DemoSkillContent = `---
name: scribe-welcome
description: A welcome skill that introduces Scribe and demonstrates skill formatting
---

# Welcome to Scribe!

This is a demo skill installed during Scribe setup. It demonstrates how skills work.

## What is Scribe?

Scribe is a skill distribution tool that syncs AI coding skills to all your coding agents.
Install a skill once, and it's automatically available in Claude Code, Cursor, GitHub Copilot,
and 40+ other agents.

## How Skills Work

Skills are markdown files with YAML frontmatter that provide context and instructions to
AI coding agents. This skill is now available in all your detected agents.

## Next Steps

1. Visit AgentHub to discover more skills
2. Use ` + "`scribe install <github-repo>`" + ` to install skills from GitHub
3. Create your own skills by adding SKILL.md files

You can safely uninstall this demo skill anytime with ` + "`scribe uninstall scribe-welcome`" + `.
`

// AreTermsAccepted checks if the user has accepted the current version of the terms.
// Returns false if terms were never accepted or if a newer version is available.
func AreTermsAccepted() (bool, error) {
	config, err := LoadConfig()
	if err != nil {
		return false, err
	}
	return config.TermsAcceptedVersion >= CurrentTermsVersion, nil
}

// AcceptTerms records the user's acceptance of the current terms version.
func AcceptTerms() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	config.TermsAcceptedAt = time.Now().Format(time.RFC3339)
	config.TermsAcceptedVersion = CurrentTermsVersion
	return SaveConfig(config)
}

// IsOnboardingCompleted checks if onboarding has been completed
func IsOnboardingCompleted() (bool, error) {
	config, err := LoadConfig()
	if err != nil {
		return false, err
	}
	return config.OnboardingCompleted, nil
}

// CompleteOnboarding marks onboarding as completed
func CompleteOnboarding() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	config.OnboardingCompleted = true
	return SaveConfig(config)
}

// DetectExistingSkills scans all detected agents for existing skills
func DetectExistingSkills() ([]ExistingSkillInfo, error) {
	agents := DetectInstalledAgents()
	var skills []ExistingSkillInfo

	for _, agent := range agents {
		skillsDir := expandPath(agent.GlobalSkillsDir)
		if !dirExists(skillsDir) {
			continue
		}

		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			Logger.Warn("failed to read agent skills dir", "agent", agent.ID, "error", err)
			continue
		}

		// Check if the skills directory itself is a git repo (monorepo pattern:
		// user cloned a repo of skills directly as the agent's skills dir)
		parentIsGitRepo := dirExists(filepath.Join(skillsDir, ".git"))

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			skillPath := filepath.Join(skillsDir, entry.Name())
			skillMdPath := filepath.Join(skillPath, SkillFileName)

			// Check if SKILL.md exists
			if _, err := os.Stat(skillMdPath); err != nil {
				continue
			}

			// Check if it's a git repo (either the skill dir itself or its parent)
			gitPath := filepath.Join(skillPath, ".git")
			isGitRepo := dirExists(gitPath) || parentIsGitRepo

			skills = append(skills, ExistingSkillInfo{
				Name:      entry.Name(),
				Path:      skillPath,
				AgentID:   agent.ID,
				AgentName: agent.DisplayName,
				IsGitRepo: isGitRepo,
			})
		}
	}

	return skills, nil
}

// DetectSkillConflicts finds skills with the same name in different agent directories
func DetectSkillConflicts(skills []ExistingSkillInfo) []SkillConflict {
	// Group skills by name
	byName := make(map[string][]ExistingSkillInfo)
	for _, skill := range skills {
		byName[skill.Name] = append(byName[skill.Name], skill)
	}

	// Find conflicts (skills with same name in multiple locations)
	var conflicts []SkillConflict
	for name, sources := range byName {
		if len(sources) > 1 {
			conflicts = append(conflicts, SkillConflict{
				Name:    name,
				Sources: sources,
			})
		}
	}

	return conflicts
}

// ImportExistingSkills imports skills from agent directories to ~/.scribe/scrolls/
// If a skill name appears in multiple agents, only the first occurrence is imported
func ImportExistingSkills(skills []ExistingSkillInfo) error {
	scrollsDir, err := GetScrollsDir()
	if err != nil {
		return err
	}

	// Track which skills we've already imported to avoid duplicates
	imported := make(map[string]bool)

	for _, skill := range skills {
		if IsSystemSkill(skill.Name) {
			Logger.Info("skipping system skill import", "name", skill.Name, "from", skill.AgentID)
			imported[skill.Name] = true
			continue
		}

		if imported[skill.Name] {
			Logger.Info("skipping duplicate skill", "name", skill.Name, "from", skill.AgentID)
			continue
		}

		destDir := filepath.Join(scrollsDir, skill.Name)

		// Extract git source info BEFORE moving (move may alter .git paths)
		meta := buildSkillMeta(skill)

		// Move the skill directory to scrolls
		if err := moveDir(skill.Path, destDir); err != nil {
			Logger.Error("failed to move skill", "name", skill.Name, "error", err)
			return fmt.Errorf("failed to import skill %s: %w", skill.Name, err)
		}

		// Write metadata file
		metaPath := filepath.Join(destDir, MetaFileName)
		if err := WriteSkillMeta(metaPath, meta); err != nil {
			Logger.Warn("failed to write skill meta", "name", skill.Name, "error", err)
		}

		imported[skill.Name] = true
		Logger.Info("imported skill", "name", skill.Name, "from", skill.AgentID)
	}

	// Sync all imported skills to all agents
	return SyncAllSkillsToAgents()
}

// ImportSelectedSkills imports only the specified skills (by path)
func ImportSelectedSkills(skillPaths []string) error {
	// Build a map of path to skill info
	allSkills, err := DetectExistingSkills()
	if err != nil {
		return err
	}

	pathToSkill := make(map[string]ExistingSkillInfo)
	for _, skill := range allSkills {
		pathToSkill[skill.Path] = skill
	}

	// Filter to only selected paths
	var selectedSkills []ExistingSkillInfo
	for _, path := range skillPaths {
		if skill, ok := pathToSkill[path]; ok {
			selectedSkills = append(selectedSkills, skill)
		}
	}

	return ImportExistingSkills(selectedSkills)
}

// DeleteExistingSkills removes all skills found in agent directories
func DeleteExistingSkills(skills []ExistingSkillInfo) error {
	for _, skill := range skills {
		if IsSystemSkill(skill.Name) {
			Logger.Info("skipping system skill delete", "name", skill.Name, "from", skill.AgentID)
			continue
		}

		if err := os.RemoveAll(skill.Path); err != nil {
			Logger.Error("failed to delete skill", "name", skill.Name, "path", skill.Path, "error", err)
			return fmt.Errorf("failed to delete skill %s: %w", skill.Name, err)
		}
		Logger.Info("deleted skill", "name", skill.Name, "from", skill.AgentID)
	}
	return nil
}

// InstallDemoSkill installs the scribe-welcome demo skill
func InstallDemoSkill() error {
	skillName := "scribe-welcome"

	// Check if already installed
	exists, err := SkillExists(skillName)
	if err != nil {
		return err
	}
	if exists {
		Logger.Info("demo skill already installed", "name", skillName)
		return nil
	}

	// Create skill directory
	skillDir, err := GetSkillDir(skillName)
	if err != nil {
		return err
	}
	if err := EnsureDir(skillDir); err != nil {
		return err
	}

	// Write SKILL.md
	skillPath := filepath.Join(skillDir, SkillFileName)
	if err := os.WriteFile(skillPath, []byte(DemoSkillContent), 0o644); err != nil {
		return err
	}

	// Write metadata
	metaPath := filepath.Join(skillDir, MetaFileName)
	meta := &SkillMeta{
		Source:      "scribe",
		SourceType:  "builtin",
		ContentHash: ComputeContentHash(DemoSkillContent),
		InstalledAt: time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}
	if err := WriteSkillMeta(metaPath, meta); err != nil {
		Logger.Warn("failed to write demo skill meta", "error", err)
	}

	// Add to default workspace
	if err := AddSkillToWorkspace(skillName, DefaultWorkspaceName); err != nil {
		Logger.Warn("failed to add demo skill to default workspace", "error", err)
	}

	// Sync to all agents
	if err := SyncAllSkillsToAgents(); err != nil {
		return err
	}

	Logger.Info("installed demo skill", "name", skillName)
	return nil
}

// extractGitRemoteURL reads the "origin" remote URL from a git repository.
// Returns empty string if the repo can't be opened or has no origin remote.
// Also checks the parent directory to support monorepo layouts where skills
// are subdirectories within a single cloned repo.
func extractGitRemoteURL(repoPath string) string {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		// Try parent directory (monorepo: skillsDir is the repo root)
		repo, err = git.PlainOpen(filepath.Dir(repoPath))
		if err != nil {
			return ""
		}
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return ""
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return ""
	}

	return urls[0]
}

// parseGitRemoteURL parses a git remote URL (HTTPS or SSH) into a SourceInfo.
// Returns nil if the URL is empty, malformed, or for an unsupported host.
// Supported hosts: github.com, gitlab.com
func parseGitRemoteURL(remoteURL string) *SourceInfo {
	if remoteURL == "" {
		return nil
	}

	var host, ownerRepo string

	// Handle various remote URL formats
	switch {
	case strings.HasPrefix(remoteURL, "git@"):
		// SSH format: git@host:owner/repo.git
		rest := strings.TrimPrefix(remoteURL, "git@")
		host, ownerRepo, _ = strings.Cut(rest, ":")
	case strings.HasPrefix(remoteURL, "https://") || strings.HasPrefix(remoteURL, "http://"):
		// HTTPS format: https://host/owner/repo.git
		trimmed := strings.TrimPrefix(remoteURL, "https://")
		trimmed = strings.TrimPrefix(trimmed, "http://")
		host, ownerRepo, _ = strings.Cut(trimmed, "/")
	default:
		return nil
	}

	// Determine source type from host
	var sourceType string
	switch host {
	case "github.com":
		sourceType = "github"
	case "gitlab.com":
		sourceType = "gitlab"
	case "bitbucket.org":
		sourceType = "bitbucket"
	default:
		sourceType = "git"
	}

	// Strip .git suffix and parse owner/repo
	ownerRepo = strings.TrimSuffix(ownerRepo, ".git")
	parts := strings.Split(ownerRepo, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return nil
	}

	var owner, repo string
	switch sourceType {
	case "gitlab", "git":
		// GitLab and self-hosted support nested groups: last component is repo
		repo = parts[len(parts)-1]
		owner = strings.Join(parts[:len(parts)-1], "/")
	default:
		owner = parts[0]
		repo = parts[1]
	}

	return &SourceInfo{
		Type:  sourceType,
		Owner: owner,
		Repo:  repo,
		URL:   fmt.Sprintf("https://%s/%s/%s", host, owner, repo),
	}
}

// buildSkillMeta creates a SkillMeta for an imported skill.
// For git-tracked skills, it attempts to extract source info from the git remote.
// Falls back to "local" if extraction fails or skill is not git-tracked.
func buildSkillMeta(skill ExistingSkillInfo) *SkillMeta {
	now := time.Now().Format(time.RFC3339)

	if skill.IsGitRepo {
		remoteURL := extractGitRemoteURL(skill.Path)
		if source := parseGitRemoteURL(remoteURL); source != nil {
			return &SkillMeta{
				Source:      formatSource(source),
				SourceType:  source.Type,
				SourceURL:   source.URL,
				ContentHash: "",
				InstalledAt: now,
				UpdatedAt:   now,
			}
		}
	}

	return &SkillMeta{
		Source:      "local",
		SourceType:  "local",
		ContentHash: "",
		InstalledAt: now,
		UpdatedAt:   now,
	}
}

// moveDir moves a directory from src to dest, preserving git history if present
func moveDir(src, dest string) error {
	// First try os.Rename which is fastest and preserves everything
	if err := os.Rename(src, dest); err == nil {
		return nil
	}

	// If rename fails (cross-device), fall back to copy + delete
	// Uses CopySkillDir from installer.go
	if err := CopySkillDir(src, dest); err != nil {
		return err
	}

	return os.RemoveAll(src)
}
