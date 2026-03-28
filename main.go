package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/hritupitu/vibetui/internal/app"
	"github.com/hritupitu/vibetui/internal/config"
)

func main() {
	cfg, err := config.Setup()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config setup:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		app.New(cfg),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
