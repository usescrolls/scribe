//go:build windows

package main

import (
	"bufio"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

var (
	ipcListener net.Listener
	ipcMu       sync.Mutex
	mutexHandle windows.Handle
)

func init() {
	TrySendToRunningInstance = trySendToRunningInstanceWindows
	CleanupIPC = cleanupIPCWindows
}

// trySendToRunningInstanceWindows checks for running instance via mutex,
// sends URL via named pipe if found.
func trySendToRunningInstanceWindows(url string) bool {
	// Try to create mutex - if it already exists, another instance is running
	mutexName, _ := windows.UTF16PtrFromString(IPCMutexName)
	handle, err := windows.CreateMutex(nil, false, mutexName)
	if err == nil && handle != 0 {
		// Check if mutex already existed
		if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
			// Another instance owns the mutex - close our handle and forward URL
			windows.CloseHandle(handle)
		} else {
			// We're the first instance - close handle (we'll create it properly in StartIPCServer)
			windows.CloseHandle(handle)
			return false
		}
	} else {
		// Error creating mutex, assume no instance running
		logger.Debug("no running instance found", "error", err)
		return false
	}

	// Another instance exists - send URL via named pipe
	conn, err := winio.DialPipe(IPCPipeName, nil)
	if err != nil {
		logger.Debug("failed to connect to named pipe", "error", err)
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

// StartIPCServer creates the named mutex and pipe server for receiving URLs
func StartIPCServer() error {
	// Create the mutex to signal we're the primary instance
	mutexName, _ := windows.UTF16PtrFromString(IPCMutexName)
	handle, err := windows.CreateMutex(nil, false, mutexName)
	if err != nil {
		return err
	}

	ipcMu.Lock()
	mutexHandle = handle
	ipcMu.Unlock()

	logger.Info("IPC server started", "pipe", IPCPipeName)

	// Start named pipe server in background
	go acceptPipeConnections()

	return nil
}

func acceptPipeConnections() {
	for {
		// Create a new pipe instance for each connection
		// PipeConfig allows setting security and buffer sizes
		cfg := &winio.PipeConfig{
			SecurityDescriptor: "", // Default security (current user only)
			InputBufferSize:    4096,
			OutputBufferSize:   4096,
		}

		listener, err := winio.ListenPipe(IPCPipeName, cfg)
		if err != nil {
			logger.Error("failed to create named pipe", "error", err)
			time.Sleep(1 * time.Second) // Backoff before retry
			continue
		}

		ipcMu.Lock()
		ipcListener = listener
		ipcMu.Unlock()

		// Accept one connection
		conn, err := listener.Accept()
		if err != nil {
			ipcMu.Lock()
			closed := ipcListener == nil
			ipcMu.Unlock()
			if closed {
				listener.Close()
				return
			}
			logger.Warn("pipe accept error", "error", err)
			listener.Close()
			continue
		}

		// Handle connection in goroutine, close listener to allow new pipe creation
		go func(c net.Conn, l net.Listener) {
			handlePipeConnection(c)
			l.Close()
		}(conn, listener)
	}
}

func handlePipeConnection(conn net.Conn) {
	defer conn.Close()

	// Set read deadline
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	reader := bufio.NewReader(conn)
	urlStr, err := reader.ReadString('\n')
	if err != nil {
		logger.Warn("pipe read error", "error", err)
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

func cleanupIPCWindows() {
	ipcMu.Lock()
	defer ipcMu.Unlock()

	// Close pipe listener
	if ipcListener != nil {
		ipcListener.Close()
		ipcListener = nil
	}

	// Release mutex
	if mutexHandle != 0 {
		windows.CloseHandle(mutexHandle)
		mutexHandle = 0
	}

	logger.Debug("IPC cleanup completed")
}

// RegisterURLSchemeHandler on Windows starts the IPC server
func RegisterURLSchemeHandler() {
	if err := StartIPCServer(); err != nil {
		logger.Error("failed to start IPC server", "error", err)
	}
}
