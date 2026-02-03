package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/usescrolls/scribe/cli"
	scribe "github.com/usescrolls/scribe/internal"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed icons/icon.png
var appIcon []byte

var wailsApp *application.App
var mainWindow *application.WebviewWindow

func main() {
	// Detection order:
	// 1. agenthub:// URL → URL scheme handler
	// 2. Known CLI command → CLI mode
	// 3. No arguments or flags → GUI mode

	if len(os.Args) > 1 {
		firstArg := os.Args[1]

		// Check for URL scheme
		if strings.HasPrefix(firstArg, "agenthub://") {
			handleURLScheme(firstArg)
			return
		}

		// Check for CLI commands
		if isCLICommand(firstArg) {
			exitCode := cli.Execute()
			os.Exit(exitCode)
		}
	}

	// GUI mode
	runGUIMode()
}

func isCLICommand(arg string) bool {
	if strings.HasPrefix(arg, "-") {
		return false
	}
	commands := cli.CLICommands()
	for _, cmd := range commands {
		if arg == cmd {
			return true
		}
	}
	return false
}

func handleURLScheme(urlArg string) {
	scribe.InitLogger(false)
	scribe.Logger.Info("URL scheme argument detected", "url", urlArg)

	// Handle the installation
	result := scribe.HandleInstallURL(urlArg)

	if result.Success {
		scribe.Logger.Info("URL scheme installation complete",
			"skills_installed", result.SkillsCount,
			"skill_names", result.SkillNames)
		fmt.Printf("Successfully installed %d skill(s): %v\n", result.SkillsCount, result.SkillNames)
	} else {
		scribe.Logger.Error("URL scheme installation failed", "error", result.ErrorMessage)
		fmt.Fprintf(os.Stderr, "Installation failed: %s\n", result.ErrorMessage)
		os.Exit(1)
	}
}

func runGUIMode() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	scribe.InitLogger(*debug)
	scribe.Logger.Info("initializing scribe", "version", scribe.Version, "debug", *debug)

	// Initialize skills system directories
	if err := scribe.EnsureScribeDirs(); err != nil {
		scribe.Logger.Error("failed to initialize scribe directories", "error", err)
		os.Exit(1)
	}

	appService := NewAppService()

	wailsApp = application.New(application.Options{
		Name:        "Scribe",
		Description: "Skills Manager for Coding Agents",
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// Set application icon (used for dock, taskbar, and app switcher)
	wailsApp.SetIcon(appIcon)

	// Handle agenthub:// URLs when app is already running (macOS)
	wailsApp.Event.OnApplicationEvent(events.Common.ApplicationLaunchedWithUrl, func(event *application.ApplicationEvent) {
		urlString := event.Context().URL()
		scribe.Logger.Info("received URL via Wails event", "url", urlString)

		// Run installation in background to not block the event handler
		go func() {
			result := scribe.HandleInstallURL(urlString)

			if result.Success {
				scribe.Logger.Info("URL scheme installation complete",
					"skills_installed", result.SkillsCount,
					"skill_names", result.SkillNames)

				// Emit events to update frontend
				wailsApp.Event.Emit("skills-updated", nil)

				// Show the window to confirm installation
				if mainWindow != nil {
					mainWindow.Show()
				}
			} else {
				scribe.Logger.Error("URL scheme installation failed", "error", result.ErrorMessage)
			}
		}()
	})

	// Create main window (hidden initially)
	mainWindow = wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Scribe",
		Width:     800,
		Height:    600,
		MinWidth:  400,
		MinHeight: 300,
		Hidden:    true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
	})

	// Hide instead of close when X is clicked, so we can reopen from tray
	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		mainWindow.Hide()
		e.Cancel()
	})

	// Create system tray
	systray := wailsApp.SystemTray.New()
	systray.SetIcon(appIcon)

	// Create tray menu
	trayMenu := wailsApp.NewMenu()
	trayMenu.Add("Open Scribe").OnClick(func(ctx *application.Context) {
		mainWindow.Show()
	})
	trayMenu.AddSeparator()

	skillCountItem := trayMenu.Add(getSkillCountLabel())
	skillCountItem.SetEnabled(false)

	workspaceItem := trayMenu.Add(getWorkspaceLabel())
	workspaceItem.SetEnabled(false)

	trayMenu.AddSeparator()
	trayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		wailsApp.Quit()
	})

	systray.SetMenu(trayMenu)

	// Click on tray icon toggles window
	systray.OnClick(func() {
		if mainWindow.IsVisible() {
			mainWindow.Hide()
		} else {
			mainWindow.Show()
			// Note: Focus() removed due to Wails v3 alpha crash on macOS 26
		}
	})

	// Update skill count and workspace periodically
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				skillCountItem.SetLabel(getSkillCountLabel())
				workspaceItem.SetLabel(getWorkspaceLabel())
			}
		}
	}()

	// Run the application
	if err := wailsApp.Run(); err != nil {
		scribe.Logger.Error("application error", "error", err)
	}
}

// AppService provides bindings for the frontend
type AppService struct{}

// NewAppService creates a new AppService
func NewAppService() *AppService {
	return &AppService{}
}

// GetVersion returns the application version
func (a *AppService) GetVersion() string {
	return scribe.Version
}

