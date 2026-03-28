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
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	lazygitConfig := cfg.LazygitConf
	if _, err := os.Stat(cfg.UserLazygit); err == nil {
		lazygitConfig = cfg.UserLazygit + "," + cfg.LazygitConf
	}

	nvimCmd := shellWrap(fmt.Sprintf("env NVIM_APPNAME=vibetui nvim -u %s -c %s", shellQuote(cfg.NvimInit), shellQuote("set notermguicolors background=dark")))
	claudeCmd := shellWrap("claude")
	lazygitCmd := shellWrap(fmt.Sprintf("lazygit --use-config-file %s", shellQuote(lazygitConfig)))
	terminalCmd := shellWrap(shell)
	docsCmd := shellWrap(docsCommand(cfg.WelcomeMD, shell))

	commands := [][]string{
		{"new-session", "-d", "-s", session, "-n", "Editor", nvimCmd},
		{"set-option", "-t", session, "-g", "mouse", "on"},
		{"set-option", "-t", session, "-g", "status", "on"},
		{"set-option", "-t", session, "-g", "status-position", "top"},
		{"set-option", "-t", session, "-g", "allow-rename", "off"},
		{"set-window-option", "-t", session + ":Editor", "pane-border-status", "top"},
		{"split-window", "-t", session + ":Editor.0", "-v", "-p", "25", lazygitCmd},
		{"select-pane", "-t", session + ":Editor.0"},
		{"split-window", "-t", session + ":Editor.0", "-h", "-p", "40", claudeCmd},
		{"select-pane", "-t", session + ":Editor.0"},
		{"new-window", "-t", session, "-n", "Terminal", nvimCmd},
		{"set-window-option", "-t", session + ":Terminal", "pane-border-status", "top"},
		{"split-window", "-t", session + ":Terminal.0", "-v", "-p", "25", terminalCmd},
		{"select-pane", "-t", session + ":Terminal.0"},
		{"split-window", "-t", session + ":Terminal.0", "-h", "-p", "40", claudeCmd},
		{"select-pane", "-t", session + ":Terminal.0"},
		{"new-window", "-t", session, "-n", "Git", lazygitCmd},
		{"set-window-option", "-t", session + ":Git", "pane-border-status", "top"},
		{"new-window", "-t", session, "-n", "Docs", nvimCmd},
		{"set-window-option", "-t", session + ":Docs", "pane-border-status", "top"},
		{"split-window", "-t", session + ":Docs.0", "-v", "-p", "25", lazygitCmd},
		{"select-pane", "-t", session + ":Docs.0"},
		{"split-window", "-t", session + ":Docs.0", "-h", "-p", "40", docsCmd},
		{"select-pane", "-t", session + ":Docs.0"},
		{"select-window", "-t", session + ":Editor"},
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
	msg := fmt.Sprintf("printf %s; exec %s", shellQuote("melo not installed\n\ncargo install --path .\nhttps://github.com/mw2000/melo\n\nPress Ctrl+C to quit.\n"), shellQuote(shell))
	return msg
}

func shellWrap(cmd string) string {
	return "sh -lc " + shellQuote(cmd)
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
