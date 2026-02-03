package scribe

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	// ScribeDirName is the directory name for Scribe data
	ScribeDirName = ".scribe"
	// ScrollsDirName is the subdirectory for skill storage
	ScrollsDirName = "scrolls"
	// WorkspacesDirName is the subdirectory for workspace definitions
	WorkspacesDirName = "workspaces"
	// ConfigFileName is the global config file name
	ConfigFileName = "config.json"
	// SkillFileName is the standard skill definition file name
	SkillFileName = "SKILL.md"
	// MetaFileName is the sidecar metadata file name
	MetaFileName = ".scribe-meta.json"
	// DefaultWorkspaceName is the name of the default workspace
	DefaultWorkspaceName = "default"
)

// GetScribeDir returns the global Scribe directory (~/.scribe)
func GetScribeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ScribeDirName), nil
}

// GetScrollsDir returns the global scrolls directory (~/.scribe/scrolls/)
// All skills are managed globally - no project-level skills
func GetScrollsDir() (string, error) {
	scribeDir, err := GetScribeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(scribeDir, ScrollsDirName), nil
}

// GetWorkspacesDir returns the workspaces directory (~/.scribe/workspaces/)
func GetWorkspacesDir() (string, error) {
	scribeDir, err := GetScribeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(scribeDir, WorkspacesDirName), nil
}

// GetConfigPath returns the global config file path (~/.scribe/config.json)
func GetConfigPath() (string, error) {
	scribeDir, err := GetScribeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(scribeDir, ConfigFileName), nil
}

// GetSkillDir returns the directory for a specific skill
func GetSkillDir(skillName string) (string, error) {
	scrollsDir, err := GetScrollsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(scrollsDir, skillName), nil
}

// GetSkillPath returns the SKILL.md path for a specific skill
func GetSkillPath(skillName string) (string, error) {
	skillDir, err := GetSkillDir(skillName)
	if err != nil {
		return "", err
	}
	return filepath.Join(skillDir, SkillFileName), nil
}

// GetMetaPath returns the .scribe-meta.json path for a specific skill
func GetMetaPath(skillName string) (string, error) {
	skillDir, err := GetSkillDir(skillName)
	if err != nil {
		return "", err
	}
	return filepath.Join(skillDir, MetaFileName), nil
}

// GetWorkspacePath returns the path for a specific workspace file
func GetWorkspacePath(name string) (string, error) {
	workspacesDir, err := GetWorkspacesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(workspacesDir, name+".json"), nil
}

// EnsureDir creates a directory and all parent directories if they don't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// EnsureScribeDirs creates all required Scribe directories
func EnsureScribeDirs() error {
	scribeDir, err := GetScribeDir()
	if err != nil {
		return err
	}

	dirs := []string{
		scribeDir,
		filepath.Join(scribeDir, ScrollsDirName),
		filepath.Join(scribeDir, WorkspacesDirName),
	}

	for _, dir := range dirs {
		if err := EnsureDir(dir); err != nil {
			return err
		}
	}

	return nil
}

// LoadConfig loads the global config from ~/.scribe/config.json
// Returns a default config if the file doesn't exist
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config
			return &Config{
				ActiveWorkspace: DefaultWorkspaceName,
			}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the global config to ~/.scribe/config.json
func SaveConfig(config *Config) error {
	if err := EnsureScribeDirs(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// ListInstalledSkills returns the names of all installed skills in the global scrolls directory
func ListInstalledSkills() ([]string, error) {
	scrollsDir, err := GetScrollsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(scrollsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var skills []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Check if SKILL.md exists in the directory
		skillPath := filepath.Join(scrollsDir, entry.Name(), SkillFileName)
		if _, err := os.Stat(skillPath); err == nil {
			skills = append(skills, entry.Name())
		}
	}

	return skills, nil
}

// SkillExists checks if a skill exists in the global scrolls directory
func SkillExists(skillName string) (bool, error) {
	skillPath, err := GetSkillPath(skillName)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(skillPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
