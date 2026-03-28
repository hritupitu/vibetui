package ui

import "github.com/charmbracelet/lipgloss"

var (
	BorderFocused   = lipgloss.ThickBorder()
	BorderUnfocused = lipgloss.RoundedBorder()
)

func PaneBorder(focused bool) lipgloss.Border {
	if focused {
		return BorderFocused
	}
	return BorderUnfocused
}

func PaneTitleStyle(title string, focused bool) string {
	if focused {
		return lipgloss.NewStyle().Bold(true).Render(" " + title + " ")
	}
	return lipgloss.NewStyle().Faint(true).Render(" " + title + " ")
}

var TabBarStyle = lipgloss.NewStyle()

var ActiveTabStyle = lipgloss.NewStyle().
	Bold(true).
	Underline(true).
	Padding(0, 2)

var InactiveTabStyle = lipgloss.NewStyle().
	Faint(true).
	Padding(0, 2)

var StatusBarStyle = lipgloss.NewStyle()

var StatusButtonActiveStyle = lipgloss.NewStyle().
	Bold(true).
	Padding(0, 1)

var StatusButtonStyle = lipgloss.NewStyle().
	Faint(true).
	Padding(0, 1)

var ConfirmDialogStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(1, 4)

var ConfirmActiveBtn = lipgloss.NewStyle().
	Reverse(true).
	Padding(0, 3).
	Margin(0, 1).
	Bold(true)

var ConfirmInactiveBtn = lipgloss.NewStyle().
	Faint(true).
	Padding(0, 3).
	Margin(0, 1)
