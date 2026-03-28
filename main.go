package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hritupitu/vibetui/internal/config"
	"github.com/hritupitu/vibetui/internal/tmuxsession"
)

func main() {
	if len(os.Args) > 1 {
		arg := strings.ToLower(strings.TrimSpace(os.Args[1]))
		if arg == "-h" || arg == "--help" || arg == "help" {
			printUsage()
			return
		}
	}

	cfg, err := config.Setup()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config setup:", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		arg := strings.ToLower(strings.TrimSpace(os.Args[1]))
		switch arg {
		case "opencode", "claude":
			cfg.Assistant, cfg.AssistantCmd, cfg.AssistantTitle = config.AssistantProfile(arg)
		case "-h", "--help", "help":
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown assistant mode %q\n\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}

	if err := tmuxsession.Launch(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  vibetui")
	fmt.Println("  vibetui opencode")
	fmt.Println("  vibetui claude")
	fmt.Println("")
	fmt.Println("Priority for assistant selection:")
	fmt.Println("  1) CLI argument (opencode/claude)")
	fmt.Println("  2) VIBETUI_ASSISTANT environment variable")
	fmt.Println("  3) ~/.config/vibetui/settings.json")
}
