package ui

import "github.com/charmbracelet/lipgloss"

var (
	// BorderFocused is used for the currently focused pane.
	BorderFocused = lipgloss.ThickBorder()
	// BorderUnfocused is used for panes that are visible but not focused.
	BorderUnfocused = lipgloss.RoundedBorder()
)

// PaneBorder selects the border style for a pane based on focus state.
func PaneBorder(focused bool) lipgloss.Border {
	if focused {
		return BorderFocused
	}
	return BorderUnfocused
}

// PaneTitleStyle styles a pane title consistently with the pane focus state.
func PaneTitleStyle(title string, focused bool) string {
	if focused {
		return lipgloss.NewStyle().Bold(true).Render(" " + title + " ")
	}
	return lipgloss.NewStyle().Faint(true).Render(" " + title + " ")
}

// TabBarStyle styles the background filler for the top tab row.
var TabBarStyle = lipgloss.NewStyle()

// ActiveTabStyle styles the selected tab.
var ActiveTabStyle = lipgloss.NewStyle().
	Bold(true).
	Underline(true).
	Padding(0, 2)

// InactiveTabStyle styles tabs that are visible but not selected.
var InactiveTabStyle = lipgloss.NewStyle().
	Faint(true).
	Padding(0, 2)

// StatusBarStyle styles the spacer area between status bar controls and hints.
var StatusBarStyle = lipgloss.NewStyle()

// StatusButtonActiveStyle styles the active pane indicator in the status bar.
var StatusButtonActiveStyle = lipgloss.NewStyle().
	Bold(true).
	Padding(0, 1)

// StatusButtonStyle styles inactive status bar buttons.
var StatusButtonStyle = lipgloss.NewStyle().
	Faint(true).
	Padding(0, 1)

// ConfirmDialogStyle styles centered modal dialogs such as quit confirmation.
var ConfirmDialogStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(1, 4)

// ConfirmActiveBtn styles the selected button inside confirmation dialogs.
var ConfirmActiveBtn = lipgloss.NewStyle().
	Reverse(true).
	Padding(0, 3).
	Margin(0, 1).
	Bold(true)

// ConfirmInactiveBtn styles unselected confirmation dialog buttons.
var ConfirmInactiveBtn = lipgloss.NewStyle().
	Faint(true).
	Padding(0, 3).
	Margin(0, 1)
