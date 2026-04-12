package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run initial setup wizard",
	Long:  `Run the initial setup wizard to configure Scribe and detect your coding agents.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOnboarding()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

// checkOnboarding verifies onboarding is complete before running other commands
// Returns true if onboarding is needed
func checkOnboarding() bool {
	completed, err := scribe.IsOnboardingCompleted()
	if err != nil {
		return false
	}
	return !completed
}

// runOnboardingIfNeeded runs onboarding before executing a command if not completed
func runOnboardingIfNeeded() error {
	if !checkOnboarding() {
		return nil
	}

	fmt.Println("Welcome to Scribe!")
	fmt.Println("Before you can use Scribe, we need to complete initial setup.")
	fmt.Println()

	return runOnboarding()
}

// checkAndPromptTerms checks if terms have been accepted and prompts if not.
// This is a lightweight gate for existing users who completed onboarding before
// terms and conditions were added. Returns nil if terms are already accepted or
// if the user accepts them now.
func checkAndPromptTerms() error {
	accepted, err := scribe.AreTermsAccepted()
	if err != nil {
		scribe.Logger.Warn("failed to check terms acceptance", "error", err)
		return nil // don't block on errors
	}
	if accepted {
		return nil
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("Scribe has updated its Terms & Conditions.")
	fmt.Println("Please review and accept before continuing:")
	fmt.Println()
	printTermsClauses()
	fmt.Println()

	fmt.Print("Do you accept these terms? [y/N] ")
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input != "y" && input != "yes" {
		fmt.Println("You must accept the terms and conditions to use Scribe.")
		return fmt.Errorf("terms and conditions not accepted")
	}

	if err := scribe.AcceptTerms(); err != nil {
		return fmt.Errorf("failed to save terms acceptance: %w", err)
	}
	fmt.Println("Terms accepted.")
	fmt.Println()
	return nil
}

// printTermsClauses prints the terms clauses from the centralized source of truth.
func printTermsClauses() {
	for i, clause := range scribe.TermsClauses {
		fmt.Printf("  %d. %s: %s\n", i+1, clause.Title, clause.Body)
		fmt.Println()
	}
}

// runOnboarding runs the interactive CLI onboarding flow
func runOnboarding() error {
	reader := bufio.NewReader(os.Stdin)

	// Step 1: Welcome
	fmt.Println("Welcome to Scribe!")
	fmt.Println("Scribe syncs AI coding skills to all your coding agents.")
	fmt.Println()

	// Step 2: Terms and Conditions
	fmt.Println("Before continuing, please review the following terms:")
	fmt.Println()
	fmt.Println("  Terms & Conditions")
	fmt.Println("  ==================")
	fmt.Println()
	printTermsClauses()
	fmt.Println()

	fmt.Print("Do you accept these terms? [y/N] ")
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input != "y" && input != "yes" {
		fmt.Println("You must accept the terms and conditions to use Scribe.")
		return fmt.Errorf("terms and conditions not accepted")
	}

	if err := scribe.AcceptTerms(); err != nil {
		return fmt.Errorf("failed to save terms acceptance: %w", err)
	}
	fmt.Println("Terms accepted.")
	fmt.Println()

	// Step 3: Detect agents
	fmt.Println("Detecting installed coding agents...")
	agents := scribe.DetectInstalledAgents()

	if len(agents) == 0 {
		fmt.Println()
		fmt.Println("No coding agents detected.")
		fmt.Println("Please install Claude Code, Cursor, or another supported agent first.")
		fmt.Println()
		fmt.Println("Supported agents include:")
		fmt.Println("  - Claude Code (https://claude.ai/claude-code)")
		fmt.Println("  - Cursor (https://cursor.sh)")
		fmt.Println("  - GitHub Copilot")
		fmt.Println("  - And 40+ more...")
		return fmt.Errorf("no coding agents detected")
	}

	fmt.Printf("Found %d coding agent(s):\n", len(agents))
	for _, agent := range agents {
		fmt.Printf("  * %s\n", agent.DisplayName)
	}
	fmt.Println()

	// Step 4: Check for existing skills
	fmt.Println("Checking for existing skills...")
	existingSkills, err := scribe.DetectExistingSkills()
	if err != nil {
		scribe.Logger.Warn("failed to detect existing skills", "error", err)
	}

	if len(existingSkills) > 0 {
		fmt.Printf("Found %d existing skill(s) in agent directories:\n", len(existingSkills))
		for _, skill := range existingSkills {
			gitMark := ""
			if skill.IsGitRepo {
				gitMark = " [git repo]"
			}
			fmt.Printf("  * %s (in %s)%s\n", skill.Name, skill.AgentName, gitMark)
		}
		fmt.Println()

		fmt.Print("Would you like to import these skills into Scribe? [Y/n/d(elete)] ")
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "", "y", "yes":
			fmt.Println("Importing skills...")
			if err := scribe.ImportExistingSkills(existingSkills); err != nil {
				return fmt.Errorf("failed to import skills: %w", err)
			}
			fmt.Println("Skills imported successfully!")
		case "d", "delete":
			fmt.Println("Deleting existing skills...")
			if err := scribe.DeleteExistingSkills(existingSkills); err != nil {
				return fmt.Errorf("failed to delete skills: %w", err)
			}
			fmt.Println("Skills deleted.")
		default:
			fmt.Println("Skipping import.")
		}
		fmt.Println()
	}

	// Step 5: Install demo skill
	fmt.Print("Would you like to install the demo skill? [Y/n] ")
	input, _ = reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "" || input == "y" || input == "yes" {
		fmt.Println("Installing demo skill...")
		if err := scribe.InstallDemoSkill(); err != nil {
			return fmt.Errorf("failed to install demo skill: %w", err)
		}

		scrollsDir, _ := scribe.GetScrollsDir()
		fmt.Printf("Installed scribe-welcome to %s/scribe-welcome\n", scrollsDir)
		fmt.Printf("Synced to %d agent(s)\n", len(agents))
	} else {
		fmt.Println("Skipping demo skill installation.")
	}
	fmt.Println()

	// Step 6: Complete
	if err := scribe.CompleteOnboarding(); err != nil {
		return fmt.Errorf("failed to complete onboarding: %w", err)
	}

	fmt.Println("Setup complete! Run `scribe list` to see your skills.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  - Visit AgentHub to discover more skills")
	fmt.Println("  - Use `scribe install <github-repo>` to install skills")
	fmt.Println("  - Create your own skills with SKILL.md files")

	return nil
}
