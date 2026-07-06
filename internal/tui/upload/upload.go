// Package upload provides the TUI component for file uploading.
package upload

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
)

type Upload struct {
	keys KeyMap
}

var _ common.Component = (*Upload)(nil)

func New() *Upload {
	return &Upload{keys: DefaultKeyMap}
}

func (u *Upload) Init() tea.Cmd { return nil }

func (u *Upload) View() string {
	return "Hello Adit"
}

func (u *Upload) Update(tea.Msg) tea.Cmd {
	return nil
}

func (u *Upload) KeyMap() help.KeyMap { return u.keys }
