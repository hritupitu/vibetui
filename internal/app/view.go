package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	panePkg "github.com/hritupitu/vibetui/internal/pane"
	"github.com/hritupitu/vibetui/internal/ui"
)

// View renders the current app state as a Bubble Tea screen.
func (m Model) View() string {
	if !m.ready {
		return "Initializing…"
	}

	switch m.state {
	case stateConfirmQuit:
		return m.renderConfirmQuit()
	case stateOpenMD:
		return m.renderOpenMD()
	}

	var sb strings.Builder

	sb.WriteString(m.tabBar.Render(m.width, viewToTabIdx(m.view)))
	sb.WriteByte('\n')

	switch m.view {
	case ViewIDE:
		sb.WriteString(m.renderIDE())
	case ViewTerminal:
		sb.WriteString(m.renderTerminal())
	case ViewGit:
		sb.WriteString(m.renderGitFull())
	case ViewDocs:
		sb.WriteString(m.renderDocs())
	}

	sb.WriteByte('\n')
	sb.WriteString(m.statusBar.Render(m.width, m.currentPaneID(), m.view == ViewDocs))

	return sb.String()
}

func (m Model) renderIDE() string {
	l := m.lay
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top,
		paneBox(pRender(m.lazyVim), l.lvIW, l.lvIH, pTitle(m.lazyVim), m.focus == FocusLazyVim),
		paneBox(pRender(m.openCode), l.ocIW, l.ocIH, pTitle(m.openCode), m.focus == FocusOpenCode),
	)
	lgBox := paneBox(pRender(m.lazyGit), l.lgIW, l.lgIH, pTitle(m.lazyGit), m.focus == FocusLazyGit)
	return lipgloss.JoinVertical(lipgloss.Left, mainRow, lgBox)
}

func (m Model) renderTerminal() string {
	l := m.lay
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top,
		paneBox(pRender(m.lazyVim), l.lvIW, l.lvIH, pTitle(m.lazyVim), m.focus == FocusLazyVim),
		paneBox(pRender(m.openCode), l.ocIW, l.ocIH, pTitle(m.openCode), m.focus == FocusOpenCode),
	)
	termBox := paneBox(pRender(m.terminal), l.lgIW, l.lgIH, pTitle(m.terminal), m.focus == FocusTerminal)
	return lipgloss.JoinVertical(lipgloss.Left, mainRow, termBox)
}

func (m Model) renderGitFull() string {
	iw := max1(m.width-2, 1)
	ih := max1(m.height-tabBarHeight-statusBarHeight-2, 1)
	return paneBox(pRender(m.lazyGit), iw, ih, pTitle(m.lazyGit), true)
}

func (m Model) renderDocs() string {
	l := m.lay
	meloContent := pRender(m.melo)
	meloTitle := pTitle(m.melo)
	if m.melo == nil {
		meloTitle = " Docs"
		meloContent = lipgloss.NewStyle().
			Width(l.ocIW).Height(l.ocIH).
			Align(lipgloss.Center, lipgloss.Center).
			Faint(true).
			Render("melo not installed\n\ncargo install --path .\ngithub.com/mw2000/melo")
	}
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top,
		paneBox(pRender(m.lazyVim), l.lvIW, l.lvIH, pTitle(m.lazyVim), m.focus == FocusLazyVim),
		paneBox(meloContent, l.ocIW, l.ocIH, meloTitle, m.focus == FocusMarkdown),
	)
	lgBox := paneBox(pRender(m.lazyGit), l.lgIW, l.lgIH, pTitle(m.lazyGit), m.focus == FocusLazyGit)
	return lipgloss.JoinVertical(lipgloss.Left, mainRow, lgBox)
}

func (m Model) renderConfirmQuit() string {
	yes, no := ui.ConfirmInactiveBtn.Render("  Yes  "), ui.ConfirmActiveBtn.Render("  No  ")
	if m.confirmYes {
		yes, no = ui.ConfirmActiveBtn.Render("  Yes  "), ui.ConfirmInactiveBtn.Render("  No  ")
	}
	body := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render("Quit vibetui?"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Center, yes, no),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		ui.ConfirmDialogStyle.Render(body))
}

func (m Model) renderOpenMD() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("Open Markdown File"),
		"",
		m.mdInput.View(),
		"",
		lipgloss.NewStyle().Faint(true).Render("Enter to open · Esc to cancel"),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		ui.ConfirmDialogStyle.Width(52).Render(body))
}

func paneBox(content string, innerW, innerH int, title string, focused bool) string {
	box := lipgloss.NewStyle().
		Width(innerW).
		Height(innerH).
		Background(lipgloss.Color("#1e1e1e")).
		Border(ui.PaneBorder(focused)).
		Render(content)
	return spliceTopBorderTitle(box, title, focused)
}

func spliceTopBorderTitle(box, title string, focused bool) string {
	nl := strings.IndexByte(box, '\n')
	if nl < 0 {
		return box
	}
	topLine, rest := box[:nl], box[nl:]

	isThick := strings.Contains(topLine, "━")
	openCorner := "╭"
	closeCorner := "╮"
	dash := "─"
	if isThick {
		openCorner = "┏"
		closeCorner = "┓"
		dash = "━"
	}

	cornerIdx := strings.Index(topLine, openCorner)
	if cornerIdx < 0 {
		return box
	}

	dashCount := strings.Count(topLine, dash)
	titleW := lipgloss.Width(title)
	if titleW+4 > dashCount {
		return box
	}

	prefix := topLine[:cornerIdx]
	closingIdx := strings.LastIndex(topLine, closeCorner)
	suffix := ""
	if closingIdx >= 0 {
		suffix = topLine[closingIdx+len(closeCorner):]
	}

	// Keep any ANSI prefix that lipgloss injected before the visible border runes
	// while replacing a short visible segment of the top border with the title.
	styledTitle := ui.PaneTitleStyle(title, focused)
	dashesAfter := dashCount - titleW - 2
	newTop := prefix + openCorner + dash + styledTitle + prefix + strings.Repeat(dash, dashesAfter) + closeCorner + suffix
	return newTop + rest
}

func pRender(p *panePkg.Pane) string {
	if p == nil {
		return ""
	}
	return p.Render()
}

func pTitle(p *panePkg.Pane) string {
	if p == nil {
		return ""
	}
	return p.Title
}

func viewToTabIdx(v ViewType) int {
	switch v {
	case ViewIDE:
		return 0
	case ViewTerminal:
		return 1
	case ViewGit:
		return 2
	case ViewDocs:
		return 3
	}
	return 0
}
