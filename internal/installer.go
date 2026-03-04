package scribe

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallSkill installs a skill to the canonical location.
// It does NOT sync to agents — that is handled by workspace logic so that
// only skills in the active workspace appear in agent folders.
func InstallSkill(skill *Skill, source *SourceInfo, opts InstallOptions, gitInfo *GitCommitInfo, emit ...ProgressEmitter) error {
	// Determine target directory (always global)
	scrollsDir, err := GetScrollsDir()
	if err != nil {
		return fmt.Errorf("failed to get scrolls directory: %w", err)
	}

	skillDir := filepath.Join(scrollsDir, skill.Name)

	// Check if a directory with this exact storage name already exists
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

// installedSkillIndex maps frontmatter names to storage info for conflict detection.
type installedSkillIndex struct {
	fmToStorage   map[string][]string // lowercase frontmatter name → [storage names]
	storageSource map[string]string   // storage name → source string
}

// buildInstalledIndex reads all installed skills and builds lookup maps.
func buildInstalledIndex() *installedSkillIndex {
	installed, _ := ListInstalledSkills()
	idx := &installedSkillIndex{
		fmToStorage:   make(map[string][]string, len(installed)),
		storageSource: make(map[string]string, len(installed)),
	}
	for _, storageName := range installed {
		skill, err := ReadSkill(storageName)
		if err != nil {
			continue
		}
		fmKey := strings.ToLower(skill.Name)
		idx.fmToStorage[fmKey] = append(idx.fmToStorage[fmKey], storageName)
		if skill.Meta != nil {
			idx.storageSource[storageName] = skill.Meta.Source
		}
	}
	return idx
}

// FilterAlreadyInstalled partitions discovered skills into new, already-installed,
// and conflicts (same frontmatter name from a different source).
// Comparisons are case-insensitive.
func FilterAlreadyInstalled(skills []*Skill, source *SourceInfo) (newSkills []*Skill, alreadyInstalledNames []string, conflicts []*Skill) {
	idx := buildInstalledIndex()
	newSource := formatSource(source)

	for _, skill := range skills {
		fmKey := strings.ToLower(skill.Name)
		existingNames := idx.fmToStorage[fmKey]

		if len(existingNames) == 0 {
			newSkills = append(newSkills, skill)
			continue
		}

		if hasSourceMatch(idx.storageSource, existingNames, newSource) {
			alreadyInstalledNames = append(alreadyInstalledNames, skill.Name)
		} else {
			conflicts = append(conflicts, skill)
		}
	}
	return
}

// hasSourceMatch checks if any of the storage names have a matching source string.
func hasSourceMatch(storageSource map[string]string, storageNames []string, targetSource string) bool {
	for _, sn := range storageNames {
		if storageSource[sn] == targetSource {
			return true
		}
	}
	return false
}

// RenameInstalledSkill renames an already-installed skill's storage directory
// and updates all workspace references and agent symlinks.
func RenameInstalledSkill(oldName, newName string) error {
	scrollsDir, err := GetScrollsDir()
	if err != nil {
		return err
	}

	oldDir := filepath.Join(scrollsDir, oldName)
	newDir := filepath.Join(scrollsDir, newName)

	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		return fmt.Errorf("skill '%s' not found", oldName)
	}
	if _, err := os.Stat(newDir); err == nil {
		return fmt.Errorf("target name '%s' already exists", newName)
	}

	// Rename the directory
	if err := os.Rename(oldDir, newDir); err != nil {
		return fmt.Errorf("failed to rename skill directory: %w", err)
	}

	// Update all workspace references
	if err := renameSkillInAllWorkspaces(oldName, newName); err != nil {
		Logger.Warn("failed to update workspace references", "old", oldName, "new", newName, "error", err)
	}

	// Update agent symlinks
	agentIDs := AgentIDs(DetectInstalledAgents())
	_ = RemoveSkillFromAgents(oldName, agentIDs)
	_ = SyncSkillToAgents(newName, agentIDs)

	return nil
}

// renameSkillInAllWorkspaces replaces oldName with newName in all workspace definitions.
func renameSkillInAllWorkspaces(oldName, newName string) error {
	workspaces, err := ListWorkspaces()
	if err != nil {
		return err
	}

	for _, ws := range workspaces {
		changed := false
		for i, s := range ws.Skills {
			if strings.EqualFold(s, oldName) {
				ws.Skills[i] = newName
				changed = true
			}
		}
		if changed {
			if err := saveWorkspace(ws); err != nil {
				Logger.Warn("failed to save workspace after rename", "workspace", ws.Name, "error", err)
			}
		}
	}
	return nil
}

// HandleNameConflicts processes skills that have the same frontmatter name as
// existing installed skills from a different source. It renames the existing
// skill(s) to source-qualified names and returns the new skills with qualified
// names ready for installation.
func HandleNameConflicts(conflicts []*Skill, source *SourceInfo) ([]*Skill, error) {
	idx := buildInstalledIndex()
	newQualifier := SourceQualifier(source)

	var qualified []*Skill
	for _, skill := range conflicts {
		fmKey := strings.ToLower(skill.Name)
		existingNames := idx.fmToStorage[fmKey]

		// Rename each existing skill that still uses its simple (unqualified) name
		for _, storageName := range existingNames {
			if IsQualifiedName(storageName) {
				continue // Already qualified from a previous conflict
			}
			existingMeta, _ := ReadSkillMeta(mustGetMetaPath(storageName))
			existingQualifier := SourceQualifierFromMeta(existingMeta)
			qualifiedExisting := QualifiedName(existingQualifier, storageName)

			if err := RenameInstalledSkill(storageName, qualifiedExisting); err != nil {
				return nil, fmt.Errorf("failed to rename existing skill '%s' to '%s': %w", storageName, qualifiedExisting, err)
			}
			Logger.Info("renamed existing skill for conflict resolution", "old", storageName, "new", qualifiedExisting)
		}

		// Set the new skill's storage name to its qualified form
		qualifiedNew := QualifiedName(newQualifier, skill.Name)
		skill.Name = qualifiedNew
		qualified = append(qualified, skill)
	}

	return qualified, nil
}

// mustGetMetaPath returns the meta path or empty string on error.
func mustGetMetaPath(skillName string) string {
	p, _ := GetMetaPath(skillName)
	return p
}

// FilterAndResolveConflicts is the high-level entry point used by all install
// paths (CLI, Wails, URL scheme).  It filters out already-installed skills,
// detects frontmatter-name conflicts from different sources, resolves them
// (renaming existing skills + qualifying new ones), and returns the final
// list of skills ready for InstallSkill.
//
// Returns (toInstall, alreadyInstalledNames, error).
func FilterAndResolveConflicts(skills []*Skill, source *SourceInfo) ([]*Skill, []string, error) {
	newSkills, alreadyInstalled, conflicts := FilterAlreadyInstalled(skills, source)

	if len(conflicts) > 0 {
		qualified, err := HandleNameConflicts(conflicts, source)
		if err != nil {
			return newSkills, alreadyInstalled, err
		}
		newSkills = append(newSkills, qualified...)
	}

	return newSkills, alreadyInstalled, nil
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
