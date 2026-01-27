//go:build darwin && cgo

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

// Declaration of the function implemented in url_handler_darwin.m
void RegisterURLHandler(void);
*/
import "C"

func init() {
	// macOS doesn't need TrySendToRunningInstance - Apple Events handles the
	// "already running" case natively via kAEGetURL
	TrySendToRunningInstance = func(url string) bool { return false }
	CleanupIPC = func() {}
}

//export urlCallbackGo
func urlCallbackGo(urlCString *C.char) {
	urlStr := C.GoString(urlCString)
	server.HandleURLScheme(urlStr)
}

// RegisterURLSchemeHandler registers the app to receive agenthub:// URLs
// while already running via Apple Events (kAEGetURL)
func RegisterURLSchemeHandler() {
	C.RegisterURLHandler()
}
