package scribe

import (
	"os"
	"path/filepath"
	"time"
)

// SystemSkillName is the canonical name for the system skill
const SystemSkillName = "scribe-cli"

// SystemSkillContent is the SKILL.md content for the system skill.
// This teaches agents all Scribe CLI commands.
const SystemSkillContent = `---
name: scribe-cli
description: Exhaustive reference for the Scribe CLI - a skills manager for coding agents
---

# Scribe CLI Reference

Scribe manages AI coding skills across 45+ agents. Install once, use everywhere.

## Core Commands

### scribe install <source>
Install skills from a source. Sources can be:
- GitHub shorthand: ` + "`scribe install owner/repo`" + `
- Full URL: ` + "`scribe install https://github.com/owner/repo`" + `
- Local directory: ` + "`scribe install ./my-skills`" + `

**Flags:**
- ` + "`-s, --skill <names>`" + ` - Select specific skills to install (comma-separated)
- ` + "`-l, --list`" + ` - List available skills without installing
- ` + "`-y, --yes`" + ` - Skip interactive prompts
- ` + "`--all`" + ` - Install all skills to all detected agents

**Outcome:** Skills are copied to ~/.scribe/scrolls/<name>/, symlinked to all detected agents, and added to the active + default workspace.

### scribe uninstall <skill-name>
Remove a skill from canonical storage, all agents, and all workspaces.

**Flags:**
- ` + "`--all`" + ` - Remove all installed skills

**Aliases:** ` + "`remove`" + `, ` + "`rm`" + `

**Outcome:** Symlinks removed from all agents, skill directory deleted from ~/.scribe/scrolls/, skill removed from all workspace definitions.

### scribe list
List all installed skills.

**Flags:**
- ` + "`--names-only`" + ` - Print only skill names, one per line
- ` + "`--json`" + ` - Output in JSON format

**Aliases:** ` + "`ls`" + `

**Outcome:** Displays a table with name, description, source, install date, and agents.

### scribe info <skill-name>
Show detailed information about an installed skill.

**Flags:**
- ` + "`--json`" + ` - Output in JSON format

**Outcome:** Shows full metadata including source, content hash, commit info, and which agents have the skill.

### scribe check [skill-name]
Check for available updates from remote sources.

**Outcome:** Compares local content hash with remote. Shows outdated/up-to-date status for each skill.

### scribe update [skill-name]
Update skills to latest version from original source.

**Flags:**
- ` + "`--all`" + ` - Update all installed skills
- ` + "`--force`" + ` - Force update even if content hash matches

**Outcome:** Fetches latest content from original source and replaces local copy. Syncs to all agents.

## Workspace Commands

### scribe workspace list
List all workspaces with skill counts and active status.

### scribe workspace create <name>
Create a new empty workspace.
- ` + "`-d, --description <text>`" + ` - Workspace description

### scribe workspace use <name>
Switch to a different workspace. Updates symlinks across all agents.

**Outcome:** Skills in the previous workspace are un-symlinked; skills in the target workspace are symlinked.

### scribe workspace add <skill-name>
Add an installed skill to the current workspace.

### scribe workspace remove <skill-name>
Remove a skill from the current workspace (does not uninstall).

### scribe workspace current
Show the active workspace and its skills.

### scribe workspace delete <name>
Delete a workspace (cannot delete "default").

## Cache Commands

### scribe cache clear
Clear the git clone cache used for fetching skills.

### scribe cache info
Show cache size and contents.

## Global Flags

- ` + "`--debug`" + ` - Enable debug logging
- ` + "`--json`" + ` - Output in JSON format (where applicable)
- ` + "`-q, --quiet`" + ` - Suppress non-essential output

## File Layout

` + "```" + `
~/.scribe/
  config.json            # Active workspace, onboarding status
  scrolls/               # Canonical skill storage
    <skill-name>/
      SKILL.md           # Skill content (YAML frontmatter + markdown)
      .scribe-meta.json  # Source tracking, content hash, timestamps
  workspaces/            # Workspace definitions
    default.json         # Default workspace (all skills)
    <name>.json          # Custom workspaces
  cache/                 # Git clone cache
` + "```" + `
`

// IsSystemSkill returns true if the given skill name is a system skill.
func IsSystemSkill(name string) bool {
	return name == SystemSkillName
}

// EnsureSystemSkill installs or updates the system skill on disk and ensures
// symlinks exist in all detected agents. Call this during CLI and GUI startup.
func EnsureSystemSkill() error {
	skillDir, err := GetSkillDir(SystemSkillName)
	if err != nil {
		return err
	}

	skillPath := filepath.Join(skillDir, SkillFileName)

	// Check if content needs updating
	existingContent, err := os.ReadFile(skillPath)
	if err != nil || string(existingContent) != SystemSkillContent {
		// Create or update the skill
		if err := EnsureDir(skillDir); err != nil {
			return err
		}

		if err := os.WriteFile(skillPath, []byte(SystemSkillContent), 0o644); err != nil {
			return err
		}

		// Write/update metadata
		metaPath := filepath.Join(skillDir, MetaFileName)
		now := time.Now().Format(time.RFC3339)
		meta := &SkillMeta{
			Source:      "scribe",
			SourceType:  "system",
			ContentHash: ComputeContentHash(SystemSkillContent),
			InstalledAt: now,
			UpdatedAt:   now,
		}

		// Preserve original installedAt if metadata already exists
		if existingMeta, err := ReadSkillMeta(metaPath); err == nil {
			meta.InstalledAt = existingMeta.InstalledAt
		}

		if err := WriteSkillMeta(metaPath, meta); err != nil {
			Logger.Warn("failed to write system skill meta", "error", err)
		}
	}

	// Always sync to agents — symlinks may be missing even if content is current
	return SyncSkillToAgents(SystemSkillName, AgentIDs(DetectInstalledAgents()))
}
