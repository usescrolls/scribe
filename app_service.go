package main

import (
	"fmt"
	"strings"
	"sync"

	scribe "github.com/usescrolls/scribe/internal"
)

// AppService provides bindings for the frontend
type AppService struct {
	mu            sync.Mutex
	pendingSource *scribe.SourceInfo
	pendingSkills []*scribe.Skill
	pendingFetch  *scribe.FetchResult
}

// NewAppService creates a new AppService
func NewAppService() *AppService {
	return &AppService{}
}

// clearPending cleans up any pending discover state (must be called with mu held)
func (a *AppService) clearPending() {
	if a.pendingFetch != nil {
		a.pendingFetch.Cleanup()
	}
	a.pendingSource = nil
	a.pendingSkills = nil
	a.pendingFetch = nil
}

// GetVersion returns the application version
func (a *AppService) GetVersion() string {
	return scribe.Version
}

// Log writes a log message from the frontend to the unified log file.
// Valid levels: "debug", "info", "warn", "error".
func (a *AppService) Log(level, message string) {
	frontendLogger := scribe.Logger.With("source", "frontend")

	switch level {
	case "debug":
		frontendLogger.Debug(message)
	case "info":
		frontendLogger.Info(message)
	case "warn":
		frontendLogger.Warn(message)
	case "error":
		frontendLogger.Error(message)
	default:
		frontendLogger.Info(message)
	}
}

// ======================================================================
// Skills API
// ======================================================================

// GetSkills returns all installed skills
func (a *AppService) GetSkills() ([]scribe.SkillInfo, error) {
	return scribe.GetAllSkillInfo()
}

// GetSkillContent returns the SKILL.md body content (markdown after frontmatter) for a skill
func (a *AppService) GetSkillContent(name string) (string, error) {
	skill, err := scribe.ReadSkill(name)
	if err != nil {
		return "", err
	}
	return skill.Content, nil
}

// GetSkillCount returns the number of installed skills
func (a *AppService) GetSkillCount() int {
	skills, err := scribe.ReadAllSkills()
	if err != nil {
		return 0
	}
	return len(skills)
}

// RemoveSkill removes a skill by name from all agents and workspaces
func (a *AppService) RemoveSkill(name string) error {
	if scribe.IsSystemSkill(name) {
		return fmt.Errorf("cannot uninstall system skill '%s'", name)
	}

	scribe.Logger.Info("AppService.RemoveSkill called", "name", name)

	// Remove from all workspaces first
	if err := scribe.RemoveSkillFromAllWorkspaces(name); err != nil {
		scribe.Logger.Error("failed to remove skill from workspaces", "name", name, "error", err)
		return err
	}

	// UninstallSkill handles removing from agents and canonical location
	if err := scribe.UninstallSkill(name); err != nil {
		scribe.Logger.Error("failed to uninstall skill", "name", name, "error", err)
		return err
	}

	scribe.Logger.Info("AppService.RemoveSkill succeeded", "name", name)

	// Emit event to update frontend
	if wailsApp != nil {
		wailsApp.Event.Emit("skills-updated", nil)
	}

	return nil
}

// UpdateSkill updates a skill to its latest version from the original source
func (a *AppService) UpdateSkill(name string) (*scribe.UpdateResult, error) {
	scribe.Logger.Info("AppService.UpdateSkill called", "name", name)

	result, err := scribe.UpdateSkill(name, false)
	if err != nil {
		return nil, err
	}

	scribe.Logger.Info("AppService.UpdateSkill succeeded", "name", name, "updated", result.Updated)

	// Emit event to update frontend
	if wailsApp != nil {
		wailsApp.Event.Emit("skills-updated", nil)
	}

	return result, nil
}

// ======================================================================
// Workspaces API
// ======================================================================

// GetWorkspaces returns all workspaces with their active status
func (a *AppService) GetWorkspaces() ([]scribe.WorkspaceInfo, error) {
	return scribe.GetWorkspaceInfo()
}

// GetActiveWorkspaceName returns the name of the active workspace
func (a *AppService) GetActiveWorkspaceName() (string, error) {
	ws, err := scribe.GetActiveWorkspace()
	if err != nil {
		return "", err
	}
	return ws.Name, nil
}

