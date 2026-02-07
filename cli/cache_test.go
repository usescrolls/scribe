package cli

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Tests for cache commands (cache.go)
// ---------------------------------------------------------------------------

func TestCacheClearCommand(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	output := captureStdout(t, func() {
		err := cacheClearCmd.RunE(cacheClearCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Cache cleared") {
		t.Errorf("expected 'Cache cleared', got: %s", output)
	}
}

func TestCacheClearCommandQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	output := captureStdout(t, func() {
		err := cacheClearCmd.RunE(cacheClearCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Cache cleared") {
		t.Error("quiet mode should suppress message")
	}
}

func TestCachePathCommand(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)

	output := captureStdout(t, func() {
		err := cachePathCmd.RunE(cachePathCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, ".scribe") || !strings.Contains(output, "cache") {
		t.Errorf("expected cache path containing '.scribe' and 'cache', got: %s", output)
	}
}
