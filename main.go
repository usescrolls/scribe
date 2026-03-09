package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"slices"
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

//go:embed icons/tray-icon.png
var trayIcon []byte

const (
	labelShowScribe = "Show Scribe"
	labelHideScribe = "Hide Scribe"
)

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
	if arg == "--help" || arg == "-h" {
		return true
	}
	if strings.HasPrefix(arg, "-") {
		return false
	}
	commands := cli.CLICommands()
	return slices.Contains(commands, arg)
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
		scribe.CloseLogger()
		os.Exit(1)
	}

	// Ensure system skill is installed and up to date
	if err := scribe.EnsureSystemSkill(); err != nil {
		scribe.Logger.Warn("failed to ensure system skill", "error", err)
	}

	// Ensure current workspace skills are synced to agents
	if err := scribe.ResyncCurrentWorkspace(); err != nil {
		scribe.Logger.Warn("failed to resync workspace on startup", "error", err)
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

		// Handle "show" action: just bring the window to front
		if strings.Contains(urlString, "agenthub://show") {
			if mainWindow != nil {
				mainWindow.Show()
				mainWindow.Focus()
			}
			return
		}

		// Run installation in background to not block the event handler
		go func() {
			result := scribe.HandleInstallURL(urlString)

			if result.Success {
				scribe.Logger.Info("URL scheme installation complete",
					"skills_installed", result.SkillsCount,
					"skill_names", result.SkillNames)

				// Emit events to update frontend
				wailsApp.Event.Emit("skills-updated", nil)
				wailsApp.Event.Emit("workspace-changed", nil)

				// Show the window to confirm installation
				if mainWindow != nil {
					mainWindow.Show()
					mainWindow.Focus()
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
	systray.SetTemplateIcon(trayIcon)

	// Create tray menu
	trayMenu := wailsApp.NewMenu()
	toggleItem := trayMenu.Add(labelShowScribe)
	toggleItem.OnClick(func(ctx *application.Context) {
		if mainWindow.IsVisible() {
			mainWindow.Hide()
			toggleItem.SetLabel(labelShowScribe)
		} else {
			mainWindow.Show()
			mainWindow.Focus()
			toggleItem.SetLabel(labelHideScribe)
		}
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

	// Update tray menu labels periodically
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if mainWindow.IsVisible() {
				toggleItem.SetLabel(labelHideScribe)
			} else {
				toggleItem.SetLabel(labelShowScribe)
			}
			skillCountItem.SetLabel(getSkillCountLabel())
			workspaceItem.SetLabel(getWorkspaceLabel())
		}
	}()

	// Run the application
	if err := wailsApp.Run(); err != nil {
		scribe.Logger.Error("application error", "error", err)
	}
	scribe.CloseLogger()
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
