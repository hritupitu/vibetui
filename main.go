package main

import (
	"fmt"
	"os"

	"github.com/hritupitu/vibetui/internal/config"
	"github.com/hritupitu/vibetui/internal/tmuxsession"
)

func main() {
	cfg, err := config.Setup()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config setup:", err)
		os.Exit(1)
	}
	if err := tmuxsession.Launch(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
