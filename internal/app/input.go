package app

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
)

// keyToBytes translates a Bubble Tea key message into the raw byte sequence
// that the focused PTY subprocess expects to receive on stdin.
// It maps named keys to their VT100/xterm escape sequences and passes
// printable runes through unchanged.
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

// mouseToBytes encodes a Bubble Tea mouse event as an SGR mouse sequence
// (CSI < Pb ; Px ; Py M/m) for delivery to the focused PTY subprocess.
// x and y are pane-local coordinates (0-based); the sequence uses 1-based columns/rows.
// Modifier bits follow the X10 convention: shift=4, alt=8, ctrl=16, motion=32.
func mouseToBytes(msg tea.MouseMsg, x, y int) []byte {
	// Map button to base button code per SGR / X10 encoding.
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
	// OR in modifier bits.
	if msg.Shift {
		b += 4
	}
	if msg.Alt {
		b += 8
	}
	if msg.Ctrl {
		b += 16
	}
	// Motion flag applies to non-wheel buttons only.
	if msg.Action == tea.MouseActionMotion &&
		msg.Button != tea.MouseButtonWheelUp &&
		msg.Button != tea.MouseButtonWheelDown &&
		msg.Button != tea.MouseButtonWheelLeft &&
		msg.Button != tea.MouseButtonWheelRight {
		b += 32
	}
	// Press = 'M', release = 'm'.
	suffix := 'M'
	if msg.Action == tea.MouseActionRelease {
		suffix = 'm'
	}
	return []byte(fmt.Sprintf("\x1b[<%d;%d;%d%c", b, x+1, y+1, suffix))
}
