package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type StatusButton struct {
	Label  string
	PaneID string
	StartX int
	EndX   int
}

type StatusBar struct {
	Buttons    []StatusButton
	ActivePane string
	openMDBtn  StatusButton
}

func NewStatusBar(paneIDs, labels []string) StatusBar {
	btns := make([]StatusButton, len(paneIDs))
	for i := range paneIDs {
		btns[i] = StatusButton{Label: labels[i], PaneID: paneIDs[i]}
	}
	return StatusBar{Buttons: btns}
}

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
