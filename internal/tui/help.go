package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// KeyMap defines key bindings for the main menu.
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Quit   key.Binding
	Esc    key.Binding
}

// DefaultKeyMap is the default set of key bindings for the main menu.
var DefaultKeyMap = KeyMap{
	Up:     key.NewBinding(key.WithKeys("k", "up"),     key.WithHelp("k", "up")),
	Down:   key.NewBinding(key.WithKeys("j", "down"),   key.WithHelp("j", "down")),
	Select: key.NewBinding(key.WithKeys("enter"),       key.WithHelp("enter", "select")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Esc:    key.NewBinding(key.WithKeys("esc"),         key.WithHelp("esc", "back")),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Up, km.Down, km.Select, km.Quit}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}
