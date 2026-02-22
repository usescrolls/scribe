package scribe

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ListWorkspaces returns all workspace definitions
func ListWorkspaces() ([]*Workspace, error) {
	workspacesDir, err := GetWorkspacesDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(workspacesDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default workspace if directory doesn't exist
			return []*Workspace{createDefaultWorkspace()}, nil
		}
		return nil, err
	}

	var workspaces []*Workspace
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".json")
		ws, err := GetWorkspace(name)
		if err != nil {
			continue
		}
		workspaces = append(workspaces, ws)
	}

	// Ensure default workspace exists
	hasDefault := false
	for _, ws := range workspaces {
		if ws.Name == DefaultWorkspaceName {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		workspaces = append([]*Workspace{createDefaultWorkspace()}, workspaces...)
	}

	return workspaces, nil
}

// GetWorkspace reads a workspace definition by name
func GetWorkspace(name string) (*Workspace, error) {
	wsPath, err := GetWorkspacePath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(wsPath)
	if err != nil {
		if os.IsNotExist(err) && name == DefaultWorkspaceName {
			return createDefaultWorkspace(), nil
		}
		return nil, err
	}

	var ws Workspace
	if err := json.Unmarshal(data, &ws); err != nil {
		return nil, err
	}

	// Inject system skills (always present in every workspace)
	ws.Skills = injectSystemSkills(ws.Skills)

	return &ws, nil
}

// CreateWorkspace creates a new workspace definition
func CreateWorkspace(ws *Workspace) error {
	if ws.Name == "" {
		return fmt.Errorf("workspace name is required")
	}

	// Check if workspace already exists
	wsPath, err := GetWorkspacePath(ws.Name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(wsPath); err == nil {
		return fmt.Errorf("workspace '%s' already exists", ws.Name)
	}

	return saveWorkspace(ws)
}

// UpdateWorkspace updates an existing workspace definition
func UpdateWorkspace(ws *Workspace) error {
	return saveWorkspace(ws)
}

// DeleteWorkspace removes a workspace definition
func DeleteWorkspace(name string) error {
	if name == DefaultWorkspaceName {
		return fmt.Errorf("cannot delete the default workspace")
	}

	wsPath, err := GetWorkspacePath(name)
	if err != nil {
		return err
	}

	// Check if this is the active workspace
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	if config.ActiveWorkspace == name {
		// Switch to default before deleting
		if err := SetActiveWorkspace(DefaultWorkspaceName); err != nil {
			return err
		}
	}

	return os.Remove(wsPath)
}

// GetActiveWorkspace returns the currently active workspace
func GetActiveWorkspace() (*Workspace, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	return GetWorkspace(config.ActiveWorkspace)
}

// SetActiveWorkspace switches to a different workspace
// This updates symlinks to match the target workspace's skill set
func SetActiveWorkspace(name string) error {
	// Load target workspace
	targetWs, err := GetWorkspace(name)
	if err != nil {
		return fmt.Errorf("workspace '%s' not found: %w", name, err)
	}

	// Load current workspace
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	currentWs, err := GetWorkspace(config.ActiveWorkspace)
	if err != nil {
		// If current workspace doesn't exist, create empty one for comparison
		currentWs = &Workspace{Name: config.ActiveWorkspace, Skills: []string{}}
	}

	// Sync workspace (add/remove symlinks)
	if err := SyncWorkspace(currentWs, targetWs); err != nil {
		return err
	}

	// Update config
	config.ActiveWorkspace = name
	return SaveConfig(config)
}

// ResyncCurrentWorkspace forces a full resync of all skills in the current workspace
// This ensures all symlinks are correctly created for installed agents
func ResyncCurrentWorkspace() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	ws, err := GetWorkspace(config.ActiveWorkspace)
	if err != nil {
		return err
	}

	agentIDs := AgentIDs(DetectInstalledAgents())

	// Sync all skills in the workspace
	for _, skillName := range ws.Skills {
		exists, _ := SkillExists(skillName)
		if !exists {
			continue
		}
		_ = SyncSkillToAgents(skillName, agentIDs)
	}

	return nil
}

