package scribe

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CheckSkillForUpdate checks if a single skill has a remote update available.
func CheckSkillForUpdate(skillName string) CheckResult {
	result := CheckResult{Name: skillName}

	if IsSystemSkill(skillName) {
		result.Error = "system skill (managed internally)"
		return result
	}

	skill, err := ReadSkill(skillName)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read skill: %v", err)
		return result
	}

	if skill.Meta == nil {
		result.Error = "no metadata (manually added skill)"
		return result
	}

	result.CurrentHash = skill.Meta.ContentHash

	if skill.Meta.SourceType == "local" {
		result.Error = "local source (cannot check for updates)"
		return result
	}

	source := ReconstructSource(skill.Meta)

	skills, fetchResult, err := FetchAndDiscoverSkills(source)
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch: %v", err)
		if IsAuthError(err) {
			errMsg += " (auth issue — see 'scribe install --help')"
		}
		result.Error = errMsg
		return result
	}
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}

	// Match by frontmatter name — for qualified storage names like
	// "alice-skills--commit", we need to match against "commit".
	matchName := FrontmatterNameFromStorage(skillName)
	var remoteSkill *Skill
	for _, s := range skills {
		if s.Name == matchName || filepath.Base(s.Path) == matchName {
			remoteSkill = s
			break
		}
	}

	if remoteSkill == nil {
		result.Error = "skill not found in remote source"
		return result
	}

	remoteContent, err := os.ReadFile(filepath.Join(remoteSkill.Path, SkillFileName))
	if err != nil {
		result.Error = fmt.Sprintf("failed to read remote skill: %v", err)
		return result
	}

	result.RemoteHash = ComputeContentHash(string(remoteContent))
	result.NeedsUpdate = result.CurrentHash != result.RemoteHash

	return result
}

// CheckSourceForUpdates fetches a source repo once and checks all given skills
// against their installed content hashes. This is much more efficient than
// calling CheckSkillForUpdate per-skill when multiple skills share a source.
// It also returns any new skills discovered in the source that aren't installed.
func CheckSourceForUpdates(sourceStr string, skillNames []string) ([]CheckResult, []DiscoveredSkill) {
	if len(skillNames) == 0 {
		return nil, nil
	}

	// Read the first skill to reconstruct the source info
	firstSkill, err := ReadSkill(skillNames[0])
	if err != nil || firstSkill.Meta == nil {
		var results []CheckResult
		for _, name := range skillNames {
			results = append(results, CheckResult{Name: name, Error: fmt.Sprintf("failed to read skill: %v", err)})
		}
		return results, nil
	}

	source := ReconstructSource(firstSkill.Meta)

	remoteSkills, fetchResult, err := FetchAndDiscoverSkills(source)
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch: %v", err)
		if IsAuthError(err) {
			errMsg += " (auth issue)"
		}
		var results []CheckResult
		for _, name := range skillNames {
			results = append(results, CheckResult{Name: name, Error: errMsg})
		}
		return results, nil
	}
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}

	// Index remote skills by name for fast lookup
	remoteByName := make(map[string]*Skill, len(remoteSkills))
	for _, s := range remoteSkills {
		remoteByName[s.Name] = s
	}

	// Build set of installed frontmatter names for new-skill detection
	installedFM := make(map[string]bool, len(skillNames))
	for _, name := range skillNames {
		installedFM[FrontmatterNameFromStorage(name)] = true
	}

	var results []CheckResult
	for _, skillName := range skillNames {
		r := CheckResult{Name: skillName}

		skill, err := ReadSkill(skillName)
		if err != nil {
			r.Error = fmt.Sprintf("failed to read skill: %v", err)
			results = append(results, r)
			continue
		}
		if skill.Meta == nil {
			r.Error = "no metadata"
			results = append(results, r)
			continue
		}

		r.CurrentHash = skill.Meta.ContentHash

		// Look up by frontmatter name for qualified storage names
		remote := remoteByName[FrontmatterNameFromStorage(skillName)]
		if remote == nil {
			r.Error = "skill not found in remote source"
			results = append(results, r)
			continue
		}

		remoteContent, err := os.ReadFile(filepath.Join(remote.Path, SkillFileName))
		if err != nil {
			r.Error = fmt.Sprintf("failed to read remote skill: %v", err)
			results = append(results, r)
			continue
		}

		r.RemoteHash = ComputeContentHash(string(remoteContent))
		r.NeedsUpdate = r.CurrentHash != r.RemoteHash
		results = append(results, r)
	}

	// Discover new skills in the source that aren't installed
	var newSkills []DiscoveredSkill
	for _, s := range remoteSkills {
		if !installedFM[s.Name] {
			newSkills = append(newSkills, DiscoveredSkill{
				Name:        s.Name,
				Description: s.Description,
			})
		}
	}

	return results, newSkills
}

