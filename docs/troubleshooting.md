# Troubleshooting

This guide covers common issues and their solutions.

## Scribe Not Starting

**macOS/Linux:**
1. Check logs: `/tmp/scribe.log`
2. Try running with debug: `./scribe -debug`

**Windows:**
1. Try running with debug: `.\scribe.exe -debug`
2. Check Windows Event Viewer for application errors

---

## Plugins Not Appearing in Claude Code

1. Verify marketplace is registered: Check `~/.claude/settings.json` for `extraKnownMarketplaces.scribe`
2. Verify marketplace.json exists: `cat ~/.scribe/.claude-plugin/marketplace.json`
3. Run `/plugin` in Claude Code to refresh

---

## URL Scheme Not Working (agenthub:// links)

**macOS:**
- URL scheme is registered via Info.plist in the app bundle
- Ensure you're running the .app bundle, not the raw binary

**Linux:**
- Check desktop entry exists: `cat ~/.local/share/applications/scribe.desktop`
- Re-register: `xdg-mime default scribe.desktop x-scheme-handler/agenthub`
- Update database: `update-desktop-database ~/.local/share/applications/`

**Windows:**
- Verify registry entry: `reg query HKCR\agenthub`
- Re-run the installer: `.\install.ps1` or `.\install.ps1 -UserInstall`
- Check that the binary path in the registry matches the actual install location
