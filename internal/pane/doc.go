// Package pane manages subprocess-backed terminal panes.
//
// Each pane runs inside a PTY, parses terminal output through vt10x, and
// exposes rendering and input forwarding helpers to the Bubble Tea app.
package pane
