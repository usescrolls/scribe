package scribe

import (
	"os"
	"path/filepath"
	"strings"
)

// AllAgents contains definitions for all supported coding agents.
// Synced with Vercel skills CLI agent definitions (skills/src/agents.ts).
var AllAgents = []Agent{
	{
		ID:              "claude-code",
		DisplayName:     "Claude Code",
		GlobalSkillsDir: "~/.claude/skills",
		GlobalConfigDir: "~/.claude",
	},
	{
		ID:              "amp",
		DisplayName:     "Amp",
		GlobalSkillsDir: "~/.config/agents/skills",
		GlobalConfigDir: "~/.config/amp",
	},
	{
		ID:              "antigravity",
		DisplayName:     "Antigravity",
		GlobalSkillsDir: "~/.gemini/antigravity/skills",
		GlobalConfigDir: "~/.gemini/antigravity",
	},
	{
		ID:              "augment",
		DisplayName:     "Augment",
		GlobalSkillsDir: "~/.augment/rules",
		GlobalConfigDir: "~/.augment",
	},
	{
		ID:              "openclaw",
		DisplayName:     "OpenClaw",
		GlobalSkillsDir: "~/.openclaw/skills",
		GlobalConfigDir: "~/.openclaw",
	},
	{
		ID:              "cline",
		DisplayName:     "Cline",
		GlobalSkillsDir: "~/.cline/skills",
		GlobalConfigDir: "~/.cline",
	},
	{
		ID:              "codebuddy",
		DisplayName:     "CodeBuddy",
		GlobalSkillsDir: "~/.codebuddy/skills",
		GlobalConfigDir: "~/.codebuddy",
	},
	{
		ID:              "codex",
		DisplayName:     "Codex",
		GlobalSkillsDir: "~/.codex/skills",
		GlobalConfigDir: "~/.codex",
	},
	{
		ID:              "command-code",
		DisplayName:     "Command Code",
		GlobalSkillsDir: "~/.commandcode/skills",
		GlobalConfigDir: "~/.commandcode",
	},
	{
		ID:              "continue",
		DisplayName:     "Continue",
		GlobalSkillsDir: "~/.continue/skills",
		GlobalConfigDir: "~/.continue",
	},
	{
		ID:              "crush",
		DisplayName:     "Crush",
		GlobalSkillsDir: "~/.config/crush/skills",
		GlobalConfigDir: "~/.config/crush",
	},
	{
		ID:              "cursor",
		DisplayName:     "Cursor",
		GlobalSkillsDir: "~/.cursor/skills",
		GlobalConfigDir: "~/.cursor",
	},
	{
		ID:              "droid",
		DisplayName:     "Droid",
		GlobalSkillsDir: "~/.factory/skills",
		GlobalConfigDir: "~/.factory",
	},
	{
		ID:              "gemini-cli",
		DisplayName:     "Gemini CLI",
		GlobalSkillsDir: "~/.gemini/skills",
		GlobalConfigDir: "~/.gemini",
	},
	{
		ID:              "github-copilot",
		DisplayName:     "GitHub Copilot",
		GlobalSkillsDir: "~/.copilot/skills",
		GlobalConfigDir: "~/.copilot",
	},
	{
		ID:              "goose",
		DisplayName:     "Goose",
		GlobalSkillsDir: "~/.config/goose/skills",
		GlobalConfigDir: "~/.config/goose",
	},
	{
		ID:              "iflow-cli",
		DisplayName:     "iFlow CLI",
		GlobalSkillsDir: "~/.iflow/skills",
		GlobalConfigDir: "~/.iflow",
	},
	{
		ID:              "junie",
		DisplayName:     "Junie",
		GlobalSkillsDir: "~/.junie/skills",
		GlobalConfigDir: "~/.junie",
	},
	{
		ID:              "kilo",
		DisplayName:     "Kilo Code",
		GlobalSkillsDir: "~/.kilocode/skills",
		GlobalConfigDir: "~/.kilocode",
	},
	{
		ID:              "kimi-cli",
		DisplayName:     "Kimi Code CLI",
		GlobalSkillsDir: "~/.config/agents/skills",
		GlobalConfigDir: "~/.kimi",
	},
	{
		ID:              "kiro-cli",
		DisplayName:     "Kiro CLI",
		GlobalSkillsDir: "~/.kiro/skills",
		GlobalConfigDir: "~/.kiro",
	},
	{
		ID:              "kode",
		DisplayName:     "Kode",
		GlobalSkillsDir: "~/.kode/skills",
		GlobalConfigDir: "~/.kode",
	},
	{
		ID:              "mcpjam",
		DisplayName:     "MCPJam",
		GlobalSkillsDir: "~/.mcpjam/skills",
		GlobalConfigDir: "~/.mcpjam",
	},
	{
		ID:              "mistral-vibe",
		DisplayName:     "Mistral Vibe",
		GlobalSkillsDir: "~/.vibe/skills",
		GlobalConfigDir: "~/.vibe",
	},
	{
		ID:              "mux",
		DisplayName:     "Mux",
		GlobalSkillsDir: "~/.mux/skills",
		GlobalConfigDir: "~/.mux",
	},
	{
		ID:              "opencode",
		DisplayName:     "OpenCode",
		GlobalSkillsDir: "~/.config/opencode/skills",
		GlobalConfigDir: "~/.config/opencode",
	},
	{
		ID:              "openhands",
		DisplayName:     "OpenHands",
		GlobalSkillsDir: "~/.openhands/skills",
		GlobalConfigDir: "~/.openhands",
	},
	{
		ID:              "pi",
		DisplayName:     "Pi",
		GlobalSkillsDir: "~/.pi/agent/skills",
		GlobalConfigDir: "~/.pi/agent",
	},
	{
		ID:              "qoder",
		DisplayName:     "Qoder",
		GlobalSkillsDir: "~/.qoder/skills",
		GlobalConfigDir: "~/.qoder",
	},
	{
		ID:              "qwen-code",
		DisplayName:     "Qwen Code",
		GlobalSkillsDir: "~/.qwen/skills",
		GlobalConfigDir: "~/.qwen",
	},
	{
		ID:              "replit",
		DisplayName:     "Replit",
		GlobalSkillsDir: "~/.config/agents/skills",
		GlobalConfigDir: "~/.config/agents",
	},
	{
		ID:              "roo",
		DisplayName:     "Roo Code",
		GlobalSkillsDir: "~/.roo/skills",
		GlobalConfigDir: "~/.roo",
	},
	{
		ID:              "trae",
		DisplayName:     "Trae",
		GlobalSkillsDir: "~/.trae/skills",
		GlobalConfigDir: "~/.trae",
	},
	{
		ID:              "trae-cn",
		DisplayName:     "Trae CN",
		GlobalSkillsDir: "~/.trae-cn/skills",
		GlobalConfigDir: "~/.trae-cn",
	},
	{
		ID:              "windsurf",
		DisplayName:     "Windsurf",
		GlobalSkillsDir: "~/.codeium/windsurf/skills",
		GlobalConfigDir: "~/.codeium/windsurf",
	},
	{
		ID:              "zencoder",
		DisplayName:     "Zencoder",
		GlobalSkillsDir: "~/.zencoder/skills",
		GlobalConfigDir: "~/.zencoder",
	},
	{
		ID:              "neovate",
		DisplayName:     "Neovate",
		GlobalSkillsDir: "~/.neovate/skills",
		GlobalConfigDir: "~/.neovate",
	},
	{
		ID:              "pochi",
		DisplayName:     "Pochi",
		GlobalSkillsDir: "~/.pochi/skills",
		GlobalConfigDir: "~/.pochi",
	},
	{
		ID:              "adal",
		DisplayName:     "AdaL",
		GlobalSkillsDir: "~/.adal/skills",
		GlobalConfigDir: "~/.adal",
	},
}

