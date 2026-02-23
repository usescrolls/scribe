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
// Tests for checkAndPromptTerms (onboarding.go)
// ---------------------------------------------------------------------------

func TestCheckAndPromptTermsAlreadyAccepted(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	// Accept terms first
	_ = scribe.AcceptTerms()

	err := checkAndPromptTerms()
	if err != nil {
		t.Errorf("expected nil error when terms already accepted, got: %v", err)
	}
}

func TestCheckAndPromptTermsAcceptedViaPrompt(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	// Pipe "y\n" to accept terms
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("y\n")
	_ = w.Close()

	output := captureStdout(t, func() {
		err := checkAndPromptTerms()
		os.Stdin = oldStdin
		if err != nil {
			t.Errorf("expected nil error after accepting terms, got: %v", err)
		}
	})
	os.Stdin = oldStdin

	if !strings.Contains(output, "Terms & Conditions") {
		t.Errorf("expected terms update message, got: %s", output)
	}
	if !strings.Contains(output, "Terms accepted") {
		t.Errorf("expected acceptance confirmation, got: %s", output)
	}

	// Verify terms were persisted
	accepted, _ := scribe.AreTermsAccepted()
	if !accepted {
		t.Error("terms should be persisted after accepting via prompt")
	}
}

func TestCheckAndPromptTermsRejected(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	// Pipe "n\n" to reject terms
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("n\n")
	_ = w.Close()

	captureStdout(t, func() {
		err := checkAndPromptTerms()
		os.Stdin = oldStdin
		if err == nil {
			t.Error("expected error when terms rejected")
		}
		if !strings.Contains(err.Error(), "terms and conditions not accepted") {
			t.Errorf("unexpected error: %v", err)
		}
	})
	os.Stdin = oldStdin

	// Verify terms were NOT persisted
	accepted, _ := scribe.AreTermsAccepted()
	if accepted {
		t.Error("terms should not be accepted after rejection")
	}
}

func TestCheckAndPromptTermsDefaultReject(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	// Pipe just enter — default is N
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("\n")
	_ = w.Close()

	captureStdout(t, func() {
		err := checkAndPromptTerms()
		os.Stdin = oldStdin
		if err == nil {
			t.Error("expected error when terms rejected by default")
		}
	})
	os.Stdin = oldStdin
}

// ---------------------------------------------------------------------------
// Tests for runOnboarding (onboarding.go)
// ---------------------------------------------------------------------------

func TestRunOnboardingTermsRejected(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// Pipe "n\n" to reject terms
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("n\n")
	_ = w.Close()

	output := captureStdout(t, func() {
		err := runOnboarding()
		os.Stdin = oldStdin
		if err == nil {
			t.Error("expected error when terms rejected")
		}
		if !strings.Contains(err.Error(), "terms and conditions not accepted") {
			t.Errorf("unexpected error: %v", err)
		}
	})
	os.Stdin = oldStdin

	if !strings.Contains(output, "Terms & Conditions") {
		t.Errorf("expected terms display, got: %s", output)
	}
	if !strings.Contains(output, "You must accept the terms") {
		t.Errorf("expected rejection message, got: %s", output)
	}

	// Verify terms were NOT persisted
	accepted, _ := scribe.AreTermsAccepted()
	if accepted {
		t.Error("terms should not be accepted after rejection")
	}
}

func TestRunOnboardingTermsRejectedDefault(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// Pipe just enter (empty input) — default is N
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("\n")
	_ = w.Close()

	captureStdout(t, func() {
		err := runOnboarding()
		os.Stdin = oldStdin
		if err == nil {
			t.Error("expected error when terms rejected by default")
		}
		if !strings.Contains(err.Error(), "terms and conditions not accepted") {
			t.Errorf("unexpected error: %v", err)
		}
	})
	os.Stdin = oldStdin
}

func TestRunOnboardingTermsAcceptedPersisted(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// Pipe "yes\n" to accept terms (tests "yes" variant), then EOF
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("yes\n")
	_ = w.Close()

	captureStdout(t, func() {
		// Will fail later (no agents), but terms should be persisted
		_ = runOnboarding()
		os.Stdin = oldStdin
	})
	os.Stdin = oldStdin

	// Verify terms were persisted
	accepted, err := scribe.AreTermsAccepted()
	if err != nil {
		t.Fatalf("AreTermsAccepted() error: %v", err)
	}
	if !accepted {
		t.Error("terms should be accepted after answering 'yes'")
	}
}

func TestRunOnboardingNoAgents(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	// Pipe "y\n" to accept terms, then EOF for subsequent prompts
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("y\n")
	_ = w.Close()

	// In temp HOME, no agents should be detected
	output := captureStdout(t, func() {
		err := runOnboarding()
		os.Stdin = oldStdin
		if err == nil {
			t.Error("expected error when no agents detected")
		}
		if !strings.Contains(err.Error(), "no coding agents detected") {
			t.Errorf("unexpected error: %v", err)
		}
	})
	os.Stdin = oldStdin

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

	// Redirect stdin: "y\n" to accept terms, then EOF to auto-skip remaining prompts
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("y\n")
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
