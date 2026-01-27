package main

import (
	"os"
	"path/filepath"
	"sync"
)

// Server holds the middleware server state
type Server struct {
	mu        sync.RWMutex
	registry  map[string]RegistryEntry
	hubDir    string // ~/.scribe
	claudeDir string // ~/.claude
}

// NewServer creates a new middleware server
func NewServer() *Server {
	homeDir, _ := os.UserHomeDir()
	return &Server{
		registry:  make(map[string]RegistryEntry),
		hubDir:    filepath.Join(homeDir, HubDirName),
		claudeDir: filepath.Join(homeDir, ".claude"),
	}
}

// Global server instance for systray callbacks
var server *Server
