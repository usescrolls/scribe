package scribe

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// authForSource resolves git credentials for a source.
// It tries the system git credential helper first (picks up gh auth, macOS Keychain, etc.),
// then falls back to SSH agent for SSH URLs, and returns nil for public repos.
func authForSource(source *SourceInfo) transport.AuthMethod {
	if isSSHURL(source.URL) {
		auth, err := gitssh.NewSSHAgentAuth("git")
		if err != nil {
			Logger.Debug("ssh agent auth not available", "error", err)
			return nil
		}
		return auth
	}

	// Try git credential helper for HTTPS URLs
	auth, err := gitCredentialFill(source.URL)
	if err != nil {
		Logger.Debug("git credential fill failed", "error", err)
		return nil
	}
	return auth
}

// gitCredentialFill uses the system git credential helper to resolve credentials.
// This picks up credentials from gh auth, macOS Keychain, Windows Credential Manager,
// git-credential-store, and any other configured credential helper.
func gitCredentialFill(cloneURL string) (*githttp.BasicAuth, error) {
	// Parse host from URL
	host := extractHost(cloneURL)
	if host == "" {
		return nil, fmt.Errorf("cannot extract host from URL: %s", cloneURL)
	}

	// Check that git is available
	if _, err := exec.LookPath("git"); err != nil {
		return nil, fmt.Errorf("git not found in PATH: %w", err)
	}

	cmd := exec.Command("git", "credential", "fill")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("protocol=https\nhost=%s\n\n", host))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git credential fill failed: %w", err)
	}

	username, password := parseCredentialOutput(string(output))
	if username == "" || password == "" {
		return nil, fmt.Errorf("no credentials returned for %s", host)
	}

	return &githttp.BasicAuth{
		Username: username,
		Password: password,
	}, nil
}

// parseCredentialOutput parses the key=value output from git credential fill.
func parseCredentialOutput(output string) (username, password string) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if k, v, ok := strings.Cut(line, "="); ok {
			switch k {
			case "username":
				username = v
			case "password":
				password = v
			}
		}
	}
	return
}

// extractHost extracts the hostname from a URL string.
func extractHost(rawURL string) string {
	// Handle https://host/... format
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(rawURL, prefix) {
			rest := strings.TrimPrefix(rawURL, prefix)
			if idx := strings.IndexByte(rest, '/'); idx != -1 {
				return rest[:idx]
			}
			return rest
		}
	}
	return ""
}

// isSSHURL returns true if the URL uses SSH format (git@host:owner/repo).
func isSSHURL(url string) bool {
	return strings.HasPrefix(url, "git@")
}

// IsAuthError returns true if the error message suggests an authentication failure.
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	authPatterns := []string{
		"authentication required",
		"authentication failed",
		"could not read username",
		"permission denied",
		"repository not found",
		"403",
		"401",
		"access denied",
		"invalid credentials",
	}
	for _, pattern := range authPatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

// AuthHintMessage returns a user-friendly hint for resolving authentication issues.
func AuthHintMessage() string {
	return `For private repositories, ensure git credentials are configured:
  - GitHub: run "gh auth login"
  - GitLab: run "glab auth login" or configure a personal access token
  - SSH: ensure your key is added to ssh-agent ("ssh-add")`
}
