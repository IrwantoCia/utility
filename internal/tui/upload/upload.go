// Package upload provides the TUI component for file uploading.
package upload

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
)

type Upload struct {
	keys      KeyMap
	helpModel help.Model
}

var _ common.Component = (*Upload)(nil)

func New() *Upload {
	return &Upload{
		keys:      DefaultKeyMap,
		helpModel: help.New(),
	}
}

func (u *Upload) Init() tea.Cmd { return nil }

func (u *Upload) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	u.helpModel, _ = u.helpModel.Update(ws)
	return nil
}

func (u *Upload) View() string {
	return "Hello Adit\n\n" + u.helpModel.View(u.keys)
}

func (u *Upload) Update(msg tea.Msg) tea.Cmd {
	u.helpModel, _ = u.helpModel.Update(msg)

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if key.Matches(keyMsg, u.keys.Esc) {
			return func() tea.Msg {
				return common.BackToMenuMsg{}
			}
		}
	}

	return nil
}