// SetActiveWorkspace switches to a different workspace
func (a *AppService) SetActiveWorkspace(name string) error {
	scribe.Logger.Info("AppService.SetActiveWorkspace called", "name", name)

	current, err := scribe.GetActiveWorkspace()
	if err != nil {
		return err
	}

	target, err := scribe.GetWorkspace(name)
	if err != nil {
		return err
	}

	// Sync symlinks to match target workspace
	if err := scribe.SyncWorkspace(current, target); err != nil {
		return err
	}

	// Update active workspace in config
	if err := scribe.SetActiveWorkspace(name); err != nil {
		return err
	}

	scribe.Logger.Info("AppService.SetActiveWorkspace succeeded", "name", name)

	// Emit event to update frontend
	if wailsApp != nil {
		wailsApp.Event.Emit("workspace-changed", name)
		wailsApp.Event.Emit("skills-updated", nil)
	}

	return nil
}

// CreateWorkspace creates a new workspace
func (a *AppService) CreateWorkspace(name, description string) error {
	ws := &scribe.Workspace{
		Name:        name,
		Description: description,
		Skills:      []string{},
	}
	return scribe.CreateWorkspace(ws)
}

// DeleteWorkspace removes a workspace
func (a *AppService) DeleteWorkspace(name string) error {
	err := scribe.DeleteWorkspace(name)
	if err == nil && wailsApp != nil {
		wailsApp.Event.Emit("workspace-changed", name)
		wailsApp.Event.Emit("skills-updated", nil)
	}
	return err
}

// AddSkillToWorkspace adds a skill to a specific workspace
func (a *AppService) AddSkillToWorkspace(skillName, workspaceName string) error {
	err := scribe.AddSkillToWorkspace(skillName, workspaceName)
	if err == nil && wailsApp != nil {
		wailsApp.Event.Emit("workspace-changed", workspaceName)
	}
	return err
}

// RemoveSkillFromWorkspace removes a skill from a specific workspace
func (a *AppService) RemoveSkillFromWorkspace(skillName, workspaceName string) error {
	if scribe.IsSystemSkill(skillName) {
		return fmt.Errorf("cannot remove system skill '%s' from workspace", skillName)
	}

	err := scribe.RemoveSkillFromWorkspace(skillName, workspaceName)
	if err == nil && wailsApp != nil {
		wailsApp.Event.Emit("workspace-changed", workspaceName)
		wailsApp.Event.Emit("skills-updated", nil)
	}
	return err
}

// ======================================================================
// Agents API
// ======================================================================

// GetAgentStatus returns the status of all supported agents
func (a *AppService) GetAgentStatus() []scribe.AgentStatus {
	scrollsDir, err := scribe.GetScrollsDir()
	if err != nil {
		scribe.Logger.Error("failed to get scrolls dir", "error", err)
		return []scribe.AgentStatus{}
	}
	return scribe.GetAgentStatus(scrollsDir)
}

// GetInstalledAgentCount returns how many agents are installed
func (a *AppService) GetInstalledAgentCount() int {
	agents := scribe.DetectInstalledAgents()
	return len(agents)
}

// GetTotalAgentCount returns the total number of supported agents
func (a *AppService) GetTotalAgentCount() int {
	return len(scribe.GetAllAgents())
}

// ResyncWorkspace forces a resync of all skills in current workspace to agents
func (a *AppService) ResyncWorkspace() error {
	err := scribe.ResyncCurrentWorkspace()
	if err == nil && wailsApp != nil {
		wailsApp.Event.Emit("skills-updated", nil)
	}
	return err
}

// ======================================================================
// Onboarding API
// ======================================================================

// IsOnboardingCompleted checks if onboarding has been completed
func (a *AppService) IsOnboardingCompleted() bool {
	completed, err := scribe.IsOnboardingCompleted()
	if err != nil {
		scribe.Logger.Error("failed to check onboarding status", "error", err)
		return false
	}
	return completed
}

// AreTermsAccepted checks if the user has accepted the terms and conditions
func (a *AppService) AreTermsAccepted() bool {
	accepted, err := scribe.AreTermsAccepted()
	if err != nil {
		scribe.Logger.Error("failed to check terms acceptance", "error", err)
		return false
	}
	return accepted
}

// AcceptTerms records the user's acceptance of the terms and conditions
func (a *AppService) AcceptTerms() error {
	return scribe.AcceptTerms()
}

