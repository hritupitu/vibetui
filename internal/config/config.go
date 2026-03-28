package config

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed nvim/init.lua
var nvimInitLua []byte

//go:embed lazygit/config.yml
var lazygitConfigYML []byte

//go:embed welcome.md
var welcomeMD []byte

type Paths struct {
	NvimInit    string
	LazygitConf string
	OpencodeTUI string
	WelcomeMD   string
	UserLazygit string
}

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
		NvimInit:    filepath.Join(dir, "init.lua"),
		LazygitConf: filepath.Join(dir, "lazygit", "config.yml"),
		OpencodeTUI: filepath.Join(dir, "opencode-tui.json"),
		WelcomeMD:   filepath.Join(dir, "welcome.md"),
		UserLazygit: filepath.Join(home, ".config", "lazygit", "config.yml"),
	}

	vibetuiCherry := filepath.Join(home, ".config", "opencode", "themes", "vibetui-cherry.json")

	if err := os.WriteFile(p.NvimInit, nvimInitLua, 0o644); err != nil {
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

	return p, nil
}

func writeIfAbsent(path string, data []byte) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.WriteFile(path, data, 0o644)
	}
	return nil
}
