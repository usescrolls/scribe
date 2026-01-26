package main

// URL scheme IPC interface
//
// On Linux/Windows, when the app is already running and a new instance is launched
// with an agenthub:// URL, the new instance should forward the URL to the running
// instance and exit. This is different from macOS which handles this via Apple Events.
//
// Each platform sets these function pointers in their init() functions.

// TrySendToRunningInstance attempts to forward a URL to an already-running instance.
// Returns true if URL was sent successfully (caller should exit).
// Returns false if no instance is running (caller should continue startup).
// This function is called BEFORE the server is fully initialized.
var TrySendToRunningInstance func(url string) bool

// CleanupIPC performs cleanup of IPC resources (socket/mutex) on shutdown.
// Called from onExit() to ensure clean shutdown.
var CleanupIPC func()

// IPC constants
const (
	// IPCSocketName is the Unix socket filename for Linux IPC
	IPCSocketName = "ipc.sock"

	// IPCMutexName is the named mutex for Windows single-instance detection
	IPCMutexName = `Global\Scribe`

	// IPCPipeName is the named pipe for Windows IPC
	IPCPipeName = `\\.\pipe\Scribe`
)
