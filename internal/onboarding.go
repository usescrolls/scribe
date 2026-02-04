package scribe

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
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

			// Check if it's a git repo
			gitPath := filepath.Join(skillPath, ".git")
			isGitRepo := dirExists(gitPath)

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
		if imported[skill.Name] {
			Logger.Info("skipping duplicate skill", "name", skill.Name, "from", skill.AgentID)
			continue
		}

		destDir := filepath.Join(scrollsDir, skill.Name)

		// Move the skill directory to scrolls
		if err := moveDir(skill.Path, destDir); err != nil {
			Logger.Error("failed to move skill", "name", skill.Name, "error", err)
			return fmt.Errorf("failed to import skill %s: %w", skill.Name, err)
		}

		// Create metadata file
		metaPath := filepath.Join(destDir, MetaFileName)
		meta := &SkillMeta{
			Source:      "local",
			SourceType:  "local",
			ContentHash: "", // Will be computed if needed
			InstalledAt: time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		}
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

// moveDir moves a directory from src to dest, preserving git history if present
func moveDir(src, dest string) error {
	// First try os.Rename which is fastest and preserves everything
	if err := os.Rename(src, dest); err == nil {
		return nil
	}

	// If rename fails (cross-device), fall back to copy + delete
	// Uses copySkillDir from installer.go
	if err := copySkillDir(src, dest); err != nil {
		return err
	}

	return os.RemoveAll(src)
}
