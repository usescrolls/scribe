package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// keystrokeReader simulates terminal raw mode by yielding one keystroke per
// Read call: single bytes for normal keys, 3-byte sequences for arrow keys.
type keystrokeReader struct {
	data []byte
	pos  int
}

func newKeystrokeReader(input string) *keystrokeReader {
	return &keystrokeReader{data: []byte(input)}
}

func (r *keystrokeReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	// Arrow key escape sequence: return all 3 bytes together
	if r.data[r.pos] == 0x1b && r.pos+2 < len(r.data) && r.data[r.pos+1] == '[' {
		n := copy(p, r.data[r.pos:r.pos+3])
		r.pos += 3
		return n, nil
	}
	// Single byte keypress
	p[0] = r.data[r.pos]
	r.pos++
	return 1, nil
}

// ---------------------------------------------------------------------------
// Tests for allIndices
// ---------------------------------------------------------------------------

func TestAllIndices(t *testing.T) {
	tests := []struct {
		n    int
		want []int
	}{
		{0, []int{}},
		{1, []int{0}},
		{3, []int{0, 1, 2}},
	}
	for _, tt := range tests {
		got := allIndices(tt.n)
		if tt.n == 0 && len(got) != 0 {
			t.Errorf("allIndices(%d) = %v, want empty", tt.n, got)
		} else if tt.n > 0 && !reflect.DeepEqual(got, tt.want) {
			t.Errorf("allIndices(%d) = %v, want %v", tt.n, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests for promptMultiSelect (non-interactive fallback)
// ---------------------------------------------------------------------------

func TestPromptMultiSelectEmpty(t *testing.T) {
	got, err := promptMultiSelect(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestPromptMultiSelectNonInteractive(t *testing.T) {
	// In tests stdin is a pipe, so promptMultiSelect should fall back to
	// returning all indices without blocking.
	items := []selectableItem{
		{Label: "a", Description: "first", Selected: true},
		{Label: "b", Description: "second", Selected: true},
		{Label: "c", Description: "third", Selected: true},
	}
	got, err := promptMultiSelect(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0, 1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// ---------------------------------------------------------------------------
// Tests for runMultiSelect (interactive logic via piped reader)
// ---------------------------------------------------------------------------

func TestRunMultiSelectEnterImmediately(t *testing.T) {
	items := []selectableItem{
		{Label: "a", Description: "first", Selected: true},
		{Label: "b", Description: "second", Selected: true},
	}
	// Simulate pressing Enter right away — all items selected
	r := newKeystrokeReader("\r")
	w := &bytes.Buffer{}

	got, err := runMultiSelect(items, r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0, 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRunMultiSelectDeselectAllThenEnter(t *testing.T) {
	items := []selectableItem{
		{Label: "a", Description: "first", Selected: true},
		{Label: "b", Description: "second", Selected: true},
	}
	// Press 'a' (deselect all) then Enter
	r := newKeystrokeReader("a\r")
	w := &bytes.Buffer{}

	got, err := runMultiSelect(items, r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no selections, got %v", got)
	}
}

func TestRunMultiSelectDeselectAllReselectAllThenEnter(t *testing.T) {
	items := []selectableItem{
		{Label: "a", Description: "first", Selected: true},
		{Label: "b", Description: "second", Selected: true},
	}
	// Press 'a' (deselect all), 'a' again (reselect all), Enter
	r := newKeystrokeReader("aa\r")
	w := &bytes.Buffer{}

	got, err := runMultiSelect(items, r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0, 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRunMultiSelectToggleSingle(t *testing.T) {
	items := []selectableItem{
		{Label: "a", Description: "first", Selected: true},
		{Label: "b", Description: "second", Selected: true},
		{Label: "c", Description: "third", Selected: true},
	}
	// Space (deselect item 0), Enter — only items 1 and 2 selected
	r := newKeystrokeReader(" \r")
	w := &bytes.Buffer{}

	got, err := runMultiSelect(items, r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRunMultiSelectNavigateAndToggle(t *testing.T) {
	items := []selectableItem{
		{Label: "a", Description: "first", Selected: true},
		{Label: "b", Description: "second", Selected: true},
		{Label: "c", Description: "third", Selected: true},
	}
	// Down arrow (cursor to 1), Space (deselect item 1), Enter
	input := "\x1b[B \r"
	r := newKeystrokeReader(input)
	w := &bytes.Buffer{}

	got, err := runMultiSelect(items, r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0, 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRunMultiSelectUpArrowBoundary(t *testing.T) {
	items := []selectableItem{
		{Label: "a", Description: "first", Selected: true},
		{Label: "b", Description: "second", Selected: false},
	}
	// Up arrow at top (no-op), Space (toggle item 0 off), Enter
	input := "\x1b[A \r"
	r := newKeystrokeReader(input)
	w := &bytes.Buffer{}

	got, err := runMultiSelect(items, r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Item 0 toggled off, item 1 was already off
	if len(got) != 0 {
		t.Errorf("expected no selections, got %v", got)
	}
}

func TestRunMultiSelectDownArrowBoundary(t *testing.T) {
	items := []selectableItem{
		{Label: "a", Description: "first", Selected: true},
		{Label: "b", Description: "second", Selected: true},
	}
	// Down, Down (stays at 1), Space (deselect item 1), Enter
	input := "\x1b[B\x1b[B \r"
	r := newKeystrokeReader(input)
	w := &bytes.Buffer{}

	got, err := runMultiSelect(items, r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRunMultiSelectCtrlC(t *testing.T) {
	items := []selectableItem{
		{Label: "a", Description: "first", Selected: true},
	}
	r := newKeystrokeReader("\x03") // Ctrl+C
	w := &bytes.Buffer{}

	_, err := runMultiSelect(items, r, w)
	if err == nil {
		t.Fatal("expected error on Ctrl+C")
	}
	if !strings.Contains(err.Error(), "canceled") {
		t.Errorf("expected 'canceled' in error, got: %v", err)
	}
}

func TestRunMultiSelectLongDescription(t *testing.T) {
	items := []selectableItem{
		{Label: "a", Description: strings.Repeat("x", 80), Selected: true},
	}
	r := newKeystrokeReader("\r")
	w := &bytes.Buffer{}

	got, err := runMultiSelect(items, r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, []int{0}) {
		t.Errorf("got %v, want [0]", got)
	}
	// Verify description was truncated in output
	if !strings.Contains(w.String(), "...") {
		t.Error("expected truncated description with '...' in output")
	}
}

func TestRunMultiSelectReaderError(t *testing.T) {
	items := []selectableItem{
		{Label: "a", Description: "first", Selected: true},
	}
	// Empty reader will return io.EOF
	r := newKeystrokeReader("")
	w := &bytes.Buffer{}

	_, err := runMultiSelect(items, r, w)
	if err == nil {
		t.Fatal("expected error on empty reader")
	}
	if err != io.EOF {
		t.Errorf("expected io.EOF, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Integration: install flow with interactive selection (non-interactive stdin)
// ---------------------------------------------------------------------------

func TestRunInstallMultipleSkillsNonInteractive(t *testing.T) {
	homeDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	// Key: installYes is false, quiet is false, so the interactive prompt
	// would fire — but stdin is a pipe, so the non-interactive fallback
	// selects all skills automatically.
	quiet = false
	installYes = false
	installListOnly = false
	installSkills = ""
	installAll = false

	// Create agent config dir
	_ = os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755)

	// Create source with multiple skills
	tmpSrc, err := os.MkdirTemp("", "scribe-multiselect-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpSrc) }()

	skill1Dir := filepath.Join(tmpSrc, "skills", "alpha")
	skill2Dir := filepath.Join(tmpSrc, "skills", "beta")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.MkdirAll(skill2Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: alpha\ndescription: Alpha skill\n---\n\nAlpha\n"), 0o644)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("---\nname: beta\ndescription: Beta skill\n---\n\nBeta\n"), 0o644)

	output := captureStdout(t, func() {
		err = runInstall(installCmd, []string{tmpSrc})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both skills should be in output and installed
	if !strings.Contains(output, "alpha") {
		t.Errorf("expected 'alpha' in output, got: %s", output)
	}
	if !strings.Contains(output, "beta") {
		t.Errorf("expected 'beta' in output, got: %s", output)
	}
}
