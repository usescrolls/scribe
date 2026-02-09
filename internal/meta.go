package scribe

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"time"
)

// ReadSkillMeta reads the .scribe-meta.json sidecar file for a skill
func ReadSkillMeta(metaPath string) (*SkillMeta, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta SkillMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// WriteSkillMeta writes the .scribe-meta.json sidecar file for a skill
func WriteSkillMeta(metaPath string, meta *SkillMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metaPath, data, 0o644)
}

// ComputeContentHash computes a SHA256 hash of the content
func ComputeContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return "sha256:" + hex.EncodeToString(hash[:])
}

// NewSkillMeta creates a new SkillMeta with the current timestamp
func NewSkillMeta(source *SourceInfo, skillPath, content string, gitInfo *GitCommitInfo) *SkillMeta {
	now := time.Now().UTC().Format(time.RFC3339)

	meta := &SkillMeta{
		Source:      formatSource(source),
		SourceType:  source.Type,
		ContentHash: ComputeContentHash(content),
		InstalledAt: now,
		UpdatedAt:   now,
	}

	if source.URL != "" {
		meta.SourceURL = source.URL
	}

	if skillPath != "" {
		meta.SkillPath = skillPath
	}

	if gitInfo != nil {
		meta.CommitHash = gitInfo.Hash
		meta.CommitDate = gitInfo.Date
	}

	return meta
}

// formatSource formats a SourceInfo into a human-readable source string
func formatSource(source *SourceInfo) string {
	switch source.Type {
	case "github":
		s := source.Owner + "/" + source.Repo
		if source.Ref != "" && source.Ref != "main" && source.Ref != "master" {
			s += "#" + source.Ref
		}
		return s
	case "gitlab":
		s := source.Owner + "/" + source.Repo
		if source.Ref != "" && source.Ref != "main" && source.Ref != "master" {
			s += "#" + source.Ref
		}
		return s
	case "bitbucket":
		s := source.Owner + "/" + source.Repo
		if source.Ref != "" && source.Ref != "main" && source.Ref != "master" {
			s += "#" + source.Ref
		}
		return s
	case "local":
		return source.LocalPath
	case "url", "well-known", "zip":
		return source.URL
	default:
		return source.URL
	}
}

// UpdateSkillMeta updates an existing SkillMeta with new content info
func UpdateSkillMeta(meta *SkillMeta, content string, gitInfo *GitCommitInfo) {
	meta.ContentHash = ComputeContentHash(content)
	meta.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if gitInfo != nil {
		meta.CommitHash = gitInfo.Hash
		meta.CommitDate = gitInfo.Date
	}
}

// LoadSkillWithMeta loads a skill and its metadata from a skill directory
func LoadSkillWithMeta(skillDir string) (*Skill, error) {
	skillPath := skillDir + "/" + SkillFileName
	metaPath := skillDir + "/" + MetaFileName

	skill, err := ParseSkillMd(skillPath)
	if err != nil {
		return nil, err
	}

	// Try to load metadata (optional)
	if meta, err := ReadSkillMeta(metaPath); err == nil {
		skill.Meta = meta
	}

	return skill, nil
}

// ListSkillsWithMeta reads all skills from a scrolls directory with their metadata
func ListSkillsWithMeta(scrollsDir string) ([]*Skill, error) {
	entries, err := os.ReadDir(scrollsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Skill{}, nil
		}
		return nil, err
	}

	var skills []*Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillDir := scrollsDir + "/" + entry.Name()
		skill, err := LoadSkillWithMeta(skillDir)
		if err != nil {
			continue // Skip invalid skills
		}

		skills = append(skills, skill)
	}

	return skills, nil
}

// SaveSkillWithMeta saves a skill's SKILL.md and .scribe-meta.json to a directory
func SaveSkillWithMeta(skillDir string, skill *Skill, source *SourceInfo, skillPathInSource string, gitInfo *GitCommitInfo) error {
	// Ensure directory exists
	if err := EnsureDir(skillDir); err != nil {
		return err
	}

	// Read original SKILL.md content to preserve formatting
	originalContent, err := os.ReadFile(skill.Path + "/" + SkillFileName)
	if err != nil {
		return err
	}

	// Write SKILL.md
	skillPath := skillDir + "/" + SkillFileName
	if err := os.WriteFile(skillPath, originalContent, 0o644); err != nil {
		return err
	}

	// Create and write metadata
	meta := NewSkillMeta(source, skillPathInSource, string(originalContent), gitInfo)
	metaPath := skillDir + "/" + MetaFileName
	if err := WriteSkillMeta(metaPath, meta); err != nil {
		return err
	}

	return nil
}

// SkillNeedsUpdate checks if a skill's content has changed from the stored hash
func SkillNeedsUpdate(skillDir, newContent string) (bool, error) {
	metaPath := skillDir + "/" + MetaFileName
	meta, err := ReadSkillMeta(metaPath)
	if err != nil {
		// If no metadata exists, consider it as needing update
		return true, nil
	}

	newHash := ComputeContentHash(newContent)
	return meta.ContentHash != newHash, nil
}
