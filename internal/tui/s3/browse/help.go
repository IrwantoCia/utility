// Package browse provides the 2-panel S3 browser coordinator.
package browse

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// KeyMap holds key bindings for the browse page.
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Esc    key.Binding
	PgUp   key.Binding
	PgDown key.Binding
}

// DefaultKeyMap is the default set of key bindings for the browse page.
var DefaultKeyMap = KeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/↓/j/k", "navigate")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↑/↓/j/k", "navigate")),
	Left:   key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/→/h/l", "panel")),
	Right:  key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("←/→/h/l", "panel")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Esc:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	PgUp:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("^U", "pg up")),
	PgDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("^D", "pg down")),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Up, km.Down, km.PgUp, km.PgDown, km.Enter, km.Left, km.Esc}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}