// ======================================================================
// Skills API
// ======================================================================

// GetSkills returns all installed skills
func (a *AppService) GetSkills() ([]scribe.SkillInfo, error) {
	return scribe.GetAllSkillInfo()
}

// GetSkillCount returns the number of installed skills
func (a *AppService) GetSkillCount() int {
	skills, err := scribe.ReadAllSkills()
	if err != nil {
		return 0
	}
	return len(skills)
}

// RemoveSkill removes a skill by name from all agents and workspaces
func (a *AppService) RemoveSkill(name string) error {
	scribe.Logger.Info("AppService.RemoveSkill called", "name", name)

	// Remove from all workspaces first
	if err := scribe.RemoveSkillFromAllWorkspaces(name); err != nil {
		scribe.Logger.Error("failed to remove skill from workspaces", "name", name, "error", err)
		return err
	}

	// UninstallSkill handles removing from agents and canonical location
	if err := scribe.UninstallSkill(name); err != nil {
		scribe.Logger.Error("failed to uninstall skill", "name", name, "error", err)
		return err
	}

	scribe.Logger.Info("AppService.RemoveSkill succeeded", "name", name)

	// Emit event to update frontend
	if wailsApp != nil {
		wailsApp.Event.Emit("skills-updated", nil)
	}

	return nil
}

// ======================================================================
// Workspaces API
// ======================================================================

// GetWorkspaces returns all workspaces with their active status
func (a *AppService) GetWorkspaces() ([]scribe.WorkspaceInfo, error) {
	return scribe.GetWorkspaceInfo()
}

// GetActiveWorkspaceName returns the name of the active workspace
func (a *AppService) GetActiveWorkspaceName() (string, error) {
	ws, err := scribe.GetActiveWorkspace()
	if err != nil {
		return "", err
	}
	return ws.Name, nil
}

// SetActiveWorkspace switches to a different workspace
func (a *AppService) SetActiveWorkspace(name string) error {
	scribe.Logger.Info("AppService.SetActiveWorkspace called", "name", name)

	current, err := scribe.GetActiveWorkspace()
	if err != nil {
		return err
	}

	target, err := scribe.GetWorkspace(name)
	if err != nil {
		return err
	}

	// Sync symlinks to match target workspace
	if err := scribe.SyncWorkspace(current, target); err != nil {
		return err
	}

	// Update active workspace in config
	if err := scribe.SetActiveWorkspace(name); err != nil {
		return err
	}

	scribe.Logger.Info("AppService.SetActiveWorkspace succeeded", "name", name)

	// Emit event to update frontend
	if wailsApp != nil {
		wailsApp.Event.Emit("workspace-changed", name)
		wailsApp.Event.Emit("skills-updated", nil)
	}

	return nil
}

// CreateWorkspace creates a new workspace
func (a *AppService) CreateWorkspace(name, description string) error {
	ws := &scribe.Workspace{
		Name:        name,
		Description: description,
		Skills:      []string{},
	}
	return scribe.CreateWorkspace(ws)
}

// DeleteWorkspace removes a workspace
func (a *AppService) DeleteWorkspace(name string) error {
	return scribe.DeleteWorkspace(name)
}

// AddSkillToWorkspace adds a skill to a specific workspace
func (a *AppService) AddSkillToWorkspace(skillName, workspaceName string) error {
	err := scribe.AddSkillToWorkspace(skillName, workspaceName)
	if err == nil && wailsApp != nil {
		wailsApp.Event.Emit("workspace-changed", workspaceName)
	}
	return err
}

// RemoveSkillFromWorkspace removes a skill from a specific workspace
func (a *AppService) RemoveSkillFromWorkspace(skillName, workspaceName string) error {
	err := scribe.RemoveSkillFromWorkspace(skillName, workspaceName)
	if err == nil && wailsApp != nil {
		wailsApp.Event.Emit("workspace-changed", workspaceName)
		wailsApp.Event.Emit("skills-updated", nil)
	}
	return err
}

// ======================================================================
// Agents API
// ======================================================================

// GetAgentStatus returns the status of all supported agents
func (a *AppService) GetAgentStatus() []scribe.AgentStatus {
	scrollsDir, err := scribe.GetScrollsDir()
	if err != nil {
		scribe.Logger.Error("failed to get scrolls dir", "error", err)
		return []scribe.AgentStatus{}
	}
	return scribe.GetAgentStatus(scrollsDir)
}

// GetInstalledAgentCount returns how many agents are installed
func (a *AppService) GetInstalledAgentCount() int {
	agents := scribe.DetectInstalledAgents()
	return len(agents)
}

// GetTotalAgentCount returns the total number of supported agents
func (a *AppService) GetTotalAgentCount() int {
	return len(scribe.GetAllAgents())
}

// Helper functions for system tray labels

func getSkillCountLabel() string {
	skills, err := scribe.ReadAllSkills()
	if err != nil {
		return "0 skills installed"
	}
	count := len(skills)
	if count == 1 {
		return "1 skill installed"
	}
	return fmt.Sprintf("%d skills installed", count)
}

func getWorkspaceLabel() string {
	ws, err := scribe.GetActiveWorkspace()
	if err != nil {
		return "Workspace: default"
	}
	return fmt.Sprintf("Workspace: %s", ws.Name)
}