// SyncWorkspace updates agent symlinks to match the target workspace
func SyncWorkspace(current, target *Workspace) error {
	agentIDs := AgentIDs(DetectInstalledAgents())

	// Find skills to remove (in current but not in target)
	toRemove := skillDiff(current.Skills, target.Skills)

	// Find skills to add (in target but not in current)
	toAdd := skillDiff(target.Skills, current.Skills)

	// Remove symlinks for skills no longer in workspace
	for _, skillName := range toRemove {
		// Log but continue on error
		_ = RemoveSkillFromAgents(skillName, agentIDs)
	}

	// Add symlinks for new skills in workspace
	for _, skillName := range toAdd {
		// Verify skill exists in canonical location
		exists, _ := SkillExists(skillName)
		if !exists {
			continue
		}
		// Log but continue on error
		_ = SyncSkillToAgents(skillName, agentIDs)
	}

	return nil
}

// AddSkillToWorkspace adds a skill to a workspace
func AddSkillToWorkspace(skillName, workspaceName string) error {
	ws, err := GetWorkspace(workspaceName)
	if err != nil {
		return err
	}

	// Check if already in workspace (case-insensitive)
	if slicesContainsFold(ws.Skills, skillName) {
		return nil // Already present
	}

	ws.Skills = append(ws.Skills, skillName)

	if err := saveWorkspace(ws); err != nil {
		return err
	}

	// If this is the active workspace, sync symlinks
	config, err := LoadConfig()
	if err != nil {
		return nil // Workspace saved, just can't sync
	}

	if config.ActiveWorkspace == workspaceName {
		return SyncSkillToAgents(skillName, AgentIDs(DetectInstalledAgents()))
	}

	return nil
}

// RemoveSkillFromWorkspace removes a skill from a workspace
func RemoveSkillFromWorkspace(skillName, workspaceName string) error {
	if IsSystemSkill(skillName) {
		return fmt.Errorf("cannot remove system skill '%s' from workspace", skillName)
	}

	ws, err := GetWorkspace(workspaceName)
	if err != nil {
		return err
	}

	// Find and remove skill (case-insensitive)
	found := false
	newSkills := make([]string, 0, len(ws.Skills))
	for _, s := range ws.Skills {
		if strings.EqualFold(s, skillName) {
			found = true
			continue
		}
		newSkills = append(newSkills, s)
	}

	if !found {
		return nil // Not in workspace
	}

	ws.Skills = newSkills

	if err := saveWorkspace(ws); err != nil {
		return err
	}

	// If this is the active workspace, remove symlinks
	config, err := LoadConfig()
	if err != nil {
		return nil
	}

	if config.ActiveWorkspace == workspaceName {
		return RemoveSkillFromAgents(skillName, AgentIDs(DetectInstalledAgents()))
	}

	return nil
}

// GetWorkspaceInfo returns WorkspaceInfo for all workspaces (for frontend)
func GetWorkspaceInfo() ([]WorkspaceInfo, error) {
	workspaces, err := ListWorkspaces()
	if err != nil {
		return nil, err
	}

	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	infos := make([]WorkspaceInfo, len(workspaces))
	for i, ws := range workspaces {
		infos[i] = WorkspaceInfo{
			Name:        ws.Name,
			Description: ws.Description,
			Skills:      ws.Skills,
			IsActive:    ws.Name == config.ActiveWorkspace,
		}
	}

	return infos, nil
}

// EnsureDefaultWorkspace creates the default workspace if it doesn't exist
func EnsureDefaultWorkspace() error {
	if err := EnsureScribeDirs(); err != nil {
		return err
	}

	wsPath, err := GetWorkspacePath(DefaultWorkspaceName)
	if err != nil {
		return err
	}

	// Check if default workspace exists
	if _, err := os.Stat(wsPath); err == nil {
		return nil // Already exists
	}

	// Create default workspace with all installed skills
	skills, err := ListInstalledSkills()
	if err != nil {
		skills = []string{}
	}

	ws := &Workspace{
		Name:        DefaultWorkspaceName,
		Description: "All installed skills",
		Skills:      skills,
	}

	return saveWorkspace(ws)
}

