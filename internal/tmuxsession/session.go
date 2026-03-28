package tmuxsession

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hritupitu/vibetui/internal/config"
)

func Launch(cfg config.Paths) error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("tmux is required: %w", err)
	}

	session := fmt.Sprintf("vibetui-%d", time.Now().Unix())
	window := session + ":vibetui"
	bottomTarget := window + ".1"
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	lazygitCmd := shellWrap(lazygitCommand(cfg))
	terminalCmd := shellWrap(shell)
	nvimCmd := shellWrap(fmt.Sprintf("env NVIM_APPNAME=vibetui nvim -u %s", shellQuote(cfg.NvimInit)))
	claudeCmd := shellWrap("claude")

	commands := [][]string{
		{"new-session", "-d", "-s", session, "-n", "vibetui", nvimCmd},
		{"split-window", "-t", window + ".0", "-h", "-p", "40", claudeCmd},
		{"select-pane", "-t", window + ".0"},
		{"split-window", "-t", window + ".0", "-v", "-p", "25", lazygitCmd},
		{"select-pane", "-t", window + ".0"},
		{"set-option", "-t", session, "-g", "mouse", "on"},
		{"set-option", "-t", session, "-g", "default-terminal", "tmux-256color"},
		{"set-option", "-t", session, "-ag", "terminal-overrides", ",*:RGB"},
		{"set-option", "-t", session, "-g", "status", "on"},
		{"set-option", "-t", session, "-g", "status-position", "top"},
		{"set-option", "-t", session, "-g", "status-justify", "left"},
		{"set-option", "-t", session, "-g", "status-left", "[Git] [Terminal]"},
		{"set-option", "-t", session, "-g", "status-left-length", "24"},
		{"set-option", "-t", session, "-g", "status-right", "Ctrl+C Exit"},
		{"set-option", "-t", session, "-g", "status-right-length", "20"},
		{"set-option", "-t", session, "-g", "window-status-format", ""},
		{"set-option", "-t", session, "-g", "window-status-current-format", ""},
		{"set-option", "-t", session, "-g", "allow-rename", "off"},
		{"set-window-option", "-t", window, "pane-border-status", "top"},
		{"set-window-option", "-t", window, "pane-border-format", "#{pane_title}"},
		{"select-pane", "-t", window + ".0", "-T", "LazyVim"},
		{"select-pane", "-t", window + ".2", "-T", "Claude"},
		{"select-pane", "-t", bottomTarget, "-T", "Git"},
		{"bind-key", "-n", "F1", "run-shell", swapCommand(bottomTarget, lazygitCmd, "Git")},
		{"bind-key", "-n", "F2", "run-shell", swapCommand(bottomTarget, terminalCmd, "Terminal")},
		{"bind-key", "-n", "C-c", "confirm-before", "-p", "Exit vibetui? (y/n)", "kill-session -t " + session},
		{"bind-key", "-n", "MouseDown1StatusLeft", "run-shell", statusClickScript(bottomTarget, lazygitCmd, terminalCmd)},
	}

	for _, args := range commands {
		if err := runTmux(args...); err != nil {
			_ = runTmux("kill-session", "-t", session)
			return err
		}
	}

	if os.Getenv("VIBETUI_NO_ATTACH") == "1" {
		return nil
	}
	if os.Getenv("TMUX") != "" {
		return runTmux("switch-client", "-t", session)
	}

	attach := exec.Command("tmux", "attach-session", "-t", session)
	attach.Stdin = os.Stdin
	attach.Stdout = os.Stdout
	attach.Stderr = os.Stderr
	return attach.Run()
}

func lazygitCommand(cfg config.Paths) string {
	configFile := cfg.LazygitConf
	if _, err := os.Stat(cfg.UserLazygit); err == nil {
		configFile = cfg.UserLazygit + "," + cfg.LazygitConf
	}
	return fmt.Sprintf("lazygit --use-config-file %s", shellQuote(configFile))
}

func statusClickScript(bottomTarget, lazygitCmd, terminalCmd string) string {
	return fmt.Sprintf(`x=#{mouse_x}; if [ "$x" -ge 0 ] && [ "$x" -le 4 ]; then %s; elif [ "$x" -ge 6 ] && [ "$x" -le 15 ]; then %s; else :; fi`, swapCommand(bottomTarget, lazygitCmd, "Git"), swapCommand(bottomTarget, terminalCmd, "Terminal"))
}

func swapCommand(target, cmd, title string) string {
	return fmt.Sprintf("tmux respawn-pane -k -t %s %s && tmux select-pane -t %s -T %s", target, shellQuote(cmd), target, shellQuote(title))
}

func runTmux(args ...string) error {
	cmd := exec.Command("tmux", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func docsCommand(welcomePath, shell string) string {
	if _, err := exec.LookPath("melo"); err == nil {
		return fmt.Sprintf("melo %s", shellQuote(welcomePath))
	}
	return fmt.Sprintf("printf %s; exec %s", shellQuote("melo not installed\n\ncargo install --path .\nhttps://github.com/mw2000/melo\n\nPress Ctrl+C to quit.\n"), shellQuote(shell))
}

func shellWrap(cmd string) string {
	return "sh -lc " + shellQuote(cmd)
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
