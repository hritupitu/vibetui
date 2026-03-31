package app

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/hritupitu/vibetui/internal/config"
	"github.com/hritupitu/vibetui/internal/pane"
	"github.com/hritupitu/vibetui/internal/ui"
)

type appState int

const (
	stateNormal appState = iota
	stateConfirmQuit
	stateOpenMD
)

// ViewType selects which bottom-area layout the app renders.
type ViewType int

const (
	// ViewIDE shows the editor, assistant, and git panes together.
	ViewIDE ViewType = iota
	// ViewTerminal swaps the bottom git pane for the shell pane.
	ViewTerminal
	// ViewGit expands the git pane to fill the content area.
	ViewGit
	// ViewDocs shows the markdown preview pane beside the editor.
	ViewDocs
)

// FocusType identifies which pane receives keyboard and mouse input.
type FocusType int

const (
	// FocusLazyVim targets the editor pane.
	FocusLazyVim FocusType = iota
	// FocusOpenCode targets the assistant pane.
	FocusOpenCode
	// FocusLazyGit targets the git pane.
	FocusLazyGit
	// FocusTerminal targets the shell pane.
	FocusTerminal
	// FocusMarkdown targets the markdown preview pane.
	FocusMarkdown
)

const (
	tabBarHeight    = 1
	statusBarHeight = 1
	lazyGitPct      = 25
	lazyVimPct      = 60
)

type layout struct {
	lvW, lvH int
	ocW, ocH int
	lgW, lgH int

	lvIW, lvIH int
	ocIW, ocIH int
	lgIW, lgIH int
}

// Model is vibetui's Bubble Tea model.
//
// It tracks layout, pane lifecycle, focused pane, and modal UI state while the
// individual pane processes are managed by the pane package.
type Model struct {
	width, height int
	ready         bool

	state      appState
	confirmYes bool
	mdInput    textinput.Model

	view  ViewType
	focus FocusType

	cfg config.Paths

	lazyVim  *pane.Pane
	openCode *pane.Pane
	lazyGit  *pane.Pane
	terminal *pane.Pane
	melo     *pane.Pane

	meloFile string

	outputCh chan pane.OutputMsg

	tabBar    ui.TabBar
	statusBar ui.StatusBar

	lay layout
}

// New builds a Model configured to launch panes from cfg.
func New(cfg config.Paths) Model {
	ti := textinput.New()
	ti.Placeholder = "path/to/file.md"
	ti.CharLimit = 256

	return Model{
		view:     ViewIDE,
		focus:    FocusLazyVim,
		cfg:      cfg,
		meloFile: cfg.WelcomeMD,
		mdInput:  ti,
		outputCh: make(chan pane.OutputMsg, 64),
		tabBar:   ui.NewTabBar([]string{"  Editor", "  Terminal", "  Git", "  Docs"}),
		statusBar: ui.NewStatusBar(
			[]string{"lazyvim", pane.ClaudePaneID, "lazygit", "terminal", "markdown"},
			[]string{"Editor", "Claude", "Git", "Terminal", "Docs"},
		),
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) currentPane() *pane.Pane {
	switch m.focus {
	case FocusLazyVim:
		return m.lazyVim
	case FocusOpenCode:
		return m.openCode
	case FocusLazyGit:
		return m.lazyGit
	case FocusTerminal:
		return m.terminal
	case FocusMarkdown:
		return m.melo
	}
	return nil
}

func (m *Model) currentPaneID() string {
	p := m.currentPane()
	if p == nil {
		return ""
	}
	return p.ID
}

func focusIDToType(id string) (FocusType, bool) {
	switch id {
	case "lazyvim":
		return FocusLazyVim, true
	case pane.ClaudePaneID:
		return FocusOpenCode, true
	case "lazygit":
		return FocusLazyGit, true
	case "terminal":
		return FocusTerminal, true
	case "markdown":
		return FocusMarkdown, true
	}
	return 0, false
}

func (m *Model) computeLayout() layout {
	var l layout

	avail := m.height - tabBarHeight - statusBarHeight

	l.lgH = avail * lazyGitPct / 100
	mainH := avail - l.lgH
	l.lgW = m.width

	l.lvW = m.width * lazyVimPct / 100
	l.lvH = mainH

	l.ocW = m.width - l.lvW
	l.ocH = mainH

	l.lvIW = max1(l.lvW-2, 1)
	l.lvIH = max1(l.lvH-2, 1)
	l.ocIW = max1(l.ocW-2, 1)
	l.ocIH = max1(l.ocH-2, 1)
	l.lgIW = max1(l.lgW-2, 1)
	l.lgIH = max1(l.lgH-2, 1)

	return l
}

func (m *Model) closeAllPanes() {
	for _, p := range []*pane.Pane{m.lazyVim, m.openCode, m.lazyGit, m.terminal, m.melo} {
		if p != nil {
			p.Close()
		}
	}
}

func max1(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func waitForOutput(ch <-chan pane.OutputMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}
