package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

// TestVersionCommand tests the version command
func TestVersionCommand(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runVersion(versionCmd, []string{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "scribe version") {
		t.Error("expected version output to contain 'scribe version'")
	}
	if !strings.Contains(output, scribe.Version) {
		t.Errorf("expected version output to contain %q", scribe.Version)
	}
}
