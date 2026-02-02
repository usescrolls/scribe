package scribe

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
)

var (
	ErrNoSkillsFound     = errors.New("no skills found in source")
	ErrInvalidSkill      = errors.New("invalid skill: missing required fields")
	ErrSkillNotFound     = errors.New("skill not found")
	ErrMissingName       = errors.New("skill missing required 'name' field in frontmatter")
	ErrMissingDesc       = errors.New("skill missing required 'description' field in frontmatter")
)

// Frontmatter patterns for parsing SKILL.md
var (
	frontmatterPattern = regexp.MustCompile(`(?s)^---\n(.+?)\n---\n?(.*)$`)
	namePattern        = regexp.MustCompile(`[^a-z0-9-]`)
)

// SkillFrontmatter represents the YAML frontmatter in a SKILL.md file
type SkillFrontmatter struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Metadata    map[string]any `yaml:",inline"` // Capture all other fields
}

// ParseSkillMd parses a SKILL.md file and returns a Skill
// The path should be the full path to the SKILL.md file
func ParseSkillMd(path string) (*Skill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill file: %w", err)
	}

	return ParseSkillContent(string(content), filepath.Dir(path))
}

// ParseSkillContent parses SKILL.md content and returns a Skill
func ParseSkillContent(content string, skillDir string) (*Skill, error) {
	// Extract frontmatter and body
	matches := frontmatterPattern.FindStringSubmatch(content)
	if matches == nil {
		return nil, fmt.Errorf("invalid skill format: no frontmatter found")
	}

	frontmatterYAML := matches[1]
	body := strings.TrimSpace(matches[2])

	// Parse frontmatter
	var fm SkillFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Validate required fields
	if fm.Name == "" {
		return nil, ErrMissingName
	}
	if fm.Description == "" {
		return nil, ErrMissingDesc
	}

	// Remove name and description from metadata (they're already extracted)
	delete(fm.Metadata, "name")
	delete(fm.Metadata, "description")

	skill := &Skill{
		Name:        fm.Name,
		Description: fm.Description,
		Path:        skillDir,
		Content:     body,
		Metadata:    fm.Metadata,
	}

	return skill, nil
}

// ValidateSkill ensures a skill has all required fields
func ValidateSkill(skill *Skill) error {
	if skill == nil {
		return ErrInvalidSkill
	}
	if skill.Name == "" {
		return ErrMissingName
	}
	if skill.Description == "" {
		return ErrMissingDesc
	}
	return nil
}

// SanitizeName converts a string to a valid skill name (lowercase kebab-case)
func SanitizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	// Replace spaces and underscores with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	// Remove invalid characters
	name = namePattern.ReplaceAllString(name, "")
	// Remove consecutive hyphens
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")
	// Limit length
	if len(name) > 255 {
		name = name[:255]
	}
	return name
}

// DiscoverSkills finds all SKILL.md files in a directory
// It searches recursively up to maxDepth, excluding common non-skill directories
func DiscoverSkills(root string) ([]*Skill, error) {
	return DiscoverSkillsWithDepth(root, 5)
}

