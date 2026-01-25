//go:build linux

package main

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ipcListener net.Listener
	ipcMu       sync.Mutex
)

func init() {
	TrySendToRunningInstance = trySendToRunningInstanceLinux
	CleanupIPC = cleanupIPCLinux
}

// getSocketPath returns the path to the IPC socket
func getSocketPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, HubDirName, IPCSocketName)
}

// trySendToRunningInstanceLinux attempts to send URL to running instance via Unix socket
func trySendToRunningInstanceLinux(url string) bool {
	socketPath := getSocketPath()

	// Set connection timeout
	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
	if err != nil {
		// No instance running or socket stale
		logger.Debug("no running instance found", "socket", socketPath, "error", err)
		return false
	}
	defer conn.Close()

	// Set write deadline
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send URL with newline delimiter
	_, err = conn.Write([]byte(url + "\n"))
	if err != nil {
		logger.Warn("failed to send URL to running instance", "error", err)
		return false
	}

	// Wait for acknowledgment
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		logger.Warn("failed to read response from running instance", "error", err)
		return false
	}

	return strings.TrimSpace(response) == "OK"
}

// StartIPCServer starts the Unix socket server for receiving URLs from new instances
func StartIPCServer() error {
	socketPath := getSocketPath()

	// Ensure parent directory exists
	socketDir := filepath.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return err
	}

	// Remove stale socket if exists
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}

	ipcMu.Lock()
	ipcListener = listener
	ipcMu.Unlock()

	// Set socket permissions (owner only)
	os.Chmod(socketPath, 0600)

	logger.Info("IPC server started", "socket", socketPath)

	go acceptIPCConnections(listener)

	return nil
}

func acceptIPCConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Check if listener was closed
			ipcMu.Lock()
			closed := ipcListener == nil
			ipcMu.Unlock()
			if closed {
				return
			}
			logger.Warn("IPC accept error", "error", err)
			continue
		}
		go handleIPCConnection(conn)
	}
}

func handleIPCConnection(conn net.Conn) {
	defer conn.Close()

	// Set read deadline
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	reader := bufio.NewReader(conn)
	urlStr, err := reader.ReadString('\n')
	if err != nil {
		logger.Warn("IPC read error", "error", err)
		return
	}

	urlStr = strings.TrimSpace(urlStr)

	if strings.HasPrefix(urlStr, "agenthub://") {
		logger.Info("received URL via IPC", "url", urlStr)
		handleURLScheme(urlStr)
		conn.Write([]byte("OK\n"))
	} else {
		logger.Warn("invalid URL received via IPC", "url", urlStr)
		conn.Write([]byte("ERROR: invalid URL\n"))
	}
}

func cleanupIPCLinux() {
	ipcMu.Lock()
	defer ipcMu.Unlock()

	if ipcListener != nil {
		ipcListener.Close()
		ipcListener = nil
	}

	// Remove socket file
	socketPath := getSocketPath()
	os.Remove(socketPath)

	logger.Debug("IPC cleanup completed")
}

// RegisterURLSchemeHandler on Linux starts the IPC server
func RegisterURLSchemeHandler() {
	if err := StartIPCServer(); err != nil {
		logger.Error("failed to start IPC server", "error", err)
	}
}
