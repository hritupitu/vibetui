package config

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

//go:embed nvim/init.lua
var nvimInitLua []byte

//go:embed lazygit/config.yml
var lazygitConfigYML []byte

//go:embed welcome.md
var welcomeMD []byte

// Paths contains the resolved config files and assistant metadata used at
// runtime.
type Paths struct {
	// NvimInit is the vibetui-specific Neovim init.lua path.
	NvimInit string
	// LazygitConf is the bundled LazyGit config path written by vibetui.
	LazygitConf string
	// OpencodeTUI is the assistant TUI config path consumed by OpenCode.
	OpencodeTUI string
	// Assistant is the normalized assistant selection identifier.
	Assistant string
	// AssistantCmd is the executable name launched for the assistant pane.
	AssistantCmd string
	// AssistantTitle is the pane title shown for the assistant pane.
	AssistantTitle string
	// SettingsJSON stores the persisted vibetui user settings.
	SettingsJSON string
	// WelcomeMD is the default markdown document opened in docs mode.
	WelcomeMD string
	// UserLazygit is the user's existing LazyGit config, if present.
	UserLazygit string
}

type userSettings struct {
	Assistant string `json:"assistant"`
}

// Setup writes the bundled vibetui configuration and returns the resolved
// runtime paths.
func Setup() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	dir := filepath.Join(home, ".config", "vibetui")

	for _, d := range []string{
		filepath.Join(dir, "lazygit"),
		filepath.Join(home, ".config", "opencode", "themes"),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return Paths{}, err
		}
	}

	p := Paths{
		NvimInit:     filepath.Join(dir, "init.lua"),
		LazygitConf:  filepath.Join(dir, "lazygit", "config.yml"),
		OpencodeTUI:  filepath.Join(dir, "opencode-tui.json"),
		SettingsJSON: filepath.Join(dir, "settings.json"),
		WelcomeMD:    filepath.Join(dir, "welcome.md"),
		UserLazygit:  filepath.Join(home, ".config", "lazygit", "config.yml"),
	}

	vibetuiCherry := filepath.Join(home, ".config", "opencode", "themes", "vibetui-cherry.json")

	if err := os.WriteFile(p.NvimInit, nvimInitLua, 0o644); err != nil {
		return Paths{}, err
	}
	if err := writeIfAbsent(p.SettingsJSON, []byte("{\n  \"assistant\": \"opencode\"\n}\n")); err != nil {
		return Paths{}, err
	}
	if err := writeIfAbsent(p.LazygitConf, lazygitConfigYML); err != nil {
		return Paths{}, err
	}
	if err := os.WriteFile(vibetuiCherry, []byte(`{
  "$schema": "https://opencode.ai/theme.json",
  "theme": {
    "primary": "#ffffff",
    "secondary": "#ffffff",
    "accent": "#ffffff",
    "text": "#ffffff",
    "textMuted": "#ffffff",
    "background": "#1e1e1e",
    "backgroundPanel": "#252526",
    "backgroundElement": "#2d2d30",
    "error": "#ffffff",
    "warning": "#ffffff",
    "success": "#ffffff",
    "info": "#ffffff",
    "border": "#3a3a3d",
    "borderActive": "#b86a94",
    "borderSubtle": "#2a2a2d",
    "diffAdded": "#ffffff",
    "diffRemoved": "#ffffff",
    "diffContext": "#ffffff",
    "diffHunkHeader": "#ffffff",
    "diffHighlightAdded": "#c0fff1",
    "diffHighlightRemoved": "#ffaacc",
    "diffAddedBg": "#1a2e2a",
    "diffRemovedBg": "#3d1228",
    "diffContextBg": "#1e1e1e",
    "diffLineNumber": "#ffffff",
    "diffAddedLineNumberBg": "#1a2e2a",
    "diffRemovedLineNumberBg": "#3d1228",
    "markdownText": "#ffffff",
    "markdownHeading": "#ffffff",
    "markdownLink": "#ffffff",
    "markdownLinkText": "#ffffff",
    "markdownCode": "#ffffff",
    "markdownBlockQuote": "#ffffff",
    "markdownEmph": "#ffffff",
    "markdownStrong": "#fff5fa",
    "markdownHorizontalRule": "#3a3a3d",
    "markdownListItem": "#ffffff",
    "markdownListEnumeration": "#ffffff",
    "markdownImage": "#ffffff",
    "markdownImageText": "#ffffff",
    "markdownCodeBlock": "#ffffff",
    "syntaxComment": "#ffffff",
    "syntaxKeyword": "#ffffff",
    "syntaxFunction": "#ffffff",
    "syntaxVariable": "#ffffff",
    "syntaxString": "#ffffff",
    "syntaxNumber": "#ffffff",
    "syntaxType": "#ffffff",
    "syntaxOperator": "#ffffff",
    "syntaxPunctuation": "#ffffff"
  }
}
`), 0o644); err != nil {
		return Paths{}, err
	}
	if err := os.WriteFile(p.OpencodeTUI, []byte("{\n  \"theme\": \"vibetui-cherry\"\n}\n"), 0o644); err != nil {
		return Paths{}, err
	}
	if err := os.WriteFile(p.WelcomeMD, welcomeMD, 0o644); err != nil {
		return Paths{}, err
	}

	p.Assistant, p.AssistantCmd, p.AssistantTitle = resolveAssistantSelection(p.SettingsJSON)

	return p, nil
}

func resolveAssistantSelection(settingsPath string) (string, string, string) {
	assistant := "opencode"

	if b, err := os.ReadFile(settingsPath); err == nil {
		var cfg userSettings
		if jsonErr := json.Unmarshal(b, &cfg); jsonErr == nil {
			value := strings.TrimSpace(cfg.Assistant)
			if value != "" {
				assistant = value
			}
		}
	}

	if envValue := strings.TrimSpace(os.Getenv("VIBETUI_ASSISTANT")); envValue != "" {
		assistant = envValue
	}

	return AssistantProfile(assistant)
}

// AssistantProfile normalizes an assistant selection and returns the stored
// assistant key, executable name, and pane title.
func AssistantProfile(value string) (string, string, string) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "claude", "claude-code", "claude code":
		return "claude", "claude", "Claude"
	case "opencode", "open-code", "open code":
		fallthrough
	default:
		return "opencode", "opencode", "OpenCode"
	}
}

func writeIfAbsent(path string, data []byte) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.WriteFile(path, data, 0o644)
	}
	return nil
}