// GetTermsClauses returns the terms and conditions clauses (single source of truth)
func (a *AppService) GetTermsClauses() []scribe.TermsClause {
	return scribe.TermsClauses
}

// GetTermsAcceptedAt returns the RFC3339 timestamp of when terms were accepted, or empty string
func (a *AppService) GetTermsAcceptedAt() string {
	config, err := scribe.LoadConfig()
	if err != nil {
		scribe.Logger.Error("failed to load config for terms timestamp", "error", err)
		return ""
	}
	return config.TermsAcceptedAt
}

// CompleteOnboarding marks onboarding as completed
func (a *AppService) CompleteOnboarding() error {
	err := scribe.CompleteOnboarding()
	if err == nil && wailsApp != nil {
		wailsApp.Event.Emit("onboarding-completed", nil)
	}
	return err
}

// DetectExistingSkills scans agent directories for existing skills
func (a *AppService) DetectExistingSkills() []scribe.ExistingSkillInfo {
	skills, err := scribe.DetectExistingSkills()
	if err != nil {
		scribe.Logger.Error("failed to detect existing skills", "error", err)
		return []scribe.ExistingSkillInfo{}
	}
	return skills
}

// DetectSkillConflicts finds skills with the same name in different agent directories
func (a *AppService) DetectSkillConflicts() []scribe.SkillConflict {
	skills, err := scribe.DetectExistingSkills()
	if err != nil {
		scribe.Logger.Error("failed to detect existing skills", "error", err)
		return []scribe.SkillConflict{}
	}
	return scribe.DetectSkillConflicts(skills)
}

// ImportAllExistingSkills imports all detected skills to ~/.scribe/scrolls/
func (a *AppService) ImportAllExistingSkills() error {
	skills, err := scribe.DetectExistingSkills()
	if err != nil {
		return err
	}
	err = scribe.ImportExistingSkills(skills)
	if err == nil && wailsApp != nil {
		wailsApp.Event.Emit("skills-updated", nil)
	}
	return err
}

// DeleteAllExistingSkills removes all detected skills from agent directories
func (a *AppService) DeleteAllExistingSkills() error {
	skills, err := scribe.DetectExistingSkills()
	if err != nil {
		return err
	}
	return scribe.DeleteExistingSkills(skills)
}

// ResolveSkillConflict imports a specific skill version when there's a naming conflict
func (a *AppService) ResolveSkillConflict(skillPath string) error {
	return scribe.ImportSelectedSkills([]string{skillPath})
}

// InstallFromSource installs skills from a source string (e.g. "owner/repo", GitHub URL, zip URL)
func (a *AppService) InstallFromSource(sourceStr string) (*scribe.InstallResult, error) {
	scribe.Logger.Info("AppService.InstallFromSource called", "source", sourceStr)

	source, err := scribe.ParseSourceString(sourceStr)
	if err != nil {
		return &scribe.InstallResult{ErrorMessage: err.Error()}, nil
	}

	// Ensure directories exist
	if err := scribe.EnsureScribeDirs(); err != nil {
		return &scribe.InstallResult{ErrorMessage: fmt.Sprintf("Failed to create directories: %v", err)}, nil
	}
	if err := scribe.EnsureDefaultWorkspace(); err != nil {
		scribe.Logger.Warn("failed to ensure default workspace", "error", err)
	}

	// Create progress emitter for frontend
	emit := func(e scribe.ProgressEvent) {
		if wailsApp != nil {
			wailsApp.Event.Emit("progress", e)
		}
	}

	// Fetch and discover skills
	skills, fetchResult, err := scribe.FetchAndDiscoverSkills(source, emit)
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}
	if err != nil {
		return &scribe.InstallResult{ErrorMessage: fmt.Sprintf("Failed to fetch skills: %v", err)}, nil
	}

	if len(skills) == 0 {
		return &scribe.InstallResult{ErrorMessage: "No skills found in source"}, nil
	}

	// Filter out already-installed skills and resolve name conflicts
	newSkills, alreadyInstalled, err := scribe.FilterAndResolveConflicts(skills, source)
	if err != nil {
		scribe.Logger.Error("failed to resolve name conflicts", "error", err)
	}
	if len(newSkills) == 0 {
		msg := fmt.Sprintf("All %d skill(s) from this source are already installed", len(alreadyInstalled))
		if len(alreadyInstalled) == 1 {
			msg = fmt.Sprintf("Skill '%s' is already installed", alreadyInstalled[0])
		}
		return &scribe.InstallResult{ErrorMessage: msg}, nil
	}

	// Extract git commit info from fetched repo
	gitInfo := scribe.GetHeadCommitInfo(fetchResult.ContentDir)

	// Install each discovered skill
	result := &scribe.InstallResult{}
	opts := scribe.InstallOptions{Yes: true, IsPrivate: fetchResult.IsPrivate}
	for _, skill := range newSkills {
		scribe.Logger.Info("installing skill from GUI", "name", skill.Name)

		if err := scribe.InstallSkill(skill, source, opts, gitInfo, emit); err != nil {
			scribe.Logger.Error("failed to install skill", "name", skill.Name, "error", err)
			continue
		}

		if err := scribe.AddSkillToActiveAndDefaultWorkspace(skill.Name); err != nil {
			scribe.Logger.Warn("failed to add to workspace", "skill", skill.Name, "error", err)
		}

		result.SkillNames = append(result.SkillNames, skill.Name)
	}

	result.SkillsCount = len(result.SkillNames)
	result.Success = result.SkillsCount > 0

	if !result.Success {
		result.ErrorMessage = "Failed to install any skills"
	}

	// Emit events to update frontend
	if wailsApp != nil {
		wailsApp.Event.Emit("skills-updated", nil)
		wailsApp.Event.Emit("workspace-changed", nil)
	}

	scribe.Logger.Info("AppService.InstallFromSource completed",
		"skills_installed", result.SkillsCount,
		"skill_names", result.SkillNames)

	return result, nil
}

