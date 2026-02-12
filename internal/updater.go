package scribe

import (
	"fmt"
	"os"
	"path/filepath"
)

// CheckSkillForUpdate checks if a single skill has a remote update available.
func CheckSkillForUpdate(skillName string) CheckResult {
	result := CheckResult{Name: skillName}

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

	var remoteSkill *Skill
	for _, s := range skills {
		if s.Name == skillName {
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

// UpdateSkill updates a skill to its latest version from its source.
// If force is true, the update proceeds even when the content hash hasn't changed.
func UpdateSkill(skillName string, force bool) (*UpdateResult, error) {
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

	// Find the specific skill — match by name or directory basename
	var newSkill *Skill
	for _, s := range skills {
		if s.Name == skillName || filepath.Base(s.Path) == skillName {
			newSkill = s
			break
		}
	}
	if newSkill == nil {
		return nil, fmt.Errorf("skill not found in source")
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
		// Backfill git info into metadata if missing (e.g. skill was installed
		// before commit tracking was added).
		if gitInfo != nil && skill.Meta.CommitHash == "" {
			metaPath, err := GetMetaPath(skillName)
			if err == nil {
				skill.Meta.CommitHash = gitInfo.Hash
				skill.Meta.CommitDate = gitInfo.Date
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