// CheckAllSourcesForUpdates checks every installed source group for updates,
// fetching each remote repository only once. Returns results keyed by source string.
func CheckAllSourcesForUpdates() map[string]SourceGroupCheckResult {
	skillNames, err := ListInstalledSkills()
	if err != nil {
		Logger.Warn("failed to list installed skills for update check", "error", err)
		return nil
	}

	// Group skill names by their source
	type sourceGroup struct {
		source     string
		sourceType string
		skills     []string
	}
	groups := make(map[string]*sourceGroup)

	for _, name := range skillNames {
		if IsSystemSkill(name) {
			continue
		}
		skill, err := ReadSkill(name)
		if err != nil || skill.Meta == nil {
			continue
		}
		if skill.Meta.SourceType == "local" || skill.Meta.SourceType == "builtin" {
			continue
		}
		key := skill.Meta.Source
		if key == "" {
			continue
		}
		g, ok := groups[key]
		if !ok {
			g = &sourceGroup{source: key, sourceType: skill.Meta.SourceType}
			groups[key] = g
		}
		g.skills = append(g.skills, name)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	results := make(map[string]SourceGroupCheckResult, len(groups))

	for key, g := range groups {
		checkResults, newSkills := CheckSourceForUpdates(g.source, g.skills)

		sgr := SourceGroupCheckResult{
			Source:             key,
			CheckedAt:          now,
			NewAvailableSkills: newSkills,
		}

		for _, cr := range checkResults {
			if cr.Error != "" && sgr.Error == "" {
				sgr.Error = cr.Error
			}
			if cr.NeedsUpdate {
				sgr.HasUpdates = true
				sgr.UpdatedSkillNames = append(sgr.UpdatedSkillNames, cr.Name)
			}
		}

		results[key] = sgr
	}

	return results
}

// UpdateSkill updates a skill to its latest version from its source.
// If force is true, the update proceeds even when the content hash hasn't changed.
func UpdateSkill(skillName string, force bool) (*UpdateResult, error) {
	if IsSystemSkill(skillName) {
		return nil, fmt.Errorf("cannot update system skill '%s' (managed internally)", skillName)
	}

	skill, err := ReadSkill(skillName)
	if err != nil {
		return nil, fmt.Errorf("skill not found: %w", err)
	}

	if skill.Meta == nil {
		return nil, fmt.Errorf("skill has no metadata, cannot update")
	}

	if skill.Meta.SourceType == "local" {
		return nil, fmt.Errorf("local source, cannot update")
	}

	// Capture old hash for result
	oldHash := skill.Meta.CommitHash
	if oldHash == "" && len(skill.Meta.ContentHash) > 13 {
		oldHash = skill.Meta.ContentHash[7:14] // truncate "sha256:abcdefg..." to "abcdefg"
	}

	source := ReconstructSource(skill.Meta)

	skills, fetchResult, err := FetchAndDiscoverSkills(source)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from source: %w", err)
	}
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}

	gitInfo := GetHeadCommitInfo(fetchResult.ContentDir)

	// Find the specific skill — match by frontmatter name, which may differ
	// from the storage name for qualified skills (e.g. "alice-skills--commit"
	// needs to match remote skill "commit").
	updateMatchName := FrontmatterNameFromStorage(skillName)
	var newSkill *Skill
	for _, s := range skills {
		if s.Name == updateMatchName || filepath.Base(s.Path) == updateMatchName {
			newSkill = s
			break
		}
	}
	if newSkill == nil {
		// Skill was removed from the source repo — clean up locally
		Logger.Info("skill removed from source, uninstalling", "skill", skillName)
		_ = RemoveSkillFromAllWorkspaces(skillName)
		_ = UninstallSkill(skillName)
		return &UpdateResult{
			SkillName: skillName,
			Removed:   true,
		}, nil
	}

	newContent, err := os.ReadFile(filepath.Join(newSkill.Path, SkillFileName))
	if err != nil {
		return nil, fmt.Errorf("failed to read new skill content: %w", err)
	}

	skillDir, err := GetSkillDir(skillName)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill directory: %w", err)
	}

	needsUpdate, err := SkillNeedsUpdate(skillDir, string(newContent))
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	// Build hash info for result
	newHash := ""
	commitDate := ""
	if gitInfo != nil {
		newHash = gitInfo.Hash
		commitDate = gitInfo.Date
	}

	if !needsUpdate && !force {
		// Always update commit info to reflect the latest HEAD so the UI
		// shows current version details after an explicit update.
		if gitInfo != nil && (skill.Meta.CommitHash != gitInfo.Hash || skill.Meta.CommitDate != gitInfo.Date) {
			metaPath, err := GetMetaPath(skillName)
			if err == nil {
				skill.Meta.CommitHash = gitInfo.Hash
				skill.Meta.CommitDate = gitInfo.Date
				skill.Meta.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
				_ = WriteSkillMeta(metaPath, skill.Meta)
			}
		}

		return &UpdateResult{
			SkillName:  skillName,
			Updated:    false,
			OldHash:    oldHash,
			NewHash:    newHash,
			CommitDate: commitDate,
		}, nil
	}

	// Replace existing skill content with new version
	if err := os.RemoveAll(skillDir); err != nil {
		return nil, fmt.Errorf("failed to remove existing: %w", err)
	}
	if err := CopySkillDir(newSkill.Path, skillDir); err != nil {
		return nil, fmt.Errorf("failed to update skill files: %w", err)
	}

	// Update metadata
	metaPath, err := GetMetaPath(skillName)
	if err != nil {
		return nil, fmt.Errorf("failed to get meta path: %w", err)
	}

	meta, err := ReadSkillMeta(metaPath)
	if err != nil {
		meta = NewSkillMeta(source, skill.Meta.SkillPath, string(newContent), gitInfo)
	} else {
		UpdateSkillMeta(meta, string(newContent), gitInfo)
	}

	if err := WriteSkillMeta(metaPath, meta); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	// Re-sync to all agents
	if err := SyncSkillToAgents(skillName, AgentIDs(DetectInstalledAgents())); err != nil {
		Logger.Warn("failed to sync to agents", "skill", skillName, "error", err)
	}

	return &UpdateResult{
		SkillName:  skillName,
		Updated:    true,
		OldHash:    oldHash,
		NewHash:    newHash,
		CommitDate: commitDate,
	}, nil
}
