# VibeTUI

A VS Code-like terminal IDE. No keybindings required.

## Panels

| Panel | What |
|-------|------|
| Editor (left) | LazyVim — full editor with file tree |
| Assistant (right) | OpenCode or Claude Code (configurable) |
| Git (bottom) | LazyGit — version control |

## Navigation

Click any panel to focus it. Use the tab bar at top to switch views.
Click the status bar buttons at the bottom to jump to a panel.
`Ctrl+\` cycles focus without the mouse.
Git and Terminal keep their state when you switch between them.

## Editor Shortcuts

| Key | Action |
|-----|--------|
| `Ctrl+P` | Find files |
| `Ctrl+B` | Toggle file explorer |
| `Ctrl+S` | Save |
| `Tab` | Next buffer tab |
| `Shift+Tab` | Previous buffer tab |

## Quit

`Ctrl+C` — exits vibetui with confirmation, except in the Terminal pane where it is sent through as a normal interrupt. Unattached vibetui tmux sessions are cleaned up automatically.
