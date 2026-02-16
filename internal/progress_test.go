package scribe

import (
	"archive/zip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// progressRecorder collects ProgressEvents emitted during an operation
type progressRecorder struct {
	mu     sync.Mutex
	events []ProgressEvent
}

func (r *progressRecorder) emit(e ProgressEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
}

func (r *progressRecorder) get() []ProgressEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]ProgressEvent, len(r.events))
	copy(out, r.events)
	return out
}

// ============================================================================
// emitProgress helper
// ============================================================================

func TestEmitProgress_NilSlice(t *testing.T) {
	// Should not panic with nil slice
	emitProgress(nil, "discover", "parsing", "Parsing...", "")
}

func TestEmitProgress_EmptySlice(t *testing.T) {
	// Should not panic with empty slice
	emitProgress([]ProgressEmitter{}, "discover", "parsing", "Parsing...", "")
}

func TestEmitProgress_NilEmitter(t *testing.T) {
	// Should not panic with nil emitter in slice
	emitProgress([]ProgressEmitter{nil}, "discover", "parsing", "Parsing...", "")
}

func TestEmitProgress_CallsEmitter(t *testing.T) {
	var rec progressRecorder
	emitProgress([]ProgressEmitter{rec.emit}, "discover", "cloning", "Cloning repository...", "")

	events := rec.get()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	e := events[0]
	if e.Phase != "discover" {
		t.Errorf("Phase = %q, want 'discover'", e.Phase)
	}
	if e.Step != "cloning" {
		t.Errorf("Step = %q, want 'cloning'", e.Step)
	}
	if e.Message != "Cloning repository..." {
		t.Errorf("Message = %q, want 'Cloning repository...'", e.Message)
	}
	if e.Detail != "" {
		t.Errorf("Detail = %q, want ''", e.Detail)
	}
}

func TestEmitProgress_WithDetail(t *testing.T) {
	var rec progressRecorder
	emitProgress([]ProgressEmitter{rec.emit}, "install", "syncing", "Syncing to Claude Code...", "Claude Code")

	events := rec.get()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Detail != "Claude Code" {
		t.Errorf("Detail = %q, want 'Claude Code'", events[0].Detail)
	}
}

func TestEmitProgress_OnlyFirstEmitter(t *testing.T) {
	// emitProgress only calls emit[0]; verify second emitter is ignored
	var rec1, rec2 progressRecorder
	emitProgress([]ProgressEmitter{rec1.emit, rec2.emit}, "discover", "scanning", "Scanning...", "")

	if len(rec1.get()) != 1 {
		t.Errorf("first emitter: expected 1 event, got %d", len(rec1.get()))
	}
	if len(rec2.get()) != 0 {
		t.Errorf("second emitter: expected 0 events, got %d", len(rec2.get()))
	}
}

// ============================================================================
// FetchAndDiscoverSkills progress events
// ============================================================================

func TestFetchAndDiscoverSkills_Local_EmitsProgress(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-progress-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	InitLoggerCLI(false)

	skillContent := "---\nname: progress-test\ndescription: Progress test\n---\n# Test\n"
	_ = os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte(skillContent), 0o644)

	source := &SourceInfo{Type: "local", LocalPath: tmpDir}

	var rec progressRecorder
	skills, _, err := FetchAndDiscoverSkills(source, rec.emit)
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills() error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	events := rec.get()

	// Should emit: parsing, scanning, done
	assertHasStep(t, events, "discover", "parsing")
	assertHasStep(t, events, "discover", "scanning")
	assertHasStep(t, events, "discover", "done")

	// "done" event should mention the skill count
	doneEvent := findEvent(events, "discover", "done")
	if doneEvent == nil {
		t.Fatal("missing done event")
	}
	if doneEvent.Message != "Found 1 skill(s)" {
		t.Errorf("done message = %q, want 'Found 1 skill(s)'", doneEvent.Message)
	}
}

