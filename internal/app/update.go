package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/bubbletea"
	"github.com/hritupitu/vibetui/internal/pane"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.lay = m.computeLayout()

		if !m.ready {
			m.ready = true
			if err := m.startPanes(); err != nil {
				return m, tea.Quit
			}
			return m, waitForOutput(m.outputCh)
		}
		m.resizePanes()
		return m, nil

	case pane.OutputMsg:
		return m, waitForOutput(m.outputCh)

	case tea.KeyMsg:
		if m.state == stateConfirmQuit {
			switch msg.String() {
			case "left", "right", "tab", "h", "l":
				m.confirmYes = !m.confirmYes
			case "enter":
				if m.confirmYes {
					m.closeAllPanes()
					return m, tea.Quit
				}
				m.state = stateNormal
			case "y":
				m.closeAllPanes()
				return m, tea.Quit
			case "n", "esc", "ctrl+c":
				m.state = stateNormal
			}
			return m, nil
		}

		if m.state == stateOpenMD {
			switch msg.String() {
			case "enter":
				if path := m.mdInput.Value(); path != "" {
					m.openMDFile(path)
				}
				m.mdInput.SetValue("")
				m.state = stateNormal
			case "esc", "ctrl+c":
				m.mdInput.SetValue("")
				m.state = stateNormal
			default:
				var cmd tea.Cmd
				m.mdInput, cmd = m.mdInput.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			m.state = stateConfirmQuit
			m.confirmYes = false
			return m, nil
		case "ctrl+o":
			if m.view == ViewDocs {
				m.state = stateOpenMD
				m.mdInput.Focus()
				return m, nil
			}
		case "ctrl+\\":
			m.cycleFocus()
			return m, nil
		case "ctrl+w":
			m.cycleFocus()
			return m, nil
		}

		if p := m.currentPane(); p != nil {
			data := keyToBytes(msg)
			if len(data) > 0 {
				_ = p.Write(data)
			}
		}
		return m, nil

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && m.handleChromeClick(msg.X, msg.Y) {
			return m, nil
		}
		m.handlePaneMouse(msg)
		return m, nil
	}

	return m, nil
}

func (m *Model) startPanes() error {
	l := m.lay
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	m.lazyVim = pane.New("lazyvim", " Editor", "nvim", "-u", m.cfg.NvimInit, "-c", "set notermguicolors background=dark").
		WithEnv("NVIM_APPNAME=vibetui")
	if err := m.lazyVim.Start(l.lvIW, l.lvIH, m.outputCh); err != nil {
		return err
	}

	m.openCode = pane.New("opencode", " AI", "opencode").
		WithEnv("OPENCODE_TUI_CONFIG=" + m.cfg.OpencodeTUI)
	if err := m.openCode.Start(l.ocIW, l.ocIH, m.outputCh); err != nil {
		return err
	}

	lgArgs := lazygitArgs(m.cfg.UserLazygit, m.cfg.LazygitConf)
	m.lazyGit = pane.New("lazygit", " Git", "lazygit", lgArgs...)
	if err := m.lazyGit.Start(l.lgIW, l.lgIH, m.outputCh); err != nil {
		return err
	}

	m.terminal = pane.New("terminal", " Terminal", shell)
	if err := m.terminal.Start(l.lgIW, l.lgIH, m.outputCh); err != nil {
		return err
	}

	if meloAvailable() {
		m.melo = pane.New("markdown", " Docs", "melo", m.meloFile)
		_ = m.melo.Start(l.ocIW, l.ocIH, m.outputCh)
	}

	return nil
}

func meloAvailable() bool {
	_, err := exec.LookPath("melo")
	return err == nil
}

func lazygitArgs(userConf, vibetuiConf string) []string {
	merged := vibetuiConf
	if _, err := os.Stat(userConf); err == nil {
		merged = userConf + "," + vibetuiConf
	}
	return []string{"--use-config-file", merged}
}

func (m *Model) openMDFile(path string) {
	if !meloAvailable() {
		return
	}
	if !filepath.IsAbs(path) {
		if cwd, err := os.Getwd(); err == nil {
			path = filepath.Join(cwd, path)
		}
	}
	m.meloFile = path

	l := m.lay
	if m.melo != nil {
		m.melo.Close()
	}
	m.melo = pane.New("markdown", " Docs", "melo", path)
	_ = m.melo.Start(l.ocIW, l.ocIH, m.outputCh)

	m.view = ViewDocs
	m.focus = FocusMarkdown
}

func (m *Model) resizePanes() {
	l := m.lay
	if m.lazyVim != nil {
		m.lazyVim.Resize(l.lvIW, l.lvIH)
	}
	if m.openCode != nil {
		m.openCode.Resize(l.ocIW, l.ocIH)
	}
	if m.lazyGit != nil {
		m.lazyGit.Resize(l.lgIW, l.lgIH)
	}
	if m.terminal != nil {
		m.terminal.Resize(l.lgIW, l.lgIH)
	}
	if m.melo != nil {
		m.melo.Resize(l.ocIW, l.ocIH)
	}
}

