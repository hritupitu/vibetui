package pane

// OutputMsg is sent to the Bubble Tea runtime whenever a pane produces
// new output.  The app re-renders on every such message.
type OutputMsg struct {
	// PaneID identifies which pane produced output.
	PaneID string
}

// ExitMsg describes a pane subprocess termination.
type ExitMsg struct {
	// PaneID identifies the pane whose subprocess exited.
	PaneID string
	// Err is the termination error reported by the subprocess, if any.
	Err error
}
