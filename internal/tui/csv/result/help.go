// Package result provides the CSV data display and filter page.
package result

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Tab   key.Binding
	Enter key.Binding
	Esc   key.Binding
	Left  key.Binding
	Right key.Binding
}

var DefaultKeyMap = KeyMap{
	Up:    key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:  key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
	Tab:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "toggle")),
	Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "filter")),
	Esc:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Left:  key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("h/←", "left")),
	Right: key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("l/→", "right")),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Up, km.Down, km.Left, km.Right, km.Tab, km.Enter, km.Esc}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}
