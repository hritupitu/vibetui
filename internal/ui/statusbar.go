package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusButton describes a clickable status bar item and the pane it targets.
type StatusButton struct {
	// Label is the button caption.
	Label string
	// PaneID is the pane identifier returned when the button is clicked.
	PaneID string
	// StartX is the inclusive left x-coordinate used for hit testing.
	StartX int
	// EndX is the inclusive right x-coordinate used for hit testing.
	EndX int
}

// StatusBar renders the bottom status row and tracks click hit boxes.
type StatusBar struct {
	// Buttons is the ordered list of pane buttons displayed in the bar.
	Buttons []StatusButton
	// ActivePane is the pane ID currently rendered as active.
	ActivePane string
	openMDBtn  StatusButton
}

// NewStatusBar creates a status bar with one button per pane ID / label pair.
func NewStatusBar(paneIDs, labels []string) StatusBar {
	btns := make([]StatusButton, len(paneIDs))
	for i := range paneIDs {
		btns[i] = StatusButton{Label: labels[i], PaneID: paneIDs[i]}
	}
	return StatusBar{Buttons: btns}
}

// Render draws the status bar and updates button hit boxes for click handling.
func (sb *StatusBar) Render(width int, activePaneID string, docsView bool) string {
	sb.ActivePane = activePaneID

	var left strings.Builder
	x := 0

	for i := range sb.Buttons {
		icon := "○ "
		var rendered string
		if sb.Buttons[i].PaneID == activePaneID {
			icon = "● "
			rendered = StatusButtonActiveStyle.Render(icon + sb.Buttons[i].Label)
		} else {
			rendered = StatusButtonStyle.Render(icon + sb.Buttons[i].Label)
		}
		w := lipgloss.Width(rendered)
		sb.Buttons[i].StartX = x
		sb.Buttons[i].EndX = x + w - 1
		x += w
		left.WriteString(rendered)
	}

	if docsView {
		rendered := StatusButtonStyle.Render("⊕ Open MD")
		w := lipgloss.Width(rendered)
		sb.openMDBtn = StatusButton{
			Label:  "Open MD",
			PaneID: "openmd",
			StartX: x,
			EndX:   x + w - 1,
		}
		x += w
		left.WriteString(rendered)
	}

	hint := StatusButtonStyle.Render("Ctrl+\\ switch · Ctrl+C quit")
	hintW := lipgloss.Width(hint)
	leftStr := left.String()
	leftW := lipgloss.Width(leftStr)
	gap := width - leftW - hintW
	if gap < 0 {
		gap = 0
	}
	filler := StatusBarStyle.Render(strings.Repeat(" ", gap))

	return leftStr + filler + hint
}

// HitTest returns the pane ID at x, or an empty string when no button matches.
func (sb *StatusBar) HitTest(x int) string {
	for _, b := range sb.Buttons {
		if x >= b.StartX && x <= b.EndX {
			return b.PaneID
		}
	}
	if x >= sb.openMDBtn.StartX && x <= sb.openMDBtn.EndX && sb.openMDBtn.PaneID != "" {
		return sb.openMDBtn.PaneID
	}
	return ""
}