func TestFetchAndDiscoverSkills_GitHub_EmitsProgress(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	remoteDir := filepath.Join(tmpDir, "progress-repo")
	createTestGitRepo(t, remoteDir, map[string]string{
		"SKILL.md": "---\nname: progress-gh\ndescription: GitHub progress\n---\n# GH\n",
	})

	source := &SourceInfo{
		Type:  "github",
		Owner: "testuser",
		Repo:  "progress-test",
		URL:   remoteDir,
	}

	var rec progressRecorder
	skills, result, err := FetchAndDiscoverSkills(source, rec.emit)
	if result != nil {
		defer result.Cleanup()
	}
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills() error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	events := rec.get()

	// Should emit: parsing, cloning, scanning, done
	assertHasStep(t, events, "discover", "parsing")
	assertHasStep(t, events, "discover", "cloning")
	assertHasStep(t, events, "discover", "scanning")
	assertHasStep(t, events, "discover", "done")
}

func TestFetchAndDiscoverSkills_GitHub_CachedEmitsFetching(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	remoteDir := filepath.Join(tmpDir, "cached-repo")
	createTestGitRepo(t, remoteDir, map[string]string{
		"SKILL.md": "---\nname: cached-skill\ndescription: Cached\n---\n# Cached\n",
	})

	source := &SourceInfo{
		Type:  "github",
		Owner: "testuser",
		Repo:  "cached-test",
		URL:   remoteDir,
	}

	// First call — clones
	_, result1, err := FetchAndDiscoverSkills(source)
	if result1 != nil {
		defer result1.Cleanup()
	}
	if err != nil {
		t.Fatalf("first FetchAndDiscoverSkills() error: %v", err)
	}

	// Second call — should fetch (repo already cached)
	var rec progressRecorder
	_, result2, err := FetchAndDiscoverSkills(source, rec.emit)
	if result2 != nil {
		defer result2.Cleanup()
	}
	if err != nil {
		t.Fatalf("second FetchAndDiscoverSkills() error: %v", err)
	}

	events := rec.get()

	// Should emit "fetching" instead of "cloning" for a cached repo
	assertHasStep(t, events, "discover", "fetching")
	assertHasStep(t, events, "discover", "scanning")
	assertHasStep(t, events, "discover", "done")
}

func TestFetchAndDiscoverSkills_Zip_EmitsProgress(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)

	// Create a zip with a skill
	tmpDir, _ := os.MkdirTemp("", "scribe-zip-progress-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	zipPath := filepath.Join(tmpDir, "skills.zip")
	zipFile, _ := os.Create(zipPath)
	w := zip.NewWriter(zipFile)
	fw, _ := w.Create("SKILL.md")
	_, _ = fw.Write([]byte("---\nname: zip-progress\ndescription: Zip\n---\n# Zip\n"))
	_ = w.Close()
	_ = zipFile.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, zipPath)
	}))
	defer srv.Close()

	source := &SourceInfo{
		Type: "zip",
		URL:  srv.URL + "/skills.zip",
	}

	var rec progressRecorder
	skills, result, err := FetchAndDiscoverSkills(source, rec.emit)
	if result != nil {
		defer result.Cleanup()
	}
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills(zip) error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	events := rec.get()

	// Should emit: parsing, downloading, extracting, scanning, done
	assertHasStep(t, events, "discover", "parsing")
	assertHasStep(t, events, "discover", "downloading")
	assertHasStep(t, events, "discover", "extracting")
	assertHasStep(t, events, "discover", "scanning")
	assertHasStep(t, events, "discover", "done")
}

// ============================================================================
// InstallSkill progress events
// ============================================================================

