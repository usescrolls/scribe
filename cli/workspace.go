package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

var (
	workspaceCmd = &cobra.Command{
		Use:   "workspace",
		Short: "Manage skill workspaces",
		Long: `Manage workspaces for organizing skills.

Workspaces let you switch between different sets of skills globally.
When you switch workspaces, symlinks are updated across all detected agents.

Examples:
  scribe workspace list
  scribe workspace create web-dev
  scribe workspace use backend
  scribe workspace add react-best-practices
  scribe workspace remove typescript-patterns
  scribe workspace current`,
	}

	workspaceListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all workspaces",
		RunE:  runWorkspaceList,
	}

	workspaceCreateCmd = &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new workspace",
		Args:  cobra.ExactArgs(1),
		RunE:  runWorkspaceCreate,
	}

	workspaceUseCmd = &cobra.Command{
		Use:   "use <name>",
		Short: "Switch to a workspace",
		Args:  cobra.ExactArgs(1),
		RunE:  runWorkspaceUse,
	}

	workspaceAddCmd = &cobra.Command{
		Use:   "add <skill-name>",
		Short: "Add a skill to the current workspace",
		Args:  cobra.ExactArgs(1),
		RunE:  runWorkspaceAdd,
	}

	workspaceRemoveCmd = &cobra.Command{
		Use:   "remove <skill-name>",
		Short: "Remove a skill from the current workspace",
		Args:  cobra.ExactArgs(1),
		RunE:  runWorkspaceRemove,
	}

	workspaceCurrentCmd = &cobra.Command{
		Use:   "current",
		Short: "Show the current workspace",
		RunE:  runWorkspaceCurrent,
	}

	workspaceDeleteCmd = &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a workspace",
		Args:  cobra.ExactArgs(1),
		RunE:  runWorkspaceDelete,
	}

	// Flags
	workspaceDescription string
)

func init() {
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceCreateCmd)
	workspaceCmd.AddCommand(workspaceUseCmd)
	workspaceCmd.AddCommand(workspaceAddCmd)
	workspaceCmd.AddCommand(workspaceRemoveCmd)
	workspaceCmd.AddCommand(workspaceCurrentCmd)
	workspaceCmd.AddCommand(workspaceDeleteCmd)

	workspaceCreateCmd.Flags().StringVarP(&workspaceDescription, "description", "d", "", "Workspace description")
}

func runWorkspaceList(cmd *cobra.Command, args []string) error {
	workspaces, err := scribe.ListWorkspaces()
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	config, err := scribe.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if jsonOutput {
		infos := make([]scribe.WorkspaceInfo, len(workspaces))
		for i, ws := range workspaces {
			infos[i] = scribe.WorkspaceInfo{
				Name:        ws.Name,
				Description: ws.Description,
				Skills:      ws.Skills,
				IsActive:    ws.Name == config.ActiveWorkspace,
			}
		}
		data, _ := json.MarshalIndent(infos, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Workspaces (%d):\n", len(workspaces))
	for _, ws := range workspaces {
		marker := "  "
		if ws.Name == config.ActiveWorkspace {
			marker = "* "
		}
		fmt.Printf("%s%s (%d skills)", marker, ws.Name, len(ws.Skills))
		if ws.Description != "" {
			fmt.Printf(" - %s", ws.Description)
		}
		fmt.Println()
	}

	return nil
}

func runWorkspaceCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	ws := &scribe.Workspace{
		Name:        name,
		Description: workspaceDescription,
		Skills:      []string{},
	}

	if err := scribe.CreateWorkspace(ws); err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	if !quiet {
		fmt.Printf("✓ Workspace '%s' created\n", name)
	}

	return nil
}

func runWorkspaceUse(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Get current workspace for comparison
	config, err := scribe.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if config.ActiveWorkspace == name {
		if !quiet {
			fmt.Printf("Already using workspace '%s'\n", name)
		}
		return nil
	}

	if !quiet {
		fmt.Printf("Switching from '%s' to '%s'...\n", config.ActiveWorkspace, name)
	}

	if err := scribe.SetActiveWorkspace(name); err != nil {
		return fmt.Errorf("failed to switch workspace: %w", err)
	}

	// Show what changed
	ws, _ := scribe.GetWorkspace(name)
	if ws != nil && !quiet {
		fmt.Printf("✓ Active workspace: %s (%d skills)\n", name, len(ws.Skills))
	}

	return nil
}

func runWorkspaceAdd(cmd *cobra.Command, args []string) error {
	skillName := args[0]

	// Get current workspace
	config, err := scribe.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if skill exists
	exists, err := scribe.SkillExists(skillName)
	if err != nil {
		return fmt.Errorf("failed to check skill: %w", err)
	}
	if !exists {
		return fmt.Errorf("skill '%s' not found. Install it first with 'scribe add'", skillName)
	}

	if err := scribe.AddSkillToWorkspace(skillName, config.ActiveWorkspace); err != nil {
		return fmt.Errorf("failed to add skill to workspace: %w", err)
	}

	if !quiet {
		fmt.Printf("✓ Added '%s' to workspace '%s'\n", skillName, config.ActiveWorkspace)
	}

	return nil
}

func runWorkspaceRemove(cmd *cobra.Command, args []string) error {
	skillName := args[0]

	if scribe.IsSystemSkill(skillName) {
		return fmt.Errorf("cannot remove system skill '%s' from workspace", skillName)
	}

	// Get current workspace
	config, err := scribe.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := scribe.RemoveSkillFromWorkspace(skillName, config.ActiveWorkspace); err != nil {
		return fmt.Errorf("failed to remove skill from workspace: %w", err)
	}

	if !quiet {
		fmt.Printf("✓ Removed '%s' from workspace '%s'\n", skillName, config.ActiveWorkspace)
	}

	return nil
}

func runWorkspaceCurrent(cmd *cobra.Command, args []string) error {
	ws, err := scribe.GetActiveWorkspace()
	if err != nil {
		return fmt.Errorf("failed to get active workspace: %w", err)
	}

	if jsonOutput {
		config, _ := scribe.LoadConfig()
		info := scribe.WorkspaceInfo{
			Name:        ws.Name,
			Description: ws.Description,
			Skills:      ws.Skills,
			IsActive:    true,
		}
		_ = config // suppress unused warning
		data, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Active workspace: %s\n", ws.Name)
	if ws.Description != "" {
		fmt.Printf("Description: %s\n", ws.Description)
	}
	fmt.Printf("Skills (%d):\n", len(ws.Skills))
	for _, skill := range ws.Skills {
		fmt.Printf("  • %s\n", skill)
	}

	return nil
}

func runWorkspaceDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := scribe.DeleteWorkspace(name); err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	if !quiet {
		fmt.Printf("✓ Workspace '%s' deleted\n", name)
	}

	return nil
}