// DiscoverFromSource fetches a source and returns discovered skills without installing
func (a *AppService) DiscoverFromSource(sourceStr string) (*scribe.DiscoverResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	scribe.Logger.Info("AppService.DiscoverFromSource called", "source", sourceStr)

	// Clean up any previous pending state
	a.clearPending()

	source, err := scribe.ParseSourceString(sourceStr)
	if err != nil {
		return nil, fmt.Errorf("invalid source: %w", err)
	}

	// Ensure directories exist
	if err := scribe.EnsureScribeDirs(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Create progress emitter for frontend
	emit := func(e scribe.ProgressEvent) {
		if wailsApp != nil {
			wailsApp.Event.Emit("progress", e)
		}
	}

	// Fetch and discover skills
	skills, fetchResult, err := scribe.FetchAndDiscoverSkills(source, emit)
	if err != nil {
		if fetchResult != nil {
			fetchResult.Cleanup()
		}
		return nil, fmt.Errorf("failed to fetch skills: %w", err)
	}

	if len(skills) == 0 {
		if fetchResult != nil {
			fetchResult.Cleanup()
		}
		return nil, fmt.Errorf("no skills found in source")
	}

	// Cache for ConfirmInstall
	a.pendingSource = source
	a.pendingSkills = skills
	a.pendingFetch = fetchResult

	// Check which skills are already installed from this source
	_, alreadyInstalledNames, _ := scribe.FilterAlreadyInstalled(skills, source)
	alreadySet := make(map[string]bool, len(alreadyInstalledNames))
	for _, name := range alreadyInstalledNames {
		alreadySet[strings.ToLower(name)] = true
	}

	// Build result
	result := &scribe.DiscoverResult{
		Source:     sourceStr,
		SourceType: source.Type,
	}
	for _, skill := range skills {
		result.Skills = append(result.Skills, scribe.DiscoveredSkill{
			Name:             skill.Name,
			Description:      skill.Description,
			AlreadyInstalled: alreadySet[strings.ToLower(skill.Name)],
		})
	}

	scribe.Logger.Info("AppService.DiscoverFromSource completed",
		"skills_found", len(skills))

	return result, nil
}

// ConfirmInstall installs previously discovered skills and optionally adds them to workspaces
func (a *AppService) ConfirmInstall(skillNames, workspaceNames []string) (*scribe.InstallResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	scribe.Logger.Info("AppService.ConfirmInstall called",
		"skills", skillNames, "workspaces", workspaceNames)

	if a.pendingSkills == nil || a.pendingSource == nil {
		return &scribe.InstallResult{ErrorMessage: "No pending discovery. Call DiscoverFromSource first."}, nil
	}

	// Build a set of requested skill names
	requested := make(map[string]bool, len(skillNames))
	for _, name := range skillNames {
		requested[name] = true
	}

	// Ensure default workspace exists
	if err := scribe.EnsureDefaultWorkspace(); err != nil {
		scribe.Logger.Warn("failed to ensure default workspace", "error", err)
	}

	// Filter to only requested skills
	var toInstall []*scribe.Skill
	for _, skill := range a.pendingSkills {
		if requested[skill.Name] {
			toInstall = append(toInstall, skill)
		}
	}

	// Extract git commit info from cached repo
	var gitInfo *scribe.GitCommitInfo
	if a.pendingFetch != nil {
		gitInfo = scribe.GetHeadCommitInfo(a.pendingFetch.ContentDir)
	}

	// Create progress emitter for frontend
	emit := func(e scribe.ProgressEvent) {
		if wailsApp != nil {
			wailsApp.Event.Emit("progress", e)
		}
	}

	// Filter out already-installed skills and resolve name conflicts
	newToInstall, alreadyInstalled, err := scribe.FilterAndResolveConflicts(toInstall, a.pendingSource)
	if err != nil {
		scribe.Logger.Error("failed to resolve name conflicts", "error", err)
	}
	if len(newToInstall) == 0 {
		msg := fmt.Sprintf("All %d skill(s) from this source are already installed", len(alreadyInstalled))
		if len(alreadyInstalled) == 1 {
			msg = fmt.Sprintf("Skill '%s' is already installed", alreadyInstalled[0])
		}
		a.clearPending()
		return &scribe.InstallResult{ErrorMessage: msg}, nil
	}

	// Install each requested skill
	result := &scribe.InstallResult{}
	isPrivate := a.pendingFetch != nil && a.pendingFetch.IsPrivate
	opts := scribe.InstallOptions{Yes: true, IsPrivate: isPrivate}
	for i, skill := range newToInstall {
		// Emit progress event so the frontend can show per-skill status
		if wailsApp != nil {
			wailsApp.Event.Emit("install-progress", map[string]any{
				"skillName": skill.Name,
				"current":   i + 1,
				"total":     len(newToInstall),
			})
		}

		emit(scribe.ProgressEvent{
			Phase:   "install",
			Step:    "start",
			Message: fmt.Sprintf("Installing %s (%d/%d)...", skill.Name, i+1, len(newToInstall)),
			Detail:  skill.Name,
		})

		scribe.Logger.Info("installing skill from GUI", "name", skill.Name)

		if err := scribe.InstallSkill(skill, a.pendingSource, opts, gitInfo, emit); err != nil {
			scribe.Logger.Error("failed to install skill", "name", skill.Name, "error", err)
			emit(scribe.ProgressEvent{
				Phase:   "install",
				Step:    "error",
				Message: fmt.Sprintf("Failed: %v", err),
				Detail:  skill.Name,
			})
			continue
		}

		// Add to user-selected workspaces only (from the UI checkboxes)
		for _, wsName := range workspaceNames {
			emit(scribe.ProgressEvent{
				Phase:   "install",
				Step:    "workspace",
				Message: fmt.Sprintf("Adding to %s...", wsName),
				Detail:  wsName,
			})
			if err := scribe.AddSkillToWorkspace(skill.Name, wsName); err != nil {
				scribe.Logger.Warn("failed to add to workspace",
					"skill", skill.Name, "workspace", wsName, "error", err)
			}
		}

		emit(scribe.ProgressEvent{
			Phase:   "install",
			Step:    "done",
			Message: fmt.Sprintf("Installed %s", skill.Name),
			Detail:  skill.Name,
		})

		result.SkillNames = append(result.SkillNames, skill.Name)
	}

	result.SkillsCount = len(result.SkillNames)
	result.Success = result.SkillsCount > 0

	if !result.Success {
		result.ErrorMessage = "Failed to install any skills"
	}

	// Cleanup pending state
	a.clearPending()

	// Emit events to update frontend
	if wailsApp != nil {
		wailsApp.Event.Emit("skills-updated", nil)
		if len(workspaceNames) > 0 {
			wailsApp.Event.Emit("workspace-changed", nil)
		}
	}

	scribe.Logger.Info("AppService.ConfirmInstall completed",
		"skills_installed", result.SkillsCount,
		"skill_names", result.SkillNames)

	return result, nil
}

// CancelDiscover cancels a pending discovery and cleans up resources
func (a *AppService) CancelDiscover() {
	a.mu.Lock()
	defer a.mu.Unlock()

	scribe.Logger.Info("AppService.CancelDiscover called")
	a.clearPending()
}

// InstallDemoSkill installs the scribe-welcome demo skill
func (a *AppService) InstallDemoSkill() error {
	err := scribe.InstallDemoSkill()
	if err == nil && wailsApp != nil {
		wailsApp.Event.Emit("skills-updated", nil)
	}
	return err
}

// ======================================================================
// Marketplace API
// ======================================================================

// SearchMarketplace searches for skill repos using the specified provider
func (a *AppService) SearchMarketplace(providerID, query string, page int) (*scribe.MarketplaceResult, error) {
	scribe.Logger.Info("AppService.SearchMarketplace called", "provider", providerID, "query", query, "page", page)
	return scribe.SearchMarketplace(providerID, query, page)
}

// GetMarketplaceProviders returns available marketplace provider IDs and display names
func (a *AppService) GetMarketplaceProviders() []scribe.MarketplaceProviderInfo {
	return scribe.GetMarketplaceProviders()
}

// GetRepoReadme fetches the README.md content for a GitHub repository
func (a *AppService) GetRepoReadme(owner, repo string) (string, error) {
	scribe.Logger.Info("AppService.GetRepoReadme called", "owner", owner, "repo", repo)
	return scribe.GetRepoReadme(owner, repo)
}

// ======================================================================
// Updates API
// ======================================================================

// CheckSourceGroupUpdates checks all installed source groups for available updates.
// Each remote repository is fetched only once regardless of how many skills it contains.
func (a *AppService) CheckSourceGroupUpdates() map[string]scribe.SourceGroupCheckResult {
	scribe.Logger.Info("AppService.CheckSourceGroupUpdates called")
	results := scribe.CheckAllSourcesForUpdates()
	if results == nil {
		return map[string]scribe.SourceGroupCheckResult{}
	}

	updatable := 0
	for _, r := range results {
		if r.HasUpdates {
			updatable++
		}
	}
	scribe.Logger.Info("source group update check completed", "groups", len(results), "withUpdates", updatable)
	return results
}

// CheckForAppUpdate queries GitHub for the latest release and returns update info
func (a *AppService) CheckForAppUpdate() (*scribe.UpdateInfo, error) {
	scribe.Logger.Info("AppService.CheckForAppUpdate called")
	info, err := scribe.CheckForUpdate("")
	if err != nil {
		scribe.Logger.Warn("update check failed", "error", err)
		return nil, err
	}
	scribe.Logger.Info("update check completed",
		"current", info.CurrentVersion,
		"latest", info.LatestVersion,
		"updateAvailable", info.UpdateAvailable)
	return info, nil
}

// IsUpdateNotificationsDisabled returns whether the user has suppressed update toasts
func (a *AppService) IsUpdateNotificationsDisabled() bool {
	disabled, err := scribe.IsUpdateNotificationsDisabled()
	if err != nil {
		scribe.Logger.Error("failed to check update notification preference", "error", err)
		return false
	}
	return disabled
}

// SetUpdateNotificationsDisabled sets the user's preference for update notifications
func (a *AppService) SetUpdateNotificationsDisabled(disabled bool) error {
	return scribe.SetUpdateNotificationsDisabled(disabled)
}

// GetInstallMethod returns how Scribe was installed ("homebrew", "app-bundle", "binary", "dev", "unknown").
func (a *AppService) GetInstallMethod() string {
	return scribe.DetectInstallMethod()
}

// UpgradeApp performs a self-update of the Scribe binary.
func (a *AppService) UpgradeApp() (*scribe.SelfUpdateResult, error) {
	scribe.Logger.Info("AppService.UpgradeApp called")
	result, err := scribe.SelfUpdate("")
	if err != nil {
		scribe.Logger.Error("self-update failed", "error", err)
		return nil, err
	}
	scribe.Logger.Info("self-update completed",
		"updated", result.Updated,
		"oldVersion", result.OldVersion,
		"newVersion", result.NewVersion)
	return result, nil
}
