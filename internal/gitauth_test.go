package scribe

import (
	"errors"
	"testing"
)

// ============================================================================
// isSSHURL (gitauth.go)
// ============================================================================

func TestIsSSHURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"git@github.com:owner/repo.git", true},
		{"git@gitlab.com:owner/repo.git", true},
		{"git@bitbucket.org:owner/repo.git", true},
		{"breitling-code@breitling-code.ghe.com:Breitling-Digital/skills.git", true},
		{"ssh://git@github.internal.example.com/owner/repo.git", true},
		{"https://github.com/owner/repo", false},
		{"http://github.com/owner/repo", false},
		{"", false},
		{"github.com/owner/repo", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := isSSHURL(tt.url); got != tt.want {
				t.Errorf("isSSHURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

// ============================================================================
// extractHost (gitauth.go)
// ============================================================================

func TestExtractHost(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://github.com/owner/repo", "github.com"},
		{"https://gitlab.com/owner/repo", "gitlab.com"},
		{"http://example.com/path", "example.com"},
		{"https://github.com", "github.com"},
		{"git@github.com:owner/repo", ""},
		{"", ""},
		{"ftp://example.com", ""},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := extractHost(tt.url); got != tt.want {
				t.Errorf("extractHost(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestHostFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"git@github.com:owner/repo.git", "github.com"},
		{"breitling-code@breitling-code.ghe.com:Breitling-Digital/skills.git", "breitling-code.ghe.com"},
		{"ssh://git@github.internal.example.com:2222/owner/repo.git", "github.internal.example.com:2222"},
		{"https://github.com/owner/repo", "github.com"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := hostFromURL(tt.url); got != tt.want {
				t.Errorf("hostFromURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestSSHUserFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"git@github.com:owner/repo.git", "git"},
		{"breitling-code@breitling-code.ghe.com:Breitling-Digital/skills.git", "breitling-code"},
		{"ssh://deploy@github.internal.example.com/owner/repo.git", "deploy"},
		{"https://github.com/owner/repo", "git"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := sshUserFromURL(tt.url); got != tt.want {
				t.Errorf("sshUserFromURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

// ============================================================================
// parseCredentialOutput (gitauth.go)
// ============================================================================

func TestParseCredentialOutput(t *testing.T) {
	tests := []struct {
		name         string
		output       string
		wantUser     string
		wantPassword string
	}{
		{
			name:         "typical output",
			output:       "protocol=https\nhost=github.com\nusername=octocat\npassword=ghp_token123\n",
			wantUser:     "octocat",
			wantPassword: "ghp_token123",
		},
		{
			name:         "empty output",
			output:       "",
			wantUser:     "",
			wantPassword: "",
		},
		{
			name:         "only username",
			output:       "username=user\n",
			wantUser:     "user",
			wantPassword: "",
		},
		{
			name:         "only password",
			output:       "password=pass\n",
			wantUser:     "",
			wantPassword: "pass",
		},
		{
			name:         "extra fields ignored",
			output:       "protocol=https\nhost=github.com\nusername=u\npassword=p\npath=foo\n",
			wantUser:     "u",
			wantPassword: "p",
		},
		{
			name:         "line without equals",
			output:       "garbage line\nusername=u\npassword=p\n",
			wantUser:     "u",
			wantPassword: "p",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, pass := parseCredentialOutput(tt.output)
			if user != tt.wantUser {
				t.Errorf("username = %q, want %q", user, tt.wantUser)
			}
			if pass != tt.wantPassword {
				t.Errorf("password = %q, want %q", pass, tt.wantPassword)
			}
		})
	}
}

// ============================================================================
// IsAuthError (gitauth.go)
// ============================================================================

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"generic error", errors.New("something went wrong"), false},
		{"authentication required", errors.New("authentication required"), true},
		{"Authentication Failed", errors.New("Authentication Failed"), true},
		{"permission denied", errors.New("ssh: permission denied"), true},
		{"repository not found", errors.New("repository not found"), true},
		{"403 forbidden", errors.New("server returned 403"), true},
		{"401 unauthorized", errors.New("server returned 401"), true},
		{"access denied", errors.New("access denied to resource"), true},
		{"invalid credentials", errors.New("invalid credentials"), true},
		{"could not read username", errors.New("could not read username for https://github.com"), true},
		{"network timeout", errors.New("connection timed out"), false},
		{"dns error", errors.New("no such host"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAuthError(tt.err); got != tt.want {
				t.Errorf("IsAuthError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// ============================================================================
// AuthHintMessage (gitauth.go)
// ============================================================================

func TestAuthHintMessage(t *testing.T) {
	msg := AuthHintMessage()
	if msg == "" {
		t.Error("AuthHintMessage() returned empty string")
	}
	// Should mention common auth methods
	if len(msg) < 20 {
		t.Errorf("AuthHintMessage() too short: %q", msg)
	}
}

// ============================================================================
// authForSource (gitauth.go)
// ============================================================================

func TestAuthForSource_NonSSH_NoGit(t *testing.T) {
	// When git is not in PATH or no credentials configured,
	// authForSource returns nil for HTTPS URLs (it gracefully degrades)
	source := &SourceInfo{
		Type: "github",
		URL:  "https://github.com/owner/repo",
	}
	// This may or may not return auth depending on the test environment.
	// We just verify it doesn't panic.
	_ = authForSource(source)
}
