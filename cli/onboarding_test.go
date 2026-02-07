package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	scribe "github.com/usescrolls/scribe/internal"
)

// ---------------------------------------------------------------------------
// Tests for checkOnboarding (onboarding.go)
// ---------------------------------------------------------------------------

func TestCheckOnboardingNotCompleted(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	// Fresh HOME, onboarding not completed
	needed := checkOnboarding()
	if !needed {
		t.Error("expected checkOnboarding() to return true for fresh install")
	}
}

func TestCheckOnboardingCompleted(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	// Complete onboarding
	_ = scribe.CompleteOnboarding()

	needed := checkOnboarding()
	if needed {
		t.Error("expected checkOnboarding() to return false after completion")
	}
}

// ---------------------------------------------------------------------------
// Tests for runOnboardingIfNeeded (onboarding.go)
// ---------------------------------------------------------------------------

func TestRunOnboardingIfNeededAlreadyComplete(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	_ = scribe.CompleteOnboarding()

	err := runOnboardingIfNeeded()
	if err != nil {
		t.Errorf("expected nil error when onboarding already complete, got: %v", err)
	}
}

func TestRunOnboardingIfNeededNotComplete(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// onboarding is not completed in fresh HOME
	// runOnboardingIfNeeded should call runOnboarding which will fail
	// because no agents are detected
	output := captureStdout(t, func() {
		err := runOnboardingIfNeeded()
		if err == nil {
			t.Error("expected error from runOnboarding (no agents)")
		}
	})

	if !strings.Contains(output, "Welcome to Scribe") {
		t.Errorf("expected welcome message, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// Tests for runOnboarding (onboarding.go)
// ---------------------------------------------------------------------------

func TestRunOnboardingNoAgents(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// In temp HOME, no agents should be detected
	output := captureStdout(t, func() {
		err := runOnboarding()
		if err == nil {
			t.Error("expected error when no agents detected")
		}
		if !strings.Contains(err.Error(), "no coding agents detected") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Welcome to Scribe") {
		t.Errorf("expected welcome message, got: %s", output)
	}
	if !strings.Contains(output, "No coding agents detected") {
		t.Errorf("expected 'No coding agents detected', got: %s", output)
	}
}

func TestRunOnboardingWithAgentDetectedNoExistingSkills(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// Create agent config dir so agents are detected
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Redirect stdin to provide "n" input for the onboarding prompt
	// Since no existing skills, onboarding should proceed to install demo skill
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	// Write empty input (just close it to auto-skip any prompts)
	_ = w.Close()

	output := captureStdout(t, func() {
		err := runOnboarding()
		// Restore stdin before checking
		os.Stdin = oldStdin
		if err != nil {
			t.Logf("onboarding error (expected for limited test): %v", err)
		}
	})
	os.Stdin = oldStdin

	// Should show agent detection
	if !strings.Contains(output, "Detecting installed coding agents") {
		t.Errorf("expected agent detection message, got: %s", output)
	}
	if !strings.Contains(output, "Found") && !strings.Contains(output, "coding agent") {
		t.Errorf("expected found agents message, got: %s", output)
	}
}