func TestInstallSkill_EmitsProgress(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create agent dir so syncing has a target
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)

	srcDir := filepath.Join(tmpDir, "src-skill")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"),
		[]byte("---\nname: emit-test\ndescription: Emit test\n---\n# Emit\n"), 0o644)

	skill := &Skill{
		Name:        "emit-test",
		Description: "Emit test",
		Path:        srcDir,
	}
	source := &SourceInfo{
		Type:  "github",
		Owner: "testuser",
		Repo:  "testrepo",
		URL:   "https://github.com/testuser/testrepo",
	}

	var rec progressRecorder
	err := InstallSkill(skill, source, InstallOptions{}, nil, rec.emit)
	if err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	events := rec.get()

	// Should emit: copying, metadata, syncing (at least one agent)
	assertHasStep(t, events, "install", "copying")
	assertHasStep(t, events, "install", "metadata")
	assertHasStep(t, events, "install", "syncing")
}

func TestInstallSkill_SyncEmitsAgentName(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create claude agent dir
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)

	srcDir := filepath.Join(tmpDir, "agent-src")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"),
		[]byte("---\nname: agent-emit\ndescription: Agent emit\n---\n# Agent\n"), 0o644)

	skill := &Skill{
		Name:        "agent-emit",
		Description: "Agent emit",
		Path:        srcDir,
	}
	source := &SourceInfo{Type: "local", LocalPath: srcDir}

	var rec progressRecorder
	err := InstallSkill(skill, source, InstallOptions{Agents: []string{"claude-code"}}, nil, rec.emit)
	if err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	events := rec.get()

	// Find the syncing event and verify it has a Detail with the agent name
	syncEvent := findEvent(events, "install", "syncing")
	if syncEvent == nil {
		t.Fatal("missing syncing event")
	}
	if syncEvent.Detail != "Claude Code" {
		t.Errorf("syncing Detail = %q, want 'Claude Code'", syncEvent.Detail)
	}
}

// ============================================================================
// SyncSkillToAgents progress events
// ============================================================================

func TestSyncSkillToAgents_EmitsPerAgent(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create skill in scrolls
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "sync-emit")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Sync"), 0o644)

	// Create agent dirs
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, ".cursor"), 0o755)

	var rec progressRecorder
	err := SyncSkillToAgents("sync-emit", []string{"claude-code", "cursor"}, rec.emit)
	if err != nil {
		t.Fatalf("SyncSkillToAgents() error: %v", err)
	}

	events := rec.get()

	// Should emit one syncing event per agent
	syncEvents := filterEvents(events, "install", "syncing")
	if len(syncEvents) != 2 {
		t.Fatalf("expected 2 syncing events, got %d", len(syncEvents))
	}

	// Verify agent names appear in Detail
	details := map[string]bool{}
	for _, e := range syncEvents {
		details[e.Detail] = true
	}
	if !details["Claude Code"] {
		t.Error("missing syncing event for Claude Code")
	}
	if !details["Cursor"] {
		t.Error("missing syncing event for Cursor")
	}
}

// ============================================================================
// Event ordering
// ============================================================================

func TestFetchAndDiscoverSkills_EventOrder(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "scribe-order-test-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	InitLoggerCLI(false)

	_ = os.WriteFile(filepath.Join(tmpDir, "SKILL.md"),
		[]byte("---\nname: order-test\ndescription: Order\n---\n# Order\n"), 0o644)

	source := &SourceInfo{Type: "local", LocalPath: tmpDir}

	var rec progressRecorder
	_, _, err := FetchAndDiscoverSkills(source, rec.emit)
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills() error: %v", err)
	}

	events := rec.get()
	if len(events) < 3 {
		t.Fatalf("expected at least 3 events, got %d", len(events))
	}

	// Verify order: parsing must come before scanning, scanning before done
	parsingIdx := indexOfStep(events, "parsing")
	scanningIdx := indexOfStep(events, "scanning")
	doneIdx := indexOfStep(events, "done")

	if parsingIdx == -1 || scanningIdx == -1 || doneIdx == -1 {
		t.Fatal("missing expected events")
	}
	if parsingIdx >= scanningIdx {
		t.Errorf("parsing (idx %d) should come before scanning (idx %d)", parsingIdx, scanningIdx)
	}
	if scanningIdx >= doneIdx {
		t.Errorf("scanning (idx %d) should come before done (idx %d)", scanningIdx, doneIdx)
	}

	// All events should have phase "discover"
	for _, e := range events {
		if e.Phase != "discover" {
			t.Errorf("expected all events to have phase 'discover', got %q", e.Phase)
		}
	}
}

