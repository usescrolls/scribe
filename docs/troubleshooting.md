# Troubleshooting

This guide covers common issues and their solutions.

## Scribe Not Starting

**macOS/Linux:**
1. Check logs: `/tmp/scribe.log`
2. Try running with debug: `./scribe --debug`
3. Verify permissions: `ls -la ~/.scribe/`

**Windows:**
1. Try running with debug: `.\scribe.exe --debug`
2. Check Windows Event Viewer for application errors

---

## Skills Not Appearing in Agent

1. Verify the skill is installed:
   ```bash
   scribe list
   ```

2. Check symlinks exist in agent's skills directory:
   ```bash
   ls -la ~/.claude/skills/    # For Claude Code
   ls -la ~/.cursor/skills/    # For Cursor
   ```

3. Verify the skill is in the active workspace:
   ```bash
   scribe workspace current
   ```

4. Try switching workspaces:
   ```bash
   scribe workspace use default
   ```

---

## Symlink Issues

**Broken symlinks:**
```bash
# Check for broken symlinks
find ~/.claude/skills -type l ! -exec test -e {} \; -print

# Reinstall the skill
scribe uninstall <skill-name>
scribe install <source>
```

**Permission denied:**
```bash
# Check directory permissions
ls -la ~/.scribe/scrolls/
ls -la ~/.claude/skills/

# Fix permissions
chmod 755 ~/.scribe/scrolls/
```

**Windows symlink issues:**
- Run Scribe as Administrator, or
- Enable Developer Mode in Windows Settings

---

## URL Scheme Not Working (agenthub:// links)

**macOS:**
- URL scheme is registered via Info.plist in the app bundle
- Ensure you're running the .app bundle, not the raw binary
- Try re-registering: `open -a Scribe`

**Linux:**
- Check desktop entry exists: `cat ~/.local/share/applications/scribe.desktop`
- Re-register: `xdg-mime default scribe.desktop x-scheme-handler/agenthub`
- Update database: `update-desktop-database ~/.local/share/applications/`

**Windows:**
- Verify registry entry: `reg query HKCR\agenthub`
- Re-run the installer: `.\install.ps1` or `.\install.ps1 -UserInstall`
- Check that the binary path in the registry matches the actual install location

---

## Installation Failures

**GitHub source issues:**
```bash
# Verify git is installed (needed for credential helper)
git --version

# Test cloning manually
git clone https://github.com/owner/repo /tmp/test-clone

# Check for authentication issues
gh auth status
```

**Zip URL issues:**
```bash
# Test download manually
curl -fsSL https://example.com/skills.zip -o /tmp/test.zip
unzip -l /tmp/test.zip
```

**Local path issues:**
```bash
# Verify path exists
ls -la ./path/to/skills/

# Check for SKILL.md files
find ./path/to/skills -name "SKILL.md"
```

---

## Private Repository Authentication

Scribe uses your existing git credentials to access private repositories. If you see errors like "authentication required", "permission denied", or "repository not found", your credentials are likely not configured.

**GitHub (recommended):**
```bash
# Install GitHub CLI and authenticate
gh auth login

# Verify authentication
gh auth status
```

**GitLab:**
```bash
# Using glab CLI
glab auth login

# Or configure a personal access token in git
git config --global credential.helper store
# Then clone any GitLab repo once to store the token
```

**SSH (all providers):**
```bash
# Ensure your SSH key is loaded
ssh-add -l

# If no keys listed, add your key
ssh-add ~/.ssh/id_ed25519

# Test SSH access
ssh -T git@github.com

# Use SSH URLs with Scribe
scribe install git@github.com:owner/private-repo.git
```

**How it works:** For HTTPS URLs, Scribe calls `git credential fill` which queries your system's git credential helper. For SSH URLs (`git@host:owner/repo.git`), Scribe connects through your SSH agent. Scribe does not store any credentials itself.

---

## Workspace Issues

**Cannot switch workspace:**
```bash
# Check workspace exists
scribe workspace list

# Create if missing
scribe workspace create default
```

**Skills not syncing after workspace switch:**
```bash
# Force resync by reinstalling
scribe workspace use default
scribe uninstall <skill-name>
scribe install <source>
```

---

## Agent Detection Issues

**Agent not detected:**
- Scribe only detects agents with existing config directories
- Install and run the agent at least once to create its config directory

**Check detected agents:**
```bash
# List detected agents (in GUI or via debug mode)
scribe --debug
```

**Manual verification:**
```bash
# Check if agent config directories exist
ls -la ~/.claude/      # Claude Code
ls -la ~/.cursor/      # Cursor
ls -la ~/.cline/       # Cline
ls -la ~/.continue/    # Continue
```

---

## Update/Check Issues

**"No updates available" but content changed:**
- Scribe uses content hashes to detect changes
- If the source was modified but the SKILL.md content is the same, no update is detected

**Update fails:**
```bash
# Check source is still accessible
scribe info <skill-name>

# Manually reinstall
scribe uninstall <skill-name>
scribe install <original-source>
```

---

## Debug Mode

For detailed logging, run Scribe with the `--debug` flag:

```bash
scribe --debug
scribe install owner/repo --debug
scribe workspace use backend --debug
```

This will output detailed information about:
- Agent detection
- Skill discovery
- Symlink creation
- Workspace operations
