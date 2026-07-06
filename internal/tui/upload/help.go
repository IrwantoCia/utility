package upload

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// KeyMap defines key bindings for the upload component.
type KeyMap struct {
	Esc key.Binding
}

// DefaultKeyMap is the default set of key bindings.
var DefaultKeyMap = KeyMap{
	Esc: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Esc}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}
