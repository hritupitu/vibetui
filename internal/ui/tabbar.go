package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Tab represents a clickable tab in the tab bar.
type Tab struct {
	Label  string
	StartX int
	EndX   int
}

// TabBar renders the top tab bar and tracks tab x-positions for click detection.
type TabBar struct {
	Tabs       []Tab
	ActiveIdx  int
	totalWidth int
}

// NewTabBar creates a TabBar with the given labels.
func NewTabBar(labels []string) TabBar {
	tabs := make([]Tab, len(labels))
	for i, l := range labels {
		tabs[i] = Tab{Label: l}
	}
	return TabBar{Tabs: tabs}
}

// Render draws the tab bar and records each tab's x-range for hit testing.
// Returns a string of exactly width characters.
func (tb *TabBar) Render(width, activeIdx int) string {
	tb.ActiveIdx = activeIdx
	tb.totalWidth = width

	var sb strings.Builder
	x := 0

	for i := range tb.Tabs {
		var label string
		var rendered string
		if i == activeIdx {
			label = " " + tb.Tabs[i].Label + " "
			rendered = ActiveTabStyle.Render(label)
		} else {
			label = " " + tb.Tabs[i].Label + " "
			rendered = InactiveTabStyle.Render(label)
		}

		// Record visual position (label is 2+len chars wide due to padding)
		tabWidth := lipgloss.Width(rendered)
		tb.Tabs[i].StartX = x
		tb.Tabs[i].EndX = x + tabWidth - 1
		x += tabWidth

		sb.WriteString(rendered)
	}

	// Fill remaining space with tab bar background
	remaining := width - x
	if remaining > 0 {
		filler := TabBarStyle.Render(strings.Repeat(" ", remaining))
		sb.WriteString(filler)
	}

	return sb.String()
}

// HitTest returns the tab index at the given x position, or -1 if none.
func (tb *TabBar) HitTest(x int) int {
	for i, t := range tb.Tabs {
		if x >= t.StartX && x <= t.EndX {
			return i
		}
	}
	return -1
}
