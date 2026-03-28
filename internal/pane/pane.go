package pane

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/ActiveState/vt10x"
	"github.com/creack/pty"
)

// Pane manages a subprocess running inside a PTY with a VT100 virtual screen.
// All output from the subprocess is processed by the vt10x emulator so Cell()
// calls accurately reflect what the program has drawn.  User input is forwarded
// via Write(); Render() produces an ANSI string of exactly Width×Height
// characters ready to be wrapped in a lipgloss border.
type Pane struct {
	ID    string
	Title string

	cmd      string
	args     []string
	extraEnv []string

	ptyFile *os.File
	vt      *vt10x.VT
	state   vt10x.State

	mu      sync.Mutex
	width   int
	height  int
	running bool
}

func New(id, title, cmd string, args ...string) *Pane {
	return &Pane{ID: id, Title: title, cmd: cmd, args: args}
}

func (p *Pane) WithEnv(vars ...string) *Pane {
	p.extraEnv = append(p.extraEnv, vars...)
	return p
}

// Start launches the subprocess inside a PTY, attaches the VT100 emulator and
// begins forwarding output notifications to outputCh.
func (p *Pane) Start(width, height int, outputCh chan<- OutputMsg) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.width = width
	p.height = height

	c := exec.Command(p.cmd, p.args...)
	c.Env = append(os.Environ(), "TERM=xterm-256color")
	c.Env = append(c.Env, p.extraEnv...)

	f, err := pty.StartWithSize(c, &pty.Winsize{
		Rows: uint16(height),
		Cols: uint16(width),
	})
	if err != nil {
		return fmt.Errorf("pane %s: start pty: %w", p.ID, err)
	}
	p.ptyFile = f
	p.running = true

	// Create the VT emulator backed by the PTY file (acts as ReadWriteCloser).
	vt, err := vt10x.Create(&p.state, f)
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("pane %s: create vt: %w", p.ID, err)
	}
	p.vt = vt
	// Set initial dimensions in the VT state.
	p.vt.Resize(width, height)

	go p.parseLoop(outputCh)
	return nil
}

// parseLoop blocks on vt.Parse() which reads and processes PTY output until EOF.
func (p *Pane) parseLoop(outputCh chan<- OutputMsg) {
	for {
		err := p.vt.Parse()
		// Always notify so the view is refreshed after every batch of output.
		outputCh <- OutputMsg{PaneID: p.ID}
		if err != nil {
			p.mu.Lock()
			p.running = false
			p.mu.Unlock()
			return
		}
	}
}

// Write sends bytes directly to the subprocess stdin via the PTY master.
func (p *Pane) Write(data []byte) error {
	p.mu.Lock()
	f := p.ptyFile
	p.mu.Unlock()
	if f == nil {
		return nil
	}
	_, err := f.Write(data)
	return err
}

// Resize resizes the PTY window and the VT screen.
func (p *Pane) Resize(width, height int) {
	p.mu.Lock()
	p.width = width
	p.height = height
	f := p.ptyFile
	vt := p.vt
	p.mu.Unlock()

	if f != nil {
		_ = pty.Setsize(f, &pty.Winsize{
			Rows: uint16(height),
			Cols: uint16(width),
		})
	}
	if vt != nil {
		vt.Resize(width, height)
	}
}

// Render converts the current VT screen state to an ANSI-colored string with
// exactly Width columns and Height rows, suitable for embedding in a lipgloss
// border box.
func (p *Pane) Render() string {
	p.mu.Lock()
	w := p.width
	h := p.height
	p.mu.Unlock()

	if w <= 0 || h <= 0 {
		return ""
	}

	// Lock the state while reading cells so Parse() cannot mutate it concurrently.
	p.state.Lock()
	defer p.state.Unlock()

	var sb strings.Builder
	sb.Grow(h * (w + 20))

	for y := 0; y < h; y++ {
		sb.WriteString("\x1b[48;2;30;30;30m")
		prevFg := vt10x.DefaultFG
		prevBg := vt10x.DefaultBG

		for x := 0; x < w; x++ {
			ch, fg, bg := p.state.Cell(x, y)
			if ch == 0 {
				ch = ' '
			}
			fg = p.normalizeForeground(fg, bg)

			if fg != prevFg || bg != prevBg {
				fg1 := vtColorToANSI(fg, false)
				bg1 := vtColorToANSI(bg, true)

				var parts []string
				if fg1 != "" {
					parts = append(parts, fg1)
				}
				if bg1 != "" {
					parts = append(parts, bg1)
				}
				if len(parts) > 0 {
					sb.WriteString("\x1b[")
					sb.WriteString(strings.Join(parts, ";"))
					sb.WriteByte('m')
				}
				prevFg = fg
				prevBg = bg
			}

			sb.WriteRune(ch)
		}

		if y < h-1 {
			sb.WriteString("\x1b[0m\x1b[48;2;30;30;30m\n")
		} else {
			sb.WriteString("\x1b[0m")
		}
	}

	return sb.String()
}

// IsRunning returns true while the managed subprocess is still alive.
func (p *Pane) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

// Close terminates the PTY and subprocess.
func (p *Pane) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.running = false
	if p.ptyFile != nil {
		_ = p.ptyFile.Close()
		p.ptyFile = nil
	}
}

func (p *Pane) normalizeForeground(fg, bg vt10x.Color) vt10x.Color {
	if p.ID != "opencode" {
		return fg
	}
	switch fg {
	case vt10x.DarkGrey, vt10x.Magenta, vt10x.Blue, vt10x.Cyan:
		if bg == vt10x.Black || bg == vt10x.DefaultBG || bg == vt10x.DarkGrey {
			return vt10x.White
		}
	}
	return fg
}

// ─── color helpers ────────────────────────────────────────────────────────────

// vtColorToANSI converts a vt10x.Color to an ANSI SGR parameter string.
// bg=true produces background codes (40–47, 100–107); bg=false gives foreground.
func vtColorToANSI(c vt10x.Color, bg bool) string {
	base := 30
	if bg {
		base = 40
	}

	switch c {
	case vt10x.Black:
		return fmt.Sprintf("%d", base)
	case vt10x.Red:
		return fmt.Sprintf("%d", base+1)
	case vt10x.Green:
		return fmt.Sprintf("%d", base+2)
	case vt10x.Yellow:
		return fmt.Sprintf("%d", base+3)
	case vt10x.Blue:
		return fmt.Sprintf("%d", base+4)
	case vt10x.Magenta:
		return fmt.Sprintf("%d", base+5)
	case vt10x.Cyan:
		return fmt.Sprintf("%d", base+6)
	case vt10x.LightGrey:
		return fmt.Sprintf("%d", base+7)
	case vt10x.DarkGrey:
		return fmt.Sprintf("%d", base+60)
	case vt10x.LightRed:
		return fmt.Sprintf("%d", base+61)
	case vt10x.LightGreen:
		return fmt.Sprintf("%d", base+62)
	case vt10x.LightYellow:
		return fmt.Sprintf("%d", base+63)
	case vt10x.LightBlue:
		return fmt.Sprintf("%d", base+64)
	case vt10x.LightMagenta:
		return fmt.Sprintf("%d", base+65)
	case vt10x.LightCyan:
		return fmt.Sprintf("%d", base+66)
	case vt10x.White:
		return fmt.Sprintf("%d", base+67)
	case vt10x.DefaultFG:
		if !bg {
			return "39"
		}
	case vt10x.DefaultBG:
		if bg {
			return "48;2;30;30;30"
		}
	}
	return ""
}
