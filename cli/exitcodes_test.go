package cli

import "testing"

// TestExitCodes tests that exit codes are defined correctly
func TestExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitError", ExitError, 1},
		{"ExitUsage", ExitUsage, 2},
		{"ExitNotFound", ExitNotFound, 3},
		{"ExitSourceFailed", ExitSourceFailed, 4},
		{"ExitRegistryError", ExitRegistryError, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("%s = %d, expected %d", tt.name, tt.code, tt.expected)
			}
		})
	}
}
