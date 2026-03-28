package pane

// OutputMsg is sent to the Bubble Tea runtime whenever a pane produces
// new output.  The app re-renders on every such message.
type OutputMsg struct {
	PaneID string
}

// ExitMsg is sent when a managed subprocess terminates.
type ExitMsg struct {
	PaneID string
	Err    error
}
