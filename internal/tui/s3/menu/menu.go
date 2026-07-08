// Package menu provides the S3 sub-menu TUI with Upload and Browse options.
package menu

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// ShowUploadMsg tells the S3 coordinator to navigate to the upload page.
type ShowUploadMsg struct{}

// ShowBrowseMsg tells the S3 coordinator to navigate to the browse page.
type ShowBrowseMsg struct{}

// Menu represents the S3 sub-menu allowing the user to choose Upload or Browse.
type Menu struct {
	items      []string
	cursor     int
	keys       KeyMap
	helpModel  help.Model
	lastWindow tea.WindowSizeMsg
}

var _ common.Component = (*Menu)(nil)

// New creates a new S3 sub-menu.
func New() *Menu {
	return &Menu{
		items:     []string{"Upload", "Browse"},
		keys:      DefaultKeyMap,
		helpModel: help.New(),
	}
}

// Init is a no-op for the static menu.
func (m *Menu) Init() tea.Cmd { return nil }

// Resize stores the window dimensions and initialises the help model.
func (m *Menu) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	m.lastWindow = ws
	m.helpModel, _ = m.helpModel.Update(ws)
	return nil
}

// View renders the menu with cursor highlighting.
func (m *Menu) View() string {
	var content strings.Builder
	content.WriteString("S3\n\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}

		line := cursor + item
		if i == m.cursor {
			line = style.Default.Highlighted.Render(line)
		}
		content.WriteString(line)
		content.WriteByte('\n')
	}

	helpStr := m.helpModel.View(m.keys)

	var buf strings.Builder
	buf.WriteString(content.String())

	// Pad to fill remaining height before help.
	for i := lipgloss.Height(buf.String()); i <= m.lastWindow.Height-lipgloss.Height(helpStr); i++ {
		buf.WriteRune('\n')
	}

	buf.WriteString(helpStr)
	return buf.String()
}

// Update handles keyboard navigation: k/up, j/down, Enter, Esc.
func (m *Menu) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(keyMsg, m.keys.Esc):
		return func() tea.Msg {
			return common.BackToMenuMsg{}
		}
	case key.Matches(keyMsg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(keyMsg, m.keys.Down):
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case key.Matches(keyMsg, m.keys.Enter):
		switch m.cursor {
		case 0:
			return func() tea.Msg {
				return ShowUploadMsg{}
			}
		case 1:
			return func() tea.Msg {
				return ShowBrowseMsg{}
			}
		}
	}

	return nil
}
