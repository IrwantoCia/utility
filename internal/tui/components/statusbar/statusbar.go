// Package statusbar provides a simple status message display component.
// It renders a bordered box around a single-line message with color-coded
// styling. When the message is empty, View returns an empty string.
//
// This component is NOT a common.Component — it has no lifecycle methods.
// Callers set the message in Update and call View in their own View method.
package statusbar

import (
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// Kind categorises the status message for styling.
type Kind int

const (
	Info    Kind = iota // normal text (white)
	Error               // red bold text
	Success             // green bold text
)

// StatusBar holds a single status message and its kind.
type StatusBar struct {
	msg  string
	kind Kind
}

// New creates an empty StatusBar.
func New() *StatusBar {
	return &StatusBar{}
}

// SetInfo sets the message with normal styling.
func (s *StatusBar) SetInfo(msg string) {
	s.msg = msg
	s.kind = Info
}

// SetError sets the message with red error styling.
func (s *StatusBar) SetError(msg string) {
	s.msg = msg
	s.kind = Error
}

// SetSuccess sets the message with green success styling.
func (s *StatusBar) SetSuccess(msg string) {
	s.msg = msg
	s.kind = Success
}

// Clear resets the status bar to empty.
func (s *StatusBar) Clear() {
	s.msg = ""
}

// View renders a bordered box containing the status message.
// Returns an empty string when there is no message.
func (s *StatusBar) View(width int) string {
	if s.msg == "" {
		return ""
	}

	var textStyle lipgloss.Style
	switch s.kind {
	case Error:
		textStyle = style.Default.StatusError
	case Success:
		textStyle = style.Default.StatusSuccess
	default:
		textStyle = style.Default.StatusText
	}

	return style.Default.StatusBox.
		Width(width).
		Render(textStyle.Render(s.msg))
}
