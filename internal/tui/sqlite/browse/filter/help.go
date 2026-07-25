// Package filter provides a WHERE-clause filter editor panel for the SQLite
// browse feature, allowing users to add, edit, and remove filter rows.
package filter

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// KeyMap holds key bindings for the filter panel.
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Delete key.Binding
	Esc    key.Binding
}

// DefaultKeyMap is the default set of key bindings for the filter panel.
var DefaultKeyMap = KeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "edit/add")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Esc:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}

// ShortHelp returns bindings for the short help bar.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Delete}
}

// FullHelp returns bindings for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{
		k.Up, k.Down, k.Enter, k.Delete, k.Esc,
	}}
}

var _ help.KeyMap = KeyMap{}
