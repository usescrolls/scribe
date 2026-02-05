package scribe

import (
	"os"
	"path/filepath"
	"strings"
)

// AllAgents contains definitions for all 40+ supported coding agents
// Based on Vercel's skills CLI agent definitions
var AllAgents = []Agent{
	// Tier 1: Major AI Coding Assistants
	{
		ID:              "claude-code",
		DisplayName:     "Claude Code",
		GlobalSkillsDir: "~/.claude/skills",
		GlobalConfigDir: "~/.claude",
	},
	{
		ID:              "cursor",
		DisplayName:     "Cursor",
		GlobalSkillsDir: "~/.cursor/skills",
		GlobalConfigDir: "~/.cursor",
	},
	{
		ID:              "github-copilot",
		DisplayName:     "GitHub Copilot",
		GlobalSkillsDir: "~/.copilot/skills",
		GlobalConfigDir: "~/.copilot",
	},
	{
		ID:              "cline",
		DisplayName:     "Cline",
		GlobalSkillsDir: "~/.cline/skills",
		GlobalConfigDir: "~/.cline",
	},
	{
		ID:              "continue",
		DisplayName:     "Continue",
		GlobalSkillsDir: "~/.continue/skills",
		GlobalConfigDir: "~/.continue",
	},
	{
		ID:              "windsurf",
		DisplayName:     "Windsurf",
		GlobalSkillsDir: "~/.codeium/windsurf/skills",
		GlobalConfigDir: "~/.codeium/windsurf",
	},
	{
		ID:              "codex",
		DisplayName:     "Codex CLI",
		GlobalSkillsDir: "~/.codex/skills",
		GlobalConfigDir: "~/.codex",
	},
	{
		ID:              "gemini",
		DisplayName:     "Gemini CLI",
		GlobalSkillsDir: "~/.gemini/skills",
		GlobalConfigDir: "~/.gemini",
	},
	{
		ID:              "goose",
		DisplayName:     "Goose",
		GlobalSkillsDir: "~/.config/goose/skills",
		GlobalConfigDir: "~/.config/goose",
	},
	{
		ID:              "opencode",
		DisplayName:     "OpenCode",
		GlobalSkillsDir: "~/.config/opencode/skills",
		GlobalConfigDir: "~/.config/opencode",
	},

	// Tier 2: VS Code Extensions & IDE Integrations
	{
		ID:              "aider",
		DisplayName:     "Aider",
		GlobalSkillsDir: "~/.aider/skills",
		GlobalConfigDir: "~/.aider",
	},
	{
		ID:              "roo-code",
		DisplayName:     "Roo Code",
		GlobalSkillsDir: "~/.roo/skills",
		GlobalConfigDir: "~/.roo",
	},
	{
		ID:              "void",
		DisplayName:     "Void",
		GlobalSkillsDir: "~/.void/skills",
		GlobalConfigDir: "~/.void",
	},
	{
		ID:              "zed",
		DisplayName:     "Zed",
		GlobalSkillsDir: "~/.config/zed/skills",
		GlobalConfigDir: "~/.config/zed",
	},
	{
		ID:              "sourcegraph-cody",
		DisplayName:     "Sourcegraph Cody",
		GlobalSkillsDir: "~/.cody/skills",
		GlobalConfigDir: "~/.cody",
	},
	{
		ID:              "tabnine",
		DisplayName:     "Tabnine",
		GlobalSkillsDir: "~/.tabnine/skills",
		GlobalConfigDir: "~/.tabnine",
	},
	{
		ID:              "codeium",
		DisplayName:     "Codeium",
		GlobalSkillsDir: "~/.codeium/skills",
		GlobalConfigDir: "~/.codeium",
	},
	{
		ID:              "amazon-q",
		DisplayName:     "Amazon Q Developer",
		GlobalSkillsDir: "~/.amazonq/skills",
		GlobalConfigDir: "~/.amazonq",
	},

	// Tier 3: Terminal/CLI Tools
	{
		ID:              "warp",
		DisplayName:     "Warp AI",
		GlobalSkillsDir: "~/.warp/skills",
		GlobalConfigDir: "~/.warp",
	},
	{
		ID:              "fig",
		DisplayName:     "Fig",
		GlobalSkillsDir: "~/.fig/skills",
		GlobalConfigDir: "~/.fig",
	},
	{
		ID:              "aichat",
		DisplayName:     "AIChat",
		GlobalSkillsDir: "~/.config/aichat/skills",
		GlobalConfigDir: "~/.config/aichat",
	},
	{
		ID:              "shell-gpt",
		DisplayName:     "ShellGPT",
		GlobalSkillsDir: "~/.config/shell_gpt/skills",
		GlobalConfigDir: "~/.config/shell_gpt",
	},

	// Tier 4: Autonomous Agents
	{
		ID:              "devin",
		DisplayName:     "Devin",
		GlobalSkillsDir: "~/.devin/skills",
		GlobalConfigDir: "~/.devin",
	},
	{
		ID:              "mentat",
		DisplayName:     "Mentat",
		GlobalSkillsDir: "~/.mentat/skills",
		GlobalConfigDir: "~/.mentat",
	},
	{
		ID:              "sweep",
		DisplayName:     "Sweep",
		GlobalSkillsDir: "~/.sweep/skills",
		GlobalConfigDir: "~/.sweep",
	},
	{
		ID:              "gpt-engineer",
		DisplayName:     "GPT Engineer",
		GlobalSkillsDir: "~/.gpt-engineer/skills",
		GlobalConfigDir: "~/.gpt-engineer",
	},
	{
		ID:              "smol-developer",
		DisplayName:     "Smol Developer",
		GlobalSkillsDir: "~/.smol-developer/skills",
		GlobalConfigDir: "~/.smol-developer",
	},
	{
		ID:              "gpt-pilot",
		DisplayName:     "GPT Pilot",
		GlobalSkillsDir: "~/.gpt-pilot/skills",
		GlobalConfigDir: "~/.gpt-pilot",
	},
	{
		ID:              "auto-gpt",
		DisplayName:     "AutoGPT",
		GlobalSkillsDir: "~/.auto-gpt/skills",
		GlobalConfigDir: "~/.auto-gpt",
	},
	{
		ID:              "agent-gpt",
		DisplayName:     "AgentGPT",
		GlobalSkillsDir: "~/.agent-gpt/skills",
		GlobalConfigDir: "~/.agent-gpt",
	},

	// Tier 5: Specialized Tools
	{
		ID:              "pr-agent",
		DisplayName:     "PR Agent",
		GlobalSkillsDir: "~/.pr-agent/skills",
		GlobalConfigDir: "~/.pr-agent",
	},
	{
		ID:              "what-the-diff",
		DisplayName:     "What The Diff",
		GlobalSkillsDir: "~/.what-the-diff/skills",
		GlobalConfigDir: "~/.what-the-diff",
	},
	{
		ID:              "codeball",
		DisplayName:     "Codeball",
		GlobalSkillsDir: "~/.codeball/skills",
		GlobalConfigDir: "~/.codeball",
	},
	{
		ID:              "coderabbit",
		DisplayName:     "CodeRabbit",
		GlobalSkillsDir: "~/.coderabbit/skills",
		GlobalConfigDir: "~/.coderabbit",
	},
	{
		ID:              "bito",
		DisplayName:     "Bito",
		GlobalSkillsDir: "~/.bito/skills",
		GlobalConfigDir: "~/.bito",
	},

	// Tier 6: JetBrains & Other IDEs
	{
		ID:              "jetbrains-ai",
		DisplayName:     "JetBrains AI",
		GlobalSkillsDir: "~/.config/JetBrains/skills",
		GlobalConfigDir: "~/.config/JetBrains",
	},
	{
		ID:              "supermaven",
		DisplayName:     "Supermaven",
		GlobalSkillsDir: "~/.supermaven/skills",
		GlobalConfigDir: "~/.supermaven",
	},

	// Tier 7: Open Source Alternatives
	{
		ID:              "ollama",
		DisplayName:     "Ollama",
		GlobalSkillsDir: "~/.ollama/skills",
		GlobalConfigDir: "~/.ollama",
	},
	{
		ID:              "jan",
		DisplayName:     "Jan",
		GlobalSkillsDir: "~/.jan/skills",
		GlobalConfigDir: "~/.jan",
	},
	{
		ID:              "llm",
		DisplayName:     "LLM CLI",
		GlobalSkillsDir: "~/.config/io.datasette.llm/skills",
		GlobalConfigDir: "~/.config/io.datasette.llm",
	},
	{
		ID:              "fabric",
		DisplayName:     "Fabric",
		GlobalSkillsDir: "~/.config/fabric/skills",
		GlobalConfigDir: "~/.config/fabric",
	},

	// Tier 8: Enterprise Tools
	{
		ID:              "pieces",
		DisplayName:     "Pieces for Developers",
		GlobalSkillsDir: "~/.pieces/skills",
		GlobalConfigDir: "~/.pieces",
	},
	{
		ID:              "mintlify",
		DisplayName:     "Mintlify Writer",
		GlobalSkillsDir: "~/.mintlify/skills",
		GlobalConfigDir: "~/.mintlify",
	},
	{
		ID:              "blackbox",
		DisplayName:     "Blackbox AI",
		GlobalSkillsDir: "~/.blackbox/skills",
		GlobalConfigDir: "~/.blackbox",
	},
	{
		ID:              "codegpt",
		DisplayName:     "CodeGPT",
		GlobalSkillsDir: "~/.codegpt/skills",
		GlobalConfigDir: "~/.codegpt",
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
