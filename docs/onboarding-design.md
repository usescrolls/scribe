# Onboarding Flow Design Document

> **Status: Fully Implemented.** This design has been implemented in both the GUI (`frontend/src/components/OnboardingWizard.vue` and `frontend/src/components/onboarding/`) and CLI (`cli/onboarding.go`), with backend logic in `internal/onboarding.go`. The `scribe setup` command runs the CLI wizard. The GUI wizard runs automatically on first launch.

## Overview

Scribe has a first-run onboarding experience that helps new users understand what Scribe does, detects their coding agents, and optionally imports existing skills into Scribe's management system.

## Key Decisions

### 1. Presentation Style
- **Full-screen wizard** (macOS Setup Assistant style)
- **Mandatory**: Users cannot access the main app until onboarding is complete
- Multi-step flow with clear progression

### 2. First Skill Installation
- Provide a **mini placeholder skill** as a demo
- Purpose: Show users how skills work without requiring external dependencies

### 3. Agent Detection Display
- Show **all agents** (both detected and non-detected)
- Detected agents: Highlighted/enabled
- Non-detected agents: Shown but grayed out (no install links)
- **Always sync to all detected agents** - no per-agent management

### 4. Existing Skills Handling
- **Import approach**: Move existing skills to `~/.scribe/scrolls/`
- Detect skills in agent directories (e.g., `~/.claude/skills/`)
- Handle both:
  - Individual skill files/folders
  - Git repositories containing skills (preserve git tracking)
- **Symlink direction is NOT negotiable**: Always `~/.scribe/scrolls/skill → ~/.agent/skills/skill`

### 5. State Persistence
- Track completion in `~/.scribe/config.json`:
  ```json
  {
    "activeWorkspace": "default",
    "onboardingCompleted": true
  }
  ```

### 6. Edge Cases
- **No agents detected**: Block onboarding with helpful message until user installs at least one coding agent
- **CLI onboarding**: Yes, CLI should have its own onboarding flow (separate from GUI)

---

## GUI Onboarding Flow

### UI Design Details
- **Progress indicators**: Dots showing current step (not numbered)
- **Transitions**: Fade into content between steps (if simple to implement, otherwise instant)
- **Layout**: Full-screen, centered content, clean and minimal

### Screen 1: Welcome
- Scribe logo/branding
- Brief tagline explaining what Scribe does
- "Get Started" button

### Screen 2: Agent Detection
- Scan for installed coding agents
- Display results:
  - Detected agents (with checkmark, highlighted)
  - Non-detected agents (grayed out, for awareness)
- Show count: "Found X coding agents on your system"
- **Blocking state**: If no agents detected, show message and prevent progression
  - Auto-rescan every 30 seconds (no manual rescan button needed)
  - Message: "We couldn't find any coding agents installed. Please install Claude Code, Cursor, or another supported agent."

### Screen 3: Existing Skills Detection
- Scan detected agent directories for existing skills
- If skills found:
  - List discovered skills with their locations
  - Explain import process
  - **Simple choice**: "Import all" OR "Delete all and start fresh"
  - Handle git repos: Preserve git tracking when moving
- If no existing skills:
  - Skip this screen entirely (proceed to Screen 4)

### Screen 4: Install First Skill
- Offer to install the placeholder/demo skill
- Explain what will happen (skill copied to scrolls, symlinked to all agents)
- Show the skill being distributed to detected agents (visual feedback)

### Screen 5: Complete
- Success message
- Quick tips or next steps
- "Open Scribe" button to enter main app

---

## CLI Onboarding Flow

When user runs `scribe` CLI for the first time (no `onboardingCompleted` flag):

```
Welcome to Scribe!
Scribe syncs AI coding skills to all your coding agents.

Detecting installed coding agents...
✓ Found 3 coding agents:
  • Claude Code (~/.claude)
  • Cursor (~/.cursor)
  • GitHub Copilot (~/.github-copilot)

Checking for existing skills...
Found 2 skills in agent directories:
  • react-patterns (in ~/.claude/skills/) [git repo]
  • typescript-tips (in ~/.cursor/skills/)

Would you like to import these skills into Scribe? [Y/n]
> y

Importing skills...
✓ Moved react-patterns to ~/.scribe/scrolls/react-patterns
✓ Moved typescript-tips to ~/.scribe/scrolls/typescript-tips
✓ Synced to all detected agents

Would you like to install a demo skill to see how Scribe works? [Y/n]
> y

Installing demo skill...
✓ Installed scribe-welcome to ~/.scribe/scrolls/scribe-welcome
✓ Synced to 3 agents

Setup complete! Run `scribe list` to see your skills.
```

