// Package csv provides the TUI component for CSV viewing and editing.
package csv

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
)

type Model struct {
	keys      KeyMap
	helpModel help.Model
}

var _ common.Component = (*Model)(nil)

func New() *Model {
	return &Model{
		keys:      DefaultKeyMap,
		helpModel: help.New(),
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	m.helpModel, _ = m.helpModel.Update(ws)
	return nil
}

func (m *Model) View() string {
	return "CSV\n\n" + m.helpModel.View(m.keys)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.helpModel, _ = m.helpModel.Update(msg)

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if key.Matches(keyMsg, m.keys.Esc) {
			return func() tea.Msg {
				return common.BackToMenuMsg{}
			}
		}
	}

	return nil
}
