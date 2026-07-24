// Package menu provides the SQLite database sub-menu page.
package menu

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Esc   key.Binding
}

var DefaultKeyMap = KeyMap{
	Up:    key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:  key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
	Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Esc:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Up, km.Down, km.Enter, km.Esc}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}