### CLI Edge Cases
- **No agents detected**: Show error message, exit with helpful guidance
- **Plain text output only** (no colors) for maximum compatibility
- **First-time users**: Force onboarding before allowing any other commands
  - If user runs `scribe install owner/repo` before completing onboarding, show interactive onboarding prompts first
  - After onboarding completes, proceed with the original command

---

## Existing Skills Import Logic

### Detection
1. For each detected agent, check their skills directory
2. Identify skill files/folders (look for SKILL.md or valid skill structure)
3. Detect if skill is a git repository (check for `.git` directory)

### Import Process
1. **Git repos**:
   - Move entire repo to `~/.scribe/scrolls/`
   - Preserve `.git` directory and all git tracking
   - Create `.scribe-meta.json` with source info
2. **Non-git skills**:
   - Move to `~/.scribe/scrolls/`
   - Create `.scribe-meta.json` marking as "local" source
3. **After import**:
   - Remove original from agent directory
   - Create symlink from `~/.scribe/scrolls/skill` → agent's skills directory
   - Sync to all other detected agents

### Conflict Handling
- If skill with same name exists in multiple agent directories:
  - Compare content hashes
  - If identical: Import once, remove duplicates
  - If different: **Ask user to choose** which version to keep (show source locations)

---

## Placeholder/Demo Skill

**Name**: `scribe-welcome`

**Purpose**: Purely demonstrative - explains Scribe, not actual coding tips

**Content** (SKILL.md):
```markdown
---
name: scribe-welcome
description: A welcome skill that introduces Scribe and demonstrates skill formatting
---

# Welcome to Scribe!

This is a demo skill installed during Scribe setup. It demonstrates how skills work.

## What is Scribe?

Scribe is a skill distribution tool that syncs AI coding skills to all your coding agents.
Install a skill once, and it's automatically available in Claude Code, Cursor, GitHub Copilot,
and 40+ other agents.

## How Skills Work

Skills are markdown files with YAML frontmatter that provide context and instructions to
AI coding agents. This skill is now available in all your detected agents.

## Next Steps

1. Visit AgentHub to discover more skills
2. Use `scribe install <github-repo>` to install skills from GitHub
3. Create your own skills by adding SKILL.md files

You can safely uninstall this demo skill anytime with `scribe uninstall scribe-welcome`.
```

---

## Technical Implementation Notes

### Frontend (Vue)

1. Create `OnboardingWizard.vue` component
2. Conditionally render in `App.vue` based on `onboardingCompleted` config
3. New composable: `useOnboarding.ts` for state management
4. New Wails bindings needed:
   - `CheckOnboardingCompleted() bool`
   - `CompleteOnboarding() error`
   - `DetectExistingSkills() []ExistingSkillInfo`
   - `ImportExistingSkills(skillNames []string) error`
   - `InstallDemoSkill() error`
   - `RescanAgents() []AgentStatus`

### Backend (Go)

1. Add `OnboardingCompleted` field to config struct
2. Implement existing skill detection in `internal/`
3. Implement import logic with git repo handling
4. Bundle demo skill content (embed in binary or generate on-the-fly)

### CLI

1. Add onboarding check to CLI entry point
2. Interactive prompts using standard input
3. Share import logic with GUI backend

---

## File Changes Required

### New Files
- `frontend/src/components/OnboardingWizard.vue`
- `frontend/src/components/onboarding/WelcomeStep.vue`
- `frontend/src/components/onboarding/AgentDetectionStep.vue`
- `frontend/src/components/onboarding/ExistingSkillsStep.vue`
- `frontend/src/components/onboarding/InstallDemoStep.vue`
- `frontend/src/components/onboarding/CompleteStep.vue`
- `frontend/src/composables/useOnboarding.ts`
- `internal/onboarding.go`
- `cli/onboarding.go`

### Modified Files
- `frontend/src/App.vue` - Add onboarding gate
- `internal/types.go` - Add config fields
- `internal/config.go` - Add onboarding methods
- `main.go` - Add onboarding bindings, CLI check
- `cli/root.go` - Add onboarding check before commands
