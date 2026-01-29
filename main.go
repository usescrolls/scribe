package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/usescrolls/scribe/cli"
	"github.com/usescrolls/scribe/internal"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed frontend/dist
var assets embed.FS

var server *scribe.Server
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

	server = scribe.NewServer()

	if err := server.Initialize(); err != nil {
		scribe.Logger.Error("failed to initialize", "error", err)
		os.Exit(1)
	}

	if err := server.Load(); err != nil {
		scribe.Logger.Warn("failed to load registry", "error", err)
	}

	server.HandleURLScheme(urlArg)

	if err := server.GenerateMarketplace(); err != nil {
		scribe.Logger.Warn("failed to generate marketplace", "error", err)
	}
}

func runGUIMode() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	scribe.InitLogger(*debug)
	scribe.Logger.Info("initializing scribe", "version", scribe.Version, "debug", *debug)

	server = scribe.NewServer()

	if err := server.Initialize(); err != nil {
		scribe.Logger.Error("failed to initialize", "error", err)
		os.Exit(1)
	}

	if err := server.Migrate(); err != nil {
		scribe.Logger.Warn("migration failed", "error", err)
	}

	if err := server.Load(); err != nil {
		scribe.Logger.Warn("failed to load registry", "error", err)
	}

	if err := server.GenerateMarketplace(); err != nil {
		scribe.Logger.Warn("failed to generate marketplace", "error", err)
	}

	appService := NewAppService(server)

	wailsApp = application.New(application.Options{
		Name:        "Scribe",
		Description: "Plugin Manager for Claude Code",
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

	// Handle agenthub:// URLs when app is already running (macOS)
	wailsApp.Event.OnApplicationEvent(events.Common.ApplicationLaunchedWithUrl, func(event *application.ApplicationEvent) {
		url := event.Context().URL()
		scribe.Logger.Info("received URL via Wails event", "url", url)
		server.HandleURLScheme(url)
		if err := server.GenerateMarketplace(); err != nil {
			scribe.Logger.Warn("failed to generate marketplace", "error", err)
		}
		wailsApp.Event.Emit("plugins-updated")
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
	systray.SetIcon(scribe.GetIcon())

	// Create tray menu
	trayMenu := wailsApp.NewMenu()
	trayMenu.Add("Open Scribe").OnClick(func(ctx *application.Context) {
		mainWindow.Show()
	})
	trayMenu.AddSeparator()

	pluginCountItem := trayMenu.Add(fmt.Sprintf("%d plugins installed", server.PluginCount()))
	pluginCountItem.SetEnabled(false)

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

	// Update plugin count periodically
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				count := server.PluginCount()
				var label string
				if count == 1 {
					label = "1 plugin installed"
				} else {
					label = fmt.Sprintf("%d plugins installed", count)
				}
				pluginCountItem.SetLabel(label)
			}
		}
	}()

	// Run the application
	if err := wailsApp.Run(); err != nil {
		scribe.Logger.Error("application error", "error", err)
	}
}

// AppService provides bindings for the frontend
type AppService struct {
	server *scribe.Server
}

// NewAppService creates a new AppService
func NewAppService(s *scribe.Server) *AppService {
	return &AppService{server: s}
}

// PluginInfo for the frontend
type PluginInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	Category    string `json:"category,omitempty"`
	Author      string `json:"author,omitempty"`
	Source      string `json:"source"`
	SourceType  string `json:"sourceType"`
	InstalledAt string `json:"installedAt"`
}

// GetPlugins returns all installed plugins
func (a *AppService) GetPlugins() []PluginInfo {
	entries := a.server.GetAllPlugins()
	result := make([]PluginInfo, 0, len(entries))
	for _, e := range entries {
		author := ""
		if e.Author != nil {
			author = e.Author.Name
		}
		result = append(result, PluginInfo{
			Name:        e.Name,
			Description: e.Description,
			Version:     e.Version,
			Category:    e.Category,
			Author:      author,
			Source:      formatSourceForUI(e.Source),
			SourceType:  e.Source.Source,
			InstalledAt: e.InstalledAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result
}

// GetPluginCount returns the number of installed plugins
func (a *AppService) GetPluginCount() int {
	return a.server.PluginCount()
}

// UninstallPlugin removes a plugin by name
func (a *AppService) UninstallPlugin(name string) error {
	return a.server.UninstallPlugin(name)
}

// UninstallAllPlugins removes all plugins
func (a *AppService) UninstallAllPlugins() error {
	return a.server.UninstallAllPlugins()
}

// FullReset performs a complete reset
func (a *AppService) FullReset() error {
	return a.server.FullReset()
}

// GetVersion returns the application version
func (a *AppService) GetVersion() string {
	return scribe.Version
}

// ClaudeCodeDetected checks if Claude Code is installed
func (a *AppService) ClaudeCodeDetected() bool {
	return a.server.ClaudeCodeDetected()
}

// HandleURL processes an agenthub:// URL
func (a *AppService) HandleURL(urlStr string) {
	a.server.HandleURLScheme(urlStr)
	if wailsApp != nil {
		wailsApp.Event.Emit("plugins-updated", nil)
	}
}

func formatSourceForUI(source scribe.PluginSource) string {
	switch source.Source {
	case "github":
		return source.Repo
	case "npm":
		return source.Package
	case "url", "git", "zip":
		return source.URL
	default:
		return source.Source
	}
}
