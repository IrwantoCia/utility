// Package reader provides the Spritz reader TUI stub.
package reader

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// BackToSpritzMenuMsg tells the Spritz coordinator to return to the Spritz sub-menu.
type BackToSpritzMenuMsg struct{}

// Reader provides a stub page for the Spritz reader.
type Reader struct {
	keys       KeyMap
	helpModel  help.Model
	lastWindow tea.WindowSizeMsg
}

var _ common.Component = (*Reader)(nil)

// New creates a new Reader stub.
func New() *Reader {
	return &Reader{
		keys:      DefaultKeyMap,
		helpModel: help.New(),
	}
}

// Init is a no-op for the static stub.
func (r *Reader) Init() tea.Cmd { return nil }

// Resize stores the window dimensions and initialises the help model.
func (r *Reader) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	r.lastWindow = ws
	r.helpModel, _ = r.helpModel.Update(ws)
	return nil
}

// View renders a centered "Coming soon" message with a help bar at the bottom.
func (r *Reader) View() string {
	msg := "Reader — Coming soon"
	title := style.Default.MenuTitle.
		Width(r.lastWindow.Width).
		Render(msg)

	helpStr := r.helpModel.View(r.keys)
	helpHeight := lipgloss.Height(helpStr)

	availableHeight := r.lastWindow.Height - helpHeight

	var s strings.Builder
	for range availableHeight / 2 {
		s.WriteRune('\n')
	}
	s.WriteString(title)

	// Pad to fill remaining height before help
	for range r.lastWindow.Height - helpHeight - lipgloss.Height(s.String()) {
		s.WriteRune('\n')
	}

	s.WriteString(helpStr)
	return s.String()
}

// Update handles keyboard input. Esc returns to the Spritz menu.
func (r *Reader) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	if key.Matches(keyMsg, r.keys.Esc) {
		return func() tea.Msg { return BackToSpritzMenuMsg{} }
	}

	return nil
}
