package transcribe

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// KeyMap defines key bindings for the transcribe component.
type KeyMap struct {
	Up   key.Binding
	Down key.Binding
	Esc  key.Binding
}

// DefaultKeyMap is the default set of key bindings.
var DefaultKeyMap = KeyMap{
	Up:   key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "up")),
	Down: key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "down")),
	Esc:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Esc, km.Up, km.Down}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}
