package cli

import (
	"os"
	"strings"
	"testing"

	scribe "github.com/usescrolls/scribe/internal"
)

// TestCLICommands tests that CLICommands returns the expected commands
func TestCLICommands(t *testing.T) {
	commands := CLICommands()

	expected := []string{"install", "uninstall", "remove", "rm", "list", "ls", "info", "version", "help", "workspace", "check", "update", "cache"}

	for _, exp := range expected {
		found := false
		for _, cmd := range commands {
			if cmd == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected command %q not found in CLICommands()", exp)
		}
	}
}

func TestExecuteVersionCommand(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// Mark onboarding as complete so it doesn't interfere
	_ = scribe.CompleteOnboarding()

	// Execute with version command - should succeed
	os.Args = []string{"scribe", "version"}
	output := captureStdout(t, func() {
		code := Execute()
		if code != ExitSuccess {
			t.Errorf("Execute() returned %d, expected %d", code, ExitSuccess)
		}
	})
	if !strings.Contains(output, "scribe version") {
		t.Errorf("expected version output, got: %s", output)
	}
}

func TestExecuteInvalidCommand(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	_ = scribe.CompleteOnboarding()

	os.Args = []string{"scribe", "nonexistent-command"}
	quiet = false
	code := Execute()
	if code != ExitError {
		t.Errorf("Execute() returned %d, expected %d for invalid command", code, ExitError)
	}
}

func TestExecuteQuietMode(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	_ = scribe.CompleteOnboarding()

	os.Args = []string{"scribe", "--quiet", "nonexistent-command"}
	quiet = true
	output := captureStdout(t, func() {
		_ = Execute()
	})
	// In quiet mode stderr output should be suppressed
	// stdout should have nothing from Execute itself
	_ = output // we mainly care that it doesn't panic
}

func TestExecuteHelpCommand(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	_ = scribe.CompleteOnboarding()

	os.Args = []string{"scribe", "help"}
	output := captureStdout(t, func() {
		code := Execute()
		if code != ExitSuccess {
			t.Errorf("Execute() returned %d, expected %d", code, ExitSuccess)
		}
	})

	if !strings.Contains(output, "Scribe CLI provides") {
		t.Errorf("expected help text, got: %s", output)
	}
}
