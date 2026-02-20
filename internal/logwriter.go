package scribe

import (
	"os"
	"path/filepath"
	"sync"
)

const (
	// LogFileName is the unified log file name
	LogFileName = "scribe.log"
	// MaxLogSize is the maximum log file size in bytes (10 MB)
	MaxLogSize = 10 * 1024 * 1024
	// TruncateTarget is the size to keep after rotation (~7 MB)
	TruncateTarget = 7 * 1024 * 1024
)

// RotatingLogWriter is a thread-safe io.Writer that writes to a log file
// with automatic size-based rotation. When the file exceeds MaxLogSize,
// it truncates from the front, keeping the most recent TruncateTarget bytes
// aligned to a newline boundary.
type RotatingLogWriter struct {
	mu       sync.Mutex
	file     *os.File
	filePath string
	size     int64
}

// NewRotatingLogWriter opens (or creates) the log file and returns a writer.
func NewRotatingLogWriter() (*RotatingLogWriter, error) {
	scribeDir, err := GetScribeDir()
	if err != nil {
		return nil, err
	}
	if err := EnsureDir(scribeDir); err != nil {
		return nil, err
	}

	logPath := filepath.Join(scribeDir, LogFileName)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	return &RotatingLogWriter{
		file:     f,
		filePath: logPath,
		size:     info.Size(),
	}, nil
}

// Write implements io.Writer. It appends data to the log file and triggers
// rotation when the file exceeds MaxLogSize.
func (w *RotatingLogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err = w.file.Write(p)
	w.size += int64(n)

	if w.size > MaxLogSize {
		w.rotate()
	}
	return n, err
}

// rotate reads the file, keeps the last TruncateTarget bytes (aligned to
// a newline boundary), and rewrites the file.
func (w *RotatingLogWriter) rotate() {
	_ = w.file.Close()

	data, err := os.ReadFile(w.filePath)
	if err != nil {
		w.file, _ = os.OpenFile(w.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		return
	}

	if int64(len(data)) <= TruncateTarget {
		w.file, _ = os.OpenFile(w.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		w.size = int64(len(data))
		return
	}

	// Find the first newline after the cut point to preserve line boundaries
	cutPoint := int64(len(data)) - TruncateTarget
	for cutPoint < int64(len(data)) && data[cutPoint] != '\n' {
		cutPoint++
	}
	if cutPoint < int64(len(data)) {
		cutPoint++ // skip the newline itself
	}

	trimmed := data[cutPoint:]

	if err := os.WriteFile(w.filePath, trimmed, 0o644); err != nil {
		w.file, _ = os.OpenFile(w.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		return
	}

	w.file, _ = os.OpenFile(w.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	w.size = int64(len(trimmed))
}

// Close closes the underlying file.
func (w *RotatingLogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// GetLogPath returns the path to the log file.
func GetLogPath() (string, error) {
	scribeDir, err := GetScribeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(scribeDir, LogFileName), nil
}
