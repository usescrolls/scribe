package scribe

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

// HubDir returns the hub directory path
func (s *Server) HubDir() string {
	return s.hubDir
}

// NewTestServer creates a server with custom paths for testing
func NewTestServer(hubDir, claudeDir string) *Server {
	return &Server{
		registry:  make(map[string]RegistryEntry),
		hubDir:    hubDir,
		claudeDir: claudeDir,
	}
}

// SetRegistryEntry sets a registry entry (for testing)
func (s *Server) SetRegistryEntry(name string, entry RegistryEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registry[name] = entry
}

// GetRegistryEntry gets a registry entry (for testing)
func (s *Server) GetRegistryEntry(name string) (RegistryEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.registry[name]
	return entry, ok
}
