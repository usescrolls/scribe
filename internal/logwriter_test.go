package scribe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestRotatingLogWriter_New(t *testing.T) {
	_ = setupTempHome(t)
	_ = EnsureScribeDirs()

	w, err := NewRotatingLogWriter()
	if err != nil {
		t.Fatalf("NewRotatingLogWriter() error: %v", err)
	}
	defer func() { _ = w.Close() }()

	if w.file == nil {
		t.Fatal("file is nil")
	}

	// Verify log file was created
	scribeDir, _ := GetScribeDir()
	logPath := filepath.Join(scribeDir, LogFileName)
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log file not created: %v", err)
	}
}

func TestRotatingLogWriter_Write(t *testing.T) {
	tmpDir := setupTempHome(t)
	_ = EnsureScribeDirs()

	w, err := NewRotatingLogWriter()
	if err != nil {
		t.Fatalf("NewRotatingLogWriter() error: %v", err)
	}
	defer func() { _ = w.Close() }()

	msg := "test log message\n"
	n, err := w.Write([]byte(msg))
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	if n != len(msg) {
		t.Errorf("Write() = %d, want %d", n, len(msg))
	}

	logPath := filepath.Join(tmpDir, ScribeDirName, LogFileName)
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(data) != msg {
		t.Errorf("log content = %q, want %q", string(data), msg)
	}
}

func TestRotatingLogWriter_Rotation(t *testing.T) {
	tmpDir := setupTempHome(t)
	_ = EnsureScribeDirs()

	w, err := NewRotatingLogWriter()
	if err != nil {
		t.Fatalf("NewRotatingLogWriter() error: %v", err)
	}
	defer func() { _ = w.Close() }()

	// Write enough data to trigger rotation (exceed 10 MB)
	line := strings.Repeat("X", 99) + "\n" // 100 bytes per line
	totalLines := (MaxLogSize / 100) + 100
	for i := 0; i < totalLines; i++ {
		_, err := fmt.Fprintf(w, "[%06d] %s", i, line)
		if err != nil {
			t.Fatalf("Write() error at line %d: %v", i, err)
		}
	}

	// File should be <= MaxLogSize after rotation
	logPath := filepath.Join(tmpDir, ScribeDirName, LogFileName)
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if info.Size() > MaxLogSize {
		t.Errorf("log file size = %d, want <= %d", info.Size(), MaxLogSize)
	}

	// File should start at a line boundary (no partial first line)
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if len(data) > 0 && data[0] == '\n' {
		t.Error("log file starts with newline (should start at line boundary)")
	}
}

func TestRotatingLogWriter_ConcurrentWrites(t *testing.T) {
	_ = setupTempHome(t)
	_ = EnsureScribeDirs()

	w, err := NewRotatingLogWriter()
	if err != nil {
		t.Fatalf("NewRotatingLogWriter() error: %v", err)
	}
	defer func() { _ = w.Close() }()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = fmt.Fprintf(w, "[goroutine-%d] message %d\n", id, j)
			}
		}(i)
	}
	wg.Wait()
}

func TestGetLogPath(t *testing.T) {
	_ = setupTempHome(t)

	logPath, err := GetLogPath()
	if err != nil {
		t.Fatalf("GetLogPath() error: %v", err)
	}

	if !strings.HasSuffix(logPath, filepath.Join(ScribeDirName, LogFileName)) {
		t.Errorf("GetLogPath() = %q, want suffix %q", logPath, filepath.Join(ScribeDirName, LogFileName))
	}
}