func (m *Model) cycleFocus() {
	switch m.view {
	case ViewIDE:
		switch m.focus {
		case FocusLazyVim:
			m.focus = FocusOpenCode
		case FocusOpenCode:
			m.focus = FocusLazyGit
		default:
			m.focus = FocusLazyVim
		}
	case ViewTerminal:
		switch m.focus {
		case FocusLazyVim:
			m.focus = FocusOpenCode
		case FocusOpenCode:
			m.focus = FocusTerminal
		default:
			m.focus = FocusLazyVim
		}
	case ViewDocs:
		switch m.focus {
		case FocusLazyVim:
			m.focus = FocusMarkdown
		case FocusMarkdown:
			m.focus = FocusLazyGit
		default:
			m.focus = FocusLazyVim
		}
	case ViewGit:
		m.focus = FocusLazyGit
	}
}

func (m *Model) handleChromeClick(x, y int) bool {
	if y == 0 {
		switch m.tabBar.HitTest(x) {
		case 0:
			m.view = ViewIDE
			if m.focus == FocusTerminal || m.focus == FocusMarkdown {
				m.focus = FocusLazyVim
			}
		case 1:
			m.view = ViewTerminal
			m.focus = FocusTerminal
		case 2:
			m.view = ViewGit
			m.focus = FocusLazyGit
		case 3:
			m.view = ViewDocs
			m.focus = FocusMarkdown
		}
		return true
	}

	if y == m.height-1 {
		id := m.statusBar.HitTest(x)
		if id == "" {
			return false
		}
		if id == "openmd" {
			m.state = stateOpenMD
			m.mdInput.Focus()
			return true
		}
		if ft, ok := focusIDToType(id); ok {
			m.focus = ft
			switch ft {
			case FocusTerminal:
				m.view = ViewTerminal
			case FocusMarkdown:
				m.view = ViewDocs
			default:
				m.view = ViewIDE
			}
		}
		return true
	}
	return false
}

func (m *Model) handlePaneMouse(msg tea.MouseMsg) {
	p, focus, localX, localY, ok := m.paneAt(msg.X, msg.Y)
	if !ok || p == nil {
		return
	}
	if focus == FocusTerminal {
		m.focus = focus
		return
	}
	if msg.Action == tea.MouseActionPress && msg.Button != tea.MouseButtonWheelUp && msg.Button != tea.MouseButtonWheelDown && msg.Button != tea.MouseButtonWheelLeft && msg.Button != tea.MouseButtonWheelRight {
		m.focus = focus
	}
	m.focus = focus
	if data := mouseToBytes(msg, localX, localY); len(data) > 0 {
		_ = p.Write(data)
	}
}

func (m *Model) paneAt(x, y int) (*pane.Pane, FocusType, int, int, bool) {

	contentY := y - tabBarHeight
	l := m.lay

	switch m.view {
	case ViewIDE:
		if contentY < l.lvH {
			if x < l.lvW {
				return m.lazyVim, FocusLazyVim, clampMouse(x-1, l.lvIW), clampMouse(contentY-1, l.lvIH), true
			} else {
				return m.openCode, FocusOpenCode, clampMouse(x-l.lvW-1, l.ocIW), clampMouse(contentY-1, l.ocIH), true
			}
		} else {
			return m.lazyGit, FocusLazyGit, clampMouse(x-1, l.lgIW), clampMouse(contentY-l.lvH-1, l.lgIH), true
		}
	case ViewTerminal:
		if contentY < l.lvH {
			if x < l.lvW {
				return m.lazyVim, FocusLazyVim, clampMouse(x-1, l.lvIW), clampMouse(contentY-1, l.lvIH), true
			} else {
				return m.openCode, FocusOpenCode, clampMouse(x-l.lvW-1, l.ocIW), clampMouse(contentY-1, l.ocIH), true
			}
		} else {
			return m.terminal, FocusTerminal, clampMouse(x-1, l.lgIW), clampMouse(contentY-l.lvH-1, l.lgIH), true
		}
	case ViewDocs:
		if contentY < l.lvH {
			if x < l.lvW {
				return m.lazyVim, FocusLazyVim, clampMouse(x-1, l.lvIW), clampMouse(contentY-1, l.lvIH), true
			} else {
				return m.melo, FocusMarkdown, clampMouse(x-l.lvW-1, l.ocIW), clampMouse(contentY-1, l.ocIH), true
			}
		} else {
			return m.lazyGit, FocusLazyGit, clampMouse(x-1, l.lgIW), clampMouse(contentY-l.lvH-1, l.lgIH), true
		}
	case ViewGit:
		return m.lazyGit, FocusLazyGit, clampMouse(x-1, max1(m.width-2, 1)), clampMouse(contentY-1, max1(m.height-tabBarHeight-statusBarHeight-2, 1)), true
	}
	return nil, 0, 0, 0, false
}

func clampMouse(v, max int) int {
	if v < 0 {
		return 0
	}
	if v >= max {
		return max - 1
	}
	return v
}

