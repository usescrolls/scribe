# Upcoming Features

This document tracks planned features and improvements for Scribe.

## Source Types

### Well-Known Endpoint
**Priority:** Low

Support for `/.well-known/skills/` discovery endpoint:
```bash
scribe install https://example.com
# Fetches https://example.com/.well-known/skills/ to discover available skills
```


### Cross-Platform Testing
Verify functionality across platforms:
- macOS: Symlink behavior
- Linux: Symlink permissions
- Windows: Junction creation and permissions

## Documentation

### Marketing Comparison
Add intro comparing Scribe with similar tools:
- Comparison with Vercel Skills CLI
- Unique value proposition (multi-agent support, workspaces)
- When to use Scribe vs alternatives


manage updates inside scribe itself.


i want to be able to browse skills marketplace inside @scribe/ . we should be able to switch between vercel marketplace (https://skills.sh) or github search (is there a free api for that?). 

---
inconsistencies on installation on windows and linux: 
Documentation vs Reality mismatches
1. installation.md advertises binaries that don't exist in CI

installation.md:19-26 lists downloads for scribe-darwin-arm64 and scribe-darwin-amd64, but the release workflow only builds darwin-arm64. There is no Intel Mac build. Same for Linux — only amd64, no arm64.

2. installation.md doesn't mention native deps for Linux

installation.md:28-31 tells Linux users to just curl the binary and run it. No mention that it needs libgtk-3, libwebkit2gtk-4.1, and libayatana-appindicator3 at runtime. This will fail silently with a dynamic linker error.

3. install.ps1 is referenced but never published to releases

installation.md:44 tells users to download install.ps1 from usescrolls.com/releases, but the release workflow never uploads it — only the raw .exe is uploaded. The install script only exists in the repo's packaging/ directory.

4. development.md acknowledges the .app bundle requirement that installation.md ignores

development.md:311 says "Must run as .app bundle for URL scheme to work (not raw binary)" — but installation.md Option 3 tells macOS users to download the raw binary and run it directly, which means no URL scheme, no dock icon, no proper app behavior.

5. Linux install.sh is also never published

The existing install.sh handles icons, .desktop file, xdg-mime registration — but it's never uploaded to releases either.

What's actually missing from distribution
Gap	Impact
No Linux native deps handling	Binary won't run at all
No install.ps1 / install.sh in releases	Install scripts exist but aren't distributed
No macOS Intel (amd64) build	Intel Mac users can't install
No Linux arm64 build	Raspberry Pi / ARM server users can't install
No Windows installer (.exe/.msi)	No URL scheme, no PATH, no Start Menu
No Linux packaging (AppImage/.deb)	No dependency resolution
No package managers (Scoop/WinGet/AUR)	No auto-updates, manual download only
upcoming-features.md doesn't track any of this	No roadmap for distribution
The core problem
The macOS story is complete: DMG + Homebrew + auto-update via tap. Windows and Linux are effectively "here's a raw binary on a GitHub Release page" with install scripts that exist in the repo but aren't distributed, and no mention of the native dependencies the binary requires to even launch.

Want me to plan out fixing this — starting with the most impactful items (publishing the install scripts, adding dep handling, adding the missing arch builds)?


---

Launch at Startup — Research Findings
Best option: Wails v3 built-in start_at_login plugin
Since Scribe is a Wails v3 app, the simplest path is Wails' own plugin. You already depend on github.com/wailsapp/wails/v3 v3.0.0-alpha.64, and the plugin lives in the same module — no new dependency needed.

Import path:


import "github.com/wailsapp/wails/v3/plugins/start_at_login"
What it does per platform:

Platform	Mechanism	Details
macOS	AppleScript → Login Items	Registers the app as a Login Item. Requires NSAppleEventsUsageDescription in Info.plist
Linux	XDG Autostart .desktop file	Creates ~/.config/autostart/<app>.desktop
Windows	Registry Run key	Writes to HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run
Integration into Scribe
1. Register the plugin in main.go runGUIMode():


import "github.com/wailsapp/wails/v3/plugins/start_at_login"

// In application.New Options:
Plugins: map[string]application.Plugin{
    "start_at_login": start_at_login.NewPlugin(start_at_login.Config{}),
},
2. API available from Go:


start_at_login.StartAtLogin(true)        // enable
start_at_login.StartAtLogin(false)       // disable
enabled, err := start_at_login.IsStartAtLogin()  // check
3. API available from frontend JS (automatically exposed):


// Enable
wails.Plugin("start_at_login", "StartAtLogin", true)
// Check
wails.Plugin("start_at_login", "IsStartAtLogin")
4. macOS Info.plist addition needed (packaging/macos/Info.plist):


<key>NSAppleEventsUsageDescription</key>
<string>Scribe needs permission to manage login items.</string>
Homebrew considerations
Homebrew casks don't have a built-in way to register login items during install. The standard pattern (used by Rectangle, Stats, HiddenBar, etc.) is:

The app itself manages auto-start via an in-app toggle (exactly what the Wails plugin provides)
The cask only handles cleanup on uninstall:

uninstall login_item: "Scribe"
Alternatively, if Scribe is distributed as a Homebrew formula (not cask), you can add a service do block for brew services start scribe support — but this is separate from the in-app toggle and more suited for headless/CLI usage.

Alternative: Native implementation (if the plugin doesn't work)
If the Wails plugin has issues with alpha.64, the fallback is a custom internal/autostart package using:

macOS: LaunchAgent plist in ~/Library/LaunchAgents/com.usescrolls.scribe.plist with RunAtLoad and KeepAlive
Linux: XDG .desktop file in ~/.config/autostart/
Windows: Registry key via golang.org/x/sys/windows/registry (pure Go, no CGO)
This is more code but gives full control over things like KeepAlive (restart on crash) and log paths.

Recommendation
Start with the Wails start_at_login plugin — it's the least code, cross-platform, already frontend-callable, and idiomatic for a Wails app. Add a toggle in the tray menu and/or the settings UI. Fall back to a native implementation only if the plugin doesn't work with your alpha version.

Want me to implement it?

---

missing update button on upstream skills

---

add a sponsor section on this project with many tiers, with a nice "Made in Switzerland", add that this is a side project that takes most of my free time and would be nice to get some ~beer~ coke zero money to motivate me more to work on this project

---

will our app work with wsl?

---

can we somehow render the markdown content of the SKILL.md of each skill when we click it?

---

do we support private repositories? can you add support to bitbucket?

---

could we have a feature to sync my scribe configuration to a github repository with the following features:
- default name of the repository for storing skills can be changed but let's name it skills;
- can be a private or public repository
- should be part of the onboarding procedure, but can be an optional skippable step
- we should be able to configure this later in the settings button
- the cli should support this as well
- we need to have a button(s) to sync the config (pull/push), but in a way that should be resilient to conflicts
Ask me about any questions you might have 