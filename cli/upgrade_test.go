package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

func TestUpgradeCommand_DevBuild(t *testing.T) {
	old := scribe.Version
	scribe.Version = "dev"
	defer func() { scribe.Version = old }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := runUpgrade(upgradeCmd, []string{})

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	if err == nil {
		t.Fatal("expected error for dev build")
	}
	if !strings.Contains(err.Error(), "cannot upgrade development builds") {
		t.Errorf("unexpected error: %q", err.Error())
	}
}

func TestUpgradeCommand_DevSuffixBuild(t *testing.T) {
	old := scribe.Version
	scribe.Version = "1.17.0-dev"
	defer func() { scribe.Version = old }()

	err := runUpgrade(upgradeCmd, []string{})

	if err == nil {
		t.Fatal("expected error for dev-suffixed build")
	}
	if !strings.Contains(err.Error(), "cannot upgrade development builds") {
		t.Errorf("unexpected error: %q", err.Error())
	}
}