func mouseToBytes(msg tea.MouseMsg, x, y int) []byte {
	b := 0
	switch msg.Button {
	case tea.MouseButtonLeft:
		b = 0
	case tea.MouseButtonMiddle:
		b = 1
	case tea.MouseButtonRight:
		b = 2
	case tea.MouseButtonWheelUp:
		b = 64
	case tea.MouseButtonWheelDown:
		b = 65
	case tea.MouseButtonWheelLeft:
		b = 66
	case tea.MouseButtonWheelRight:
		b = 67
	case tea.MouseButtonNone:
		if msg.Action != tea.MouseActionMotion {
			return nil
		}
		b = 35
	default:
		return nil
	}
	if msg.Shift {
		b += 4
	}
	if msg.Alt {
		b += 8
	}
	if msg.Ctrl {
		b += 16
	}
	if msg.Action == tea.MouseActionMotion && msg.Button != tea.MouseButtonWheelUp && msg.Button != tea.MouseButtonWheelDown && msg.Button != tea.MouseButtonWheelLeft && msg.Button != tea.MouseButtonWheelRight {
		b += 32
	}
	suffix := 'M'
	if msg.Action == tea.MouseActionRelease {
		suffix = 'm'
	}
	return []byte(fmt.Sprintf("\x1b[<%d;%d;%d%c", b, x+1, y+1, suffix))
}

func keyToBytes(msg tea.KeyMsg) []byte {
	if msg.Type == tea.KeyRunes {
		return []byte(string(msg.Runes))
	}
	switch msg.Type {
	case tea.KeyEnter:
		return []byte{'\r'}
	case tea.KeyBackspace:
		return []byte{127}
	case tea.KeyDelete:
		return []byte{'\x1b', '[', '3', '~'}
	case tea.KeyTab:
		return []byte{'\t'}
	case tea.KeyEsc:
		return []byte{'\x1b'}
	case tea.KeySpace:
		return []byte{' '}
	case tea.KeyUp:
		return []byte{'\x1b', '[', 'A'}
	case tea.KeyDown:
		return []byte{'\x1b', '[', 'B'}
	case tea.KeyRight:
		return []byte{'\x1b', '[', 'C'}
	case tea.KeyLeft:
		return []byte{'\x1b', '[', 'D'}
	case tea.KeyHome:
		return []byte{'\x1b', '[', 'H'}
	case tea.KeyEnd:
		return []byte{'\x1b', '[', 'F'}
	case tea.KeyPgUp:
		return []byte{'\x1b', '[', '5', '~'}
	case tea.KeyPgDown:
		return []byte{'\x1b', '[', '6', '~'}
	case tea.KeyCtrlA:
		return []byte{1}
	case tea.KeyCtrlB:
		return []byte{2}
	case tea.KeyCtrlD:
		return []byte{4}
	case tea.KeyCtrlE:
		return []byte{5}
	case tea.KeyCtrlF:
		return []byte{6}
	case tea.KeyCtrlG:
		return []byte{7}
	case tea.KeyCtrlH:
		return []byte{8}
	case tea.KeyCtrlJ:
		return []byte{10}
	case tea.KeyCtrlK:
		return []byte{11}
	case tea.KeyCtrlL:
		return []byte{12}
	case tea.KeyCtrlN:
		return []byte{14}
	case tea.KeyCtrlO:
		return []byte{15}
	case tea.KeyCtrlP:
		return []byte{16}
	case tea.KeyCtrlQ:
		return []byte{17}
	case tea.KeyCtrlR:
		return []byte{18}
	case tea.KeyCtrlS:
		return []byte{19}
	case tea.KeyCtrlT:
		return []byte{20}
	case tea.KeyCtrlU:
		return []byte{21}
	case tea.KeyCtrlV:
		return []byte{22}
	case tea.KeyCtrlX:
		return []byte{24}
	case tea.KeyCtrlY:
		return []byte{25}
	case tea.KeyCtrlZ:
		return []byte{26}
	case tea.KeyF1:
		return []byte{'\x1b', 'O', 'P'}
	case tea.KeyF2:
		return []byte{'\x1b', 'O', 'Q'}
	case tea.KeyF3:
		return []byte{'\x1b', 'O', 'R'}
	case tea.KeyF4:
		return []byte{'\x1b', 'O', 'S'}
	case tea.KeyF5:
		return []byte{'\x1b', '[', '1', '5', '~'}
	case tea.KeyF6:
		return []byte{'\x1b', '[', '1', '7', '~'}
	case tea.KeyF7:
		return []byte{'\x1b', '[', '1', '8', '~'}
	case tea.KeyF8:
		return []byte{'\x1b', '[', '1', '9', '~'}
	case tea.KeyF9:
		return []byte{'\x1b', '[', '2', '0', '~'}
	case tea.KeyF10:
		return []byte{'\x1b', '[', '2', '1', '~'}
	case tea.KeyF11:
		return []byte{'\x1b', '[', '2', '3', '~'}
	case tea.KeyF12:
		return []byte{'\x1b', '[', '2', '4', '~'}
	}
	return nil
}