// agentsByID is a map for fast agent lookup
var agentsByID map[string]*Agent

func init() {
	agentsByID = make(map[string]*Agent, len(AllAgents))
	for i := range AllAgents {
		agentsByID[AllAgents[i].ID] = &AllAgents[i]
	}
}

// GetAgent returns an agent by its ID, or nil if not found
func GetAgent(id string) *Agent {
	return agentsByID[id]
}

// GetAllAgents returns a copy of all agent definitions
func GetAllAgents() []Agent {
	agents := make([]Agent, len(AllAgents))
	copy(agents, AllAgents)
	return agents
}

// DetectInstalledAgents returns agents that have their global config directory present
func DetectInstalledAgents() []Agent {
	var installed []Agent
	for _, agent := range AllAgents {
		configDir := expandPath(agent.GlobalConfigDir)
		if dirExists(configDir) {
			installed = append(installed, agent)
		}
	}
	return installed
}

// GetAgentStatus returns the status of all agents for the frontend
func GetAgentStatus(scrollsDir string) []AgentStatus {
	statuses := make([]AgentStatus, len(AllAgents))
	for i, agent := range AllAgents {
		configDir := expandPath(agent.GlobalConfigDir)
		skillsDir := expandPath(agent.GlobalSkillsDir)

		installed := dirExists(configDir)
		skillCount := 0
		if installed && dirExists(skillsDir) {
			skillCount = countSkillsInDir(skillsDir)
		}

		statuses[i] = AgentStatus{
			ID:              agent.ID,
			DisplayName:     agent.DisplayName,
			Installed:       installed,
			SkillCount:      skillCount,
			GlobalSkillsDir: skillsDir,
		}
	}
	return statuses
}

// ExpandAgentPath expands ~ in agent paths to the user's home directory
func ExpandAgentPath(path string) string {
	return expandPath(path)
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// countSkillsInDir counts the number of skill directories (containing SKILL.md) in a directory.
// It follows symlinks, since skills in agent directories are typically symlinked from ~/.scribe/scrolls/.
func countSkillsInDir(dir string) int {
	count := 0
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		// Use os.Stat (not Lstat) to follow symlinks
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		if info.IsDir() {
			skillPath := filepath.Join(fullPath, "SKILL.md")
			if _, err := os.Stat(skillPath); err == nil {
				count++
			}
		}
	}
	return count
}
