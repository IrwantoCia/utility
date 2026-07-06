package csv

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Esc   key.Binding
	Enter key.Binding
}

var DefaultKeyMap = KeyMap{
	Esc:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "pick file")),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Esc, km.Enter}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}