func TestInstallSkill_EventOrder(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)

	srcDir := filepath.Join(tmpDir, "order-src")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"),
		[]byte("---\nname: order-install\ndescription: Order\n---\n# O\n"), 0o644)

	skill := &Skill{Name: "order-install", Description: "Order", Path: srcDir}
	source := &SourceInfo{Type: "local", LocalPath: srcDir}

	var rec progressRecorder
	err := InstallSkill(skill, source, InstallOptions{}, nil, rec.emit)
	if err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	events := rec.get()

	// Verify order: copying → metadata → syncing
	copyIdx := indexOfStep(events, "copying")
	metaIdx := indexOfStep(events, "metadata")
	syncIdx := indexOfStep(events, "syncing")

	if copyIdx == -1 || metaIdx == -1 || syncIdx == -1 {
		t.Fatalf("missing expected events (copying=%d, metadata=%d, syncing=%d)", copyIdx, metaIdx, syncIdx)
	}
	if copyIdx >= metaIdx {
		t.Errorf("copying (idx %d) should come before metadata (idx %d)", copyIdx, metaIdx)
	}
	if metaIdx >= syncIdx {
		t.Errorf("metadata (idx %d) should come before syncing (idx %d)", metaIdx, syncIdx)
	}

	// All events should have phase "install"
	for _, e := range events {
		if e.Phase != "install" {
			t.Errorf("expected all events to have phase 'install', got %q", e.Phase)
		}
	}
}

// ============================================================================
// Without emitter (backward compat)
// ============================================================================

func TestFetchAndDiscoverSkills_WithoutEmitter(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "scribe-nonemit-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	InitLoggerCLI(false)

	_ = os.WriteFile(filepath.Join(tmpDir, "SKILL.md"),
		[]byte("---\nname: no-emit\ndescription: No emitter\n---\n# Test\n"), 0o644)

	source := &SourceInfo{Type: "local", LocalPath: tmpDir}

	// Calling without emitter should work exactly as before
	skills, _, err := FetchAndDiscoverSkills(source)
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills() without emitter error: %v", err)
	}
	if len(skills) != 1 || skills[0].Name != "no-emit" {
		t.Errorf("unexpected skills: %v", skills)
	}
}

func TestInstallSkill_WithoutEmitter(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	srcDir := filepath.Join(tmpDir, "no-emit-src")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"),
		[]byte("---\nname: no-emit-install\ndescription: No emit\n---\n# NE\n"), 0o644)

	skill := &Skill{Name: "no-emit-install", Description: "No emit", Path: srcDir}
	source := &SourceInfo{Type: "local", LocalPath: srcDir}

	// Calling without emitter should work exactly as before
	err := InstallSkill(skill, source, InstallOptions{}, nil)
	if err != nil {
		t.Fatalf("InstallSkill() without emitter error: %v", err)
	}
}

// ============================================================================
// Test helpers
// ============================================================================

func assertHasStep(t *testing.T, events []ProgressEvent, phase, step string) {
	t.Helper()
	for _, e := range events {
		if e.Phase == phase && e.Step == step {
			return
		}
	}
	t.Errorf("missing event with phase=%q step=%q in %d events", phase, step, len(events))
}

func findEvent(events []ProgressEvent, phase, step string) *ProgressEvent {
	for i := range events {
		if events[i].Phase == phase && events[i].Step == step {
			return &events[i]
		}
	}
	return nil
}

func filterEvents(events []ProgressEvent, phase, step string) []ProgressEvent {
	var out []ProgressEvent
	for _, e := range events {
		if e.Phase == phase && e.Step == step {
			out = append(out, e)
		}
	}
	return out
}

func indexOfStep(events []ProgressEvent, step string) int {
	for i, e := range events {
		if e.Step == step {
			return i
		}
	}
	return -1
}
