// Package reader provides the Spritz reader TUI.
package reader

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// KeyMap holds key bindings for the reader page.
type KeyMap struct {
	Esc key.Binding
}

// DefaultKeyMap is the default set of key bindings (only Esc for back).
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