// saveWorkspace writes a workspace definition to disk
func saveWorkspace(ws *Workspace) error {
	if err := EnsureScribeDirs(); err != nil {
		return err
	}

	wsPath, err := GetWorkspacePath(ws.Name)
	if err != nil {
		return err
	}

	// Strip system skills before persisting (they are injected at read time)
	filtered := make([]string, 0, len(ws.Skills))
	for _, s := range ws.Skills {
		if !IsSystemSkill(s) {
			filtered = append(filtered, s)
		}
	}
	toSave := &Workspace{
		Name:        ws.Name,
		Description: ws.Description,
		Skills:      filtered,
	}

	data, err := json.MarshalIndent(toSave, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(wsPath, data, 0o644)
}

// createDefaultWorkspace creates a default workspace with all installed skills
func createDefaultWorkspace() *Workspace {
	skills, _ := ListInstalledSkills()
	return &Workspace{
		Name:        DefaultWorkspaceName,
		Description: "All installed skills",
		Skills:      injectSystemSkills(skills),
	}
}

// injectSystemSkills ensures system skills are always present in a skill list.
// System skills are prepended so they appear first.
func injectSystemSkills(skills []string) []string {
	if slicesContainsFold(skills, SystemSkillName) {
		return skills // Already present
	}
	return append([]string{SystemSkillName}, skills...)
}

// skillDiff returns skills in a that are not in b (case-insensitive)
func skillDiff(a, b []string) []string {
	bSet := make(map[string]bool, len(b))
	for _, s := range b {
		bSet[strings.ToLower(s)] = true
	}

	var diff []string
	for _, s := range a {
		if !bSet[strings.ToLower(s)] {
			diff = append(diff, s)
		}
	}
	return diff
}

// slicesContainsFold checks if a string slice contains a value (case-insensitive)
func slicesContainsFold(slice []string, val string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, val) {
			return true
		}
	}
	return false
}

// AddSkillToActiveAndDefaultWorkspace adds a skill to the default workspace.
// If the default workspace is active, the skill is also synced to agents.
// If a different workspace is active, the skill is stored but not synced —
// the user must explicitly add it to their active workspace.
func AddSkillToActiveAndDefaultWorkspace(skillName string) error {
	// Only add to default workspace — respect the user's active workspace
	return AddSkillToWorkspace(skillName, DefaultWorkspaceName)
}

// RemoveSkillFromAllWorkspaces removes a skill from all workspaces
// This is called when uninstalling a skill
func RemoveSkillFromAllWorkspaces(skillName string) error {
	if IsSystemSkill(skillName) {
		return nil // Silently ignore -- system skills are virtual in workspaces
	}

	workspaces, err := ListWorkspaces()
	if err != nil {
		return err
	}

	for _, ws := range workspaces {
		// Find and remove skill (case-insensitive)
		newSkills := make([]string, 0, len(ws.Skills))
		for _, s := range ws.Skills {
			if !strings.EqualFold(s, skillName) {
				newSkills = append(newSkills, s)
			}
		}
		ws.Skills = newSkills
		// Continue with other workspaces on error
		_ = saveWorkspace(ws)
	}

	return nil
}

// RebuildDefaultWorkspace rebuilds the default workspace to include all installed skills
func RebuildDefaultWorkspace() error {
	skills, err := ListInstalledSkills()
	if err != nil {
		return err
	}

	ws := &Workspace{
		Name:        DefaultWorkspaceName,
		Description: "All installed skills",
		Skills:      skills,
	}

	return saveWorkspace(ws)
}

// CleanWorkspaces removes references to skills that no longer exist
func CleanWorkspaces() error {
	installed, err := ListInstalledSkills()
	if err != nil {
		return err
	}

	installedSet := make(map[string]bool, len(installed))
	for _, s := range installed {
		installedSet[strings.ToLower(s)] = true
	}

	workspaces, err := ListWorkspaces()
	if err != nil {
		return err
	}

	workspacesDir, err := GetWorkspacesDir()
	if err != nil {
		return err
	}

	for _, ws := range workspaces {
		changed := false
		newSkills := make([]string, 0, len(ws.Skills))
		for _, s := range ws.Skills {
			if installedSet[strings.ToLower(s)] {
				newSkills = append(newSkills, s)
			} else {
				changed = true
			}
		}

		if changed {
			ws.Skills = newSkills
			wsPath := filepath.Join(workspacesDir, ws.Name+".json")
			data, err := json.MarshalIndent(ws, "", "  ")
			if err != nil {
				continue
			}
			_ = os.WriteFile(wsPath, data, 0o644)
		}
	}

	return nil
}
