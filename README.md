# vibetui

`vibetui` is a tmux-powered terminal IDE layout designed for Ghostty and other modern terminals.

It launches:
- **LazyVim** on the left
- **Claude Code** on the right
- **LazyGit** on the bottom, with a one-click swap to a plain terminal

The goal is a mouse-friendly, VS Code-like terminal workspace without forcing users to memorize a lot of Vim or tmux commands.

## What vibetui does

When you run `vibetui`, it:

1. creates `~/.config/vibetui/` if needed
2. writes bundled config for Neovim and LazyGit
3. starts a fresh tmux session
4. opens a 3-pane layout:
   - **left:** LazyVim
   - **right:** Claude Code (`claude`)
   - **bottom:** LazyGit, swappable with Terminal

## Requirements

These tools must be installed and available on your `PATH`:

- `tmux`
- `nvim`
- `lazygit`
- `claude`
- `go` (to build/install `vibetui`)

> `vibetui` shells out to these CLIs directly. If one is missing, the app will not work correctly.

---

## 1) Install dependencies

### macOS (recommended)

Install Homebrew if you do not already have it:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

Then install the required tools:

```bash
brew install tmux
brew install neovim
brew install lazygit
brew install go
```

Install Claude Code:

```bash
curl -fsSL https://claude.ai/install.sh | bash
```

Alternative Claude install on macOS:

```bash
brew install --cask claude-code
```

### Linux

Install the basics with your package manager:

```bash
# Debian / Ubuntu
sudo apt install tmux neovim

# Fedora / RHEL
sudo dnf install tmux neovim
```

Install Go using the official instructions:

- https://go.dev/doc/install

Install Claude Code:

```bash
curl -fsSL https://claude.ai/install.sh | bash
```

Install LazyGit:

```bash
go install github.com/jesseduffield/lazygit@latest
```

> On some Linux distros, the packaged Neovim version is very old. If that happens, use a newer package source or Homebrew on Linux.

---

## 2) Authenticate Claude Code

Run:

```bash
claude
```

On first launch, Claude Code opens a browser so you can sign in.

If the browser does not open automatically, follow the CLI prompt and copy the login URL manually.

Claude Code requires a supported Anthropic plan/account.

More info:
- https://code.claude.com/docs/en/setup
- https://code.claude.com/docs/en/authentication

---

## 3) Verify your dependencies

Run this before installing `vibetui`:

```bash
tmux -V
nvim --version
lazygit --version
claude --version
go version
```

If any command fails, fix that first.

---

## 4) Install vibetui

### Option A: build from a local clone

Clone the repo:

```bash
git clone https://github.com/hritupitu/vibetui.git
cd vibetui
```

Build a local binary:

```bash
go build -o vibetui .
```

Run it directly:

```bash
./vibetui
```

### Option B: install globally from your local clone

From the repo root:

```bash
go install .
```

That installs the binary into your Go bin directory, usually:

```bash
~/go/bin
```

### Option C: install globally from GitHub

```bash
go install github.com/hritupitu/vibetui@latest
```

---

## 5) Add Go bin to your PATH

If `vibetui` installs successfully but the command is not found, your Go bin directory is probably not on `PATH`.

### zsh (default on modern macOS)

Add this to `~/.zshrc`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

Then reload your shell:

```bash
source ~/.zshrc
```

### bash

Add this to `~/.bashrc` or `~/.bash_profile`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

Then reload:

```bash
source ~/.bashrc
```

### fish

```fish
fish_add_path $HOME/go/bin
```

### Confirm it worked

```bash
which vibetui
vibetui
```

---

## 6) First run

Start `vibetui`:

```bash
vibetui
```

On launch, it automatically writes bundled config to:

```bash
~/.config/vibetui/
```

This currently includes:

- `~/.config/vibetui/init.lua`
- `~/.config/vibetui/lazygit/config.yml`
- `~/.config/vibetui/welcome.md`

`vibetui` also writes:

- `~/.config/vibetui/opencode-tui.json`
- `~/.config/opencode/themes/vibetui-cherry.json`

Those extra theme files are harmless and are created automatically by the current config bootstrap.

---

## Layout and controls

### Pane layout

- **Left:** LazyVim
- **Right:** Claude Code
- **Bottom:** LazyGit or Terminal

### Bottom pane switching

- click **`[Git]`** in the top status bar to show LazyGit
- click **`[Terminal]`** in the top status bar to show a shell
- press **`F1`** for Git
- press **`F2`** for Terminal

### Exit

- press **`Ctrl+C`**
- vibetui shows a confirmation prompt before closing the tmux session

### LazyVim behavior inside vibetui

The bundled Neovim config is tuned to feel less modal than a stock Vim setup:

- editable file buffers try to keep you in insert mode
- special side panels like Neo-tree stay interactive
- files autosave aggressively
- `Ctrl+S` saves manually if you want it
- `Ctrl+P` opens file search
- `Ctrl+B` toggles the file tree

---

## Updating vibetui

### If you installed from a local clone

```bash
cd /path/to/vibetui
git pull
go install .
```

### If you installed from GitHub

```bash
go install github.com/hritupitu/vibetui@latest
```

---

## Troubleshooting

### `vibetui: command not found`

Your Go bin directory is not on `PATH`.

Fix it by adding:

```bash
export PATH="$HOME/go/bin:$PATH"
```

to your shell config, then restart the shell.

### `tmux is required`

Install tmux and verify:

```bash
tmux -V
```

### `claude: command not found`

Install Claude Code and verify:

```bash
claude --version
```

Then run:

```bash
claude
```

to complete login.

### `lazygit: command not found`

Install LazyGit and verify:

```bash
lazygit --version
```

### `nvim: command not found`

Install Neovim and verify:

```bash
nvim --version
```

### Colors or mouse behavior seem wrong

Use a modern terminal with good tmux support and truecolor enabled. Ghostty is the intended target setup.

### Claude pane opens but is not authenticated

Run this once outside vibetui:

```bash
claude
```

and finish the browser login flow.

---

## Development

Build locally:

```bash
make build
```

Run locally:

```bash
make run
```

Clean local build artifact:

```bash
make clean
```

---

## Current architecture

The active runtime path is tmux-based.

The app entrypoint calls:

- `config.Setup()`
- `tmuxsession.Launch(...)`

So the current user-facing product is the tmux launcher, not the older Bubble Tea prototype still present in the repo.
