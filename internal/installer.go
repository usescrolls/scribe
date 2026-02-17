package scribe

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// InstallSkill installs a skill to the canonical location and syncs to agents
func InstallSkill(skill *Skill, source *SourceInfo, opts InstallOptions, gitInfo *GitCommitInfo, emit ...ProgressEmitter) error {
	// Determine target directory (always global)
	scrollsDir, err := GetScrollsDir()
	if err != nil {
		return fmt.Errorf("failed to get scrolls directory: %w", err)
	}

	skillDir := filepath.Join(scrollsDir, skill.Name)

	// Check if skill already exists
	if _, err := os.Stat(skillDir); err == nil {
		return fmt.Errorf("skill '%s' already exists", skill.Name)
	}

	// Copy skill to canonical location
	emitProgress(emit, "install", "copying", "Copying files...", "")
	if err := CopySkillDir(skill.Path, skillDir); err != nil {
		return fmt.Errorf("failed to copy skill: %w", err)
	}

	// Write metadata
	emitProgress(emit, "install", "metadata", "Writing metadata...", "")
	skillPathInSource := ""
	if source.Subpath != "" {
		skillPathInSource = source.Subpath
	}

	// Read the SKILL.md content for hashing
	skillContent, err := os.ReadFile(filepath.Join(skillDir, SkillFileName))
	if err != nil {
		return fmt.Errorf("failed to read skill content: %w", err)
	}

	meta := NewSkillMeta(source, skillPathInSource, string(skillContent), gitInfo)
	meta.IsPrivate = opts.IsPrivate
	metaPath := filepath.Join(skillDir, MetaFileName)
	if err := WriteSkillMeta(metaPath, meta); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Sync to all detected agents
	agents := AgentIDs(DetectInstalledAgents())
	if err := SyncSkillToAgents(skill.Name, agents, emit...); err != nil {
		return fmt.Errorf("failed to sync to agents: %w", err)
	}

	return nil
}

// UninstallSkill removes a skill from canonical storage and all agents
func UninstallSkill(skillName string) error {
	if IsSystemSkill(skillName) {
		return fmt.Errorf("cannot uninstall system skill '%s'", skillName)
	}

	// Remove symlinks from all agents first
	agentIDs := AgentIDs(DetectInstalledAgents())

	// Log but continue - we still want to remove the canonical copy
	_ = RemoveSkillFromAgents(skillName, agentIDs)

	// Remove from canonical location
	skillDir, err := GetSkillDir(skillName)
	if err != nil {
		return err
	}

	return os.RemoveAll(skillDir)
}

// SyncSkillToAgents creates symlinks for a skill in all specified agents' directories
func SyncSkillToAgents(skillName string, agentIDs []string, emit ...ProgressEmitter) error {
	scrollsDir, err := GetScrollsDir()
	if err != nil {
		return err
	}

	skillDir := filepath.Join(scrollsDir, skillName)

	for _, agentID := range agentIDs {
		agent := GetAgent(agentID)
		if agent == nil {
			continue
		}

		emitProgress(emit, "install", "syncing", fmt.Sprintf("Syncing to %s...", agent.DisplayName), agent.DisplayName)

		agentSkillsDir := expandPath(agent.GlobalSkillsDir)

		// Ensure agent skills directory exists
		if err := EnsureDir(agentSkillsDir); err != nil {
			continue
		}

		linkPath := filepath.Join(agentSkillsDir, skillName)

		// Remove existing link/directory if present
		_ = os.RemoveAll(linkPath)

		// Create symlink
		if err := CreateSymlink(skillDir, linkPath); err != nil {
			// Fall back to copy if symlink fails
			if err := CopySkillDir(skillDir, linkPath); err != nil {
				continue
			}
		}
	}

	return nil
}

// RemoveSkillFromAgents removes symlinks for a skill from all specified agents' directories
func RemoveSkillFromAgents(skillName string, agentIDs []string) error {
	for _, agentID := range agentIDs {
		agent := GetAgent(agentID)
		if agent == nil {
			continue
		}

		agentSkillsDir := expandPath(agent.GlobalSkillsDir)
		linkPath := filepath.Join(agentSkillsDir, skillName)

		// Remove the symlink or directory
		_ = os.RemoveAll(linkPath)
	}

	return nil
}

// CreateSymlink creates a symlink from target to link
// On Windows, it creates a junction for directories
func CreateSymlink(target, link string) error {
	// Use relative paths for portability
	relPath, err := filepath.Rel(filepath.Dir(link), target)
	if err != nil {
		// Fall back to absolute path
		relPath = target
	}

	// On Windows, use junction for directories
	if runtime.GOOS == "windows" {
		return createWindowsJunction(target, link)
	}

	return os.Symlink(relPath, link)
}

// createWindowsJunction creates a directory junction on Windows
func createWindowsJunction(target, link string) error {
	// On Windows, we need to use absolute paths for junctions
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	// Use mklink /J command via cmd
	// This requires the link path to not exist
	if err := os.RemoveAll(link); err != nil && !os.IsNotExist(err) {
		return err
	}

	// For simplicity, fall back to regular symlink
	// Windows 10+ with developer mode supports symlinks without admin
	return os.Symlink(absTarget, link)
}

// CopySkillDir copies a skill directory to a new location
func CopySkillDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath)
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// SyncAllSkillsToAgents syncs all installed skills to all detected agents
func SyncAllSkillsToAgents() error {
	skills, err := ListInstalledSkills()
	if err != nil {
		return err
	}

	agentIDs := AgentIDs(DetectInstalledAgents())

	for _, skillName := range skills {
		if err := SyncSkillToAgents(skillName, agentIDs); err != nil {
			// Log but continue with other skills
			continue
		}
	}

	return nil
}

// IsSymlink checks if a path is a symbolic link
func IsSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// GetSymlinkTarget returns the target of a symbolic link
func GetSymlinkTarget(path string) (string, error) {
	return os.Readlink(path)
}
