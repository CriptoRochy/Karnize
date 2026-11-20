// Package window shows the Karnize panel in its own window.
//
// The panel is a page served by the running Karnize on 127.0.0.1; this package
// only puts it in a window. It runs as a separate process (`karnize panel`)
// because on macOS both the menu bar icon and a web view want the main thread.
package window

const (
	Width  = 1000
	Height = 720
)
