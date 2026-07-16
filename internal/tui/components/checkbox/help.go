package checkbox

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// KeyMap holds the key bindings for the checkbox component.
type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Space key.Binding
	Enter key.Binding
	Esc   key.Binding
}

// DefaultKeyMap returns the default key bindings.
var DefaultKeyMap = KeyMap{
	Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Space: key.NewBinding(key.WithKeys(" ", "space"), key.WithHelp("space", "toggle")),
	Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	Esc:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
}

// ShortHelp returns the short help view.
func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Up, km.Down, km.Space, km.Enter, km.Esc}
}

// FullHelp returns the full help view.
func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}
