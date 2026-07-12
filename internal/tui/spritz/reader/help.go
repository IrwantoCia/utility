// Package reader provides the Spritz reader TUI.
package reader

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// KeyMap holds key bindings for the reader page.
type KeyMap struct {
	Esc   key.Binding
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Space key.Binding
	Chunk key.Binding
	Plus  key.Binding
	Minus key.Binding
	Next  key.Binding
	Prev  key.Binding
}

// DefaultKeyMap is the default set of key bindings.
var DefaultKeyMap = KeyMap{
	Esc:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Up:    key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:  key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
	Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Space: key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "play/pause")),
	Chunk: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "mode")),
	Plus:  key.NewBinding(key.WithKeys("=", "+"), key.WithHelp("+", "faster")),
	Minus: key.NewBinding(key.WithKeys("-"), key.WithHelp("-", "slower")),
	Next:  key.NewBinding(key.WithKeys("j", "right"), key.WithHelp("j/→", "next word")),
	Prev:  key.NewBinding(key.WithKeys("k", "left"), key.WithHelp("k/←", "prev word")),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Up, km.Down, km.Enter, km.Esc, km.Space, km.Plus, km.Minus, km.Next, km.Prev, km.Chunk}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}