// DiscoverSkillsWithDepth finds all SKILL.md files with a custom max depth
func DiscoverSkillsWithDepth(root string, maxDepth int) ([]*Skill, error) {
	var skills []*Skill

	// Directories to skip during discovery
	skipDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		"dist":         true,
		"build":        true,
		"vendor":       true,
		"__pycache__":  true,
		".venv":        true,
		"venv":         true,
		".tox":         true,
		"target":       true, // Rust/Java builds
	}

	// Check for legacy plugin.json (for warning purposes only)
	legacyPluginPath := filepath.Join(root, ".claude-plugin", "plugin.json")
	if _, err := os.Stat(legacyPluginPath); err == nil {
		// Legacy plugin format detected - just log and continue
		// The recursive search will find any SKILL.md files in skills/ subdirectory
	}

	// First, check for SKILL.md directly in root (single-skill repo)
	rootSkillPath := filepath.Join(root, SkillFileName)
	if skill, err := ParseSkillMd(rootSkillPath); err == nil {
		skills = append(skills, skill)
	}

	// Search common directories first
	commonDirs := []string{"skills", ".claude/skills", ".cursor/skills", ".github/skills"}
	for _, dir := range commonDirs {
		dirPath := filepath.Join(root, dir)
		if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
			found, _ := discoverSkillsInDir(dirPath, 2) // Shallow search in known dirs
			skills = append(skills, found...)
		}
	}

	// Recursive search for remaining skills
	rootDepth := strings.Count(root, string(os.PathSeparator))

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}

		// Skip excluded directories
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			// Check depth
			currentDepth := strings.Count(path, string(os.PathSeparator)) - rootDepth
			if currentDepth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		// Look for SKILL.md files
		if d.Name() == SkillFileName {
			skill, err := ParseSkillMd(path)
			if err != nil {
				return nil // Skip invalid skills
			}
			// Avoid duplicates from root and common dirs
			if !skillInList(skills, skill.Name) {
				skills = append(skills, skill)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(skills) == 0 {
		return nil, ErrNoSkillsFound
	}

	return skills, nil
}

// discoverSkillsInDir finds skills in a specific directory with limited depth
func discoverSkillsInDir(dir string, maxDepth int) ([]*Skill, error) {
	var skills []*Skill

	baseDepth := strings.Count(dir, string(os.PathSeparator))

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			currentDepth := strings.Count(path, string(os.PathSeparator)) - baseDepth
			if currentDepth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		if d.Name() == SkillFileName {
			skill, err := ParseSkillMd(path)
			if err != nil {
				return nil
			}
			skills = append(skills, skill)
		}

		return nil
	})

	return skills, err
}

// skillInList checks if a skill with the given name is already in the list
func skillInList(skills []*Skill, name string) bool {
	for _, s := range skills {
		if s.Name == name {
			return true
		}
	}
	return false
}

// ReadSkill reads a skill from the canonical storage location
func ReadSkill(skillName string) (*Skill, error) {
	skillPath, err := GetSkillPath(skillName)
	if err != nil {
		return nil, err
	}

	skill, err := ParseSkillMd(skillPath)
	if err != nil {
		return nil, err
	}

	// Load metadata if available
	metaPath, err := GetMetaPath(skillName)
	if err != nil {
		return skill, nil // Return skill without meta
	}

	meta, err := ReadSkillMeta(metaPath)
	if err == nil {
		skill.Meta = meta
	}

	return skill, nil
}

// ReadAllSkills reads all skills from the canonical storage location
func ReadAllSkills() ([]*Skill, error) {
	names, err := ListInstalledSkills()
	if err != nil {
		return nil, err
	}

	var skills []*Skill
	for _, name := range names {
		skill, err := ReadSkill(name)
		if err != nil {
			continue // Skip invalid skills
		}
		skills = append(skills, skill)
	}

	return skills, nil
}

// GetSkillInfo converts a Skill to a SkillInfo for the frontend
func GetSkillInfo(skill *Skill) SkillInfo {
	info := SkillInfo{
		Name:        skill.Name,
		Description: skill.Description,
		Agents:      []string{},
	}

	if skill.Meta != nil {
		info.Source = skill.Meta.Source
		info.SourceType = skill.Meta.SourceType
		info.InstalledAt = skill.Meta.InstalledAt
	}

	return info
}

// GetAllSkillInfo returns SkillInfo for all installed skills
func GetAllSkillInfo() ([]SkillInfo, error) {
	skills, err := ReadAllSkills()
	if err != nil {
		return nil, err
	}

	infos := make([]SkillInfo, len(skills))
	for i, skill := range skills {
		infos[i] = GetSkillInfo(skill)
		// Find which agents have this skill
		infos[i].Agents = getAgentsWithSkill(skill.Name)
	}

	return infos, nil
}

// GetAgentsWithSkill returns a list of agent IDs that have the given skill installed
func GetAgentsWithSkill(skillName string) []string {
	return getAgentsWithSkill(skillName)
}

// getAgentsWithSkill returns a list of agent IDs that have the given skill installed
func getAgentsWithSkill(skillName string) []string {
	var agents []string
	for _, agent := range DetectInstalledAgents() {
		skillsDir := expandPath(agent.GlobalSkillsDir)
		skillPath := filepath.Join(skillsDir, skillName, SkillFileName)
		if _, err := os.Stat(skillPath); err == nil {
			agents = append(agents, agent.ID)
		}
	}
	return agents
}
