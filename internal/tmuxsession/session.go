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
	meloCmd := shellWrap(docsCommand(cfg.WelcomeMD, shell))
	nvimCmd := shellWrap(fmt.Sprintf("env NVIM_APPNAME=vibetui nvim -u %s -c %s", shellQuote(cfg.NvimInit), shellQuote("set notermguicolors background=dark")))
	claudeCmd := shellWrap("claude")

	commands := [][]string{
		{"new-session", "-d", "-s", session, "-n", "vibetui", nvimCmd},
		{"split-window", "-t", window + ".0", "-h", "-p", "40", claudeCmd},
		{"select-pane", "-t", window + ".0"},
		{"split-window", "-t", window + ".0", "-v", "-p", "25", lazygitCmd},
		{"select-pane", "-t", window + ".0"},
		{"set-option", "-t", session, "-g", "mouse", "on"},
		{"set-option", "-t", session, "-g", "status", "on"},
		{"set-option", "-t", session, "-g", "status-position", "top"},
		{"set-option", "-t", session, "-g", "status-justify", "left"},
		{"set-option", "-t", session, "-g", "status-left", "[Git] [Terminal] [Melo]"},
		{"set-option", "-t", session, "-g", "status-left-length", "40"},
		{"set-option", "-t", session, "-g", "status-right", "Ctrl+C Exit"},
		{"set-option", "-t", session, "-g", "status-right-length", "20"},
		{"set-option", "-t", session, "-g", "window-status-format", ""},
		{"set-option", "-t", session, "-g", "window-status-current-format", ""},
		{"set-option", "-t", session, "-g", "allow-rename", "off"},
		{"set-window-option", "-t", window, "pane-border-status", "top"},
		{"bind-key", "-n", "F1", "respawn-pane", "-k", "-t", bottomTarget, lazygitCmd},
		{"bind-key", "-n", "F2", "respawn-pane", "-k", "-t", bottomTarget, terminalCmd},
		{"bind-key", "-n", "F3", "respawn-pane", "-k", "-t", bottomTarget, meloCmd},
		{"bind-key", "-n", "C-c", "confirm-before", "-p", "Exit vibetui? (y/n)", "kill-session -t " + session},
		{"bind-key", "-n", "MouseDown1StatusLeft", "run-shell", statusClickScript(session, bottomTarget, lazygitCmd, terminalCmd, meloCmd)},
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

func statusClickScript(session, bottomTarget, lazygitCmd, terminalCmd, meloCmd string) string {
	return fmt.Sprintf(`x=#{mouse_x}; if [ "$x" -ge 0 ] && [ "$x" -le 4 ]; then tmux respawn-pane -k -t %s %s; elif [ "$x" -ge 6 ] && [ "$x" -le 15 ]; then tmux respawn-pane -k -t %s %s; elif [ "$x" -ge 17 ] && [ "$x" -le 22 ]; then tmux respawn-pane -k -t %s %s; else :; fi`, bottomTarget, shellQuote(lazygitCmd), bottomTarget, shellQuote(terminalCmd), bottomTarget, shellQuote(meloCmd))
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
