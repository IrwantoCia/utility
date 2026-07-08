// Package browse provides the 2-panel S3 browser coordinator.
package browse

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// KeyMap holds key bindings for the browse page.
type KeyMap struct {
	Esc key.Binding
	Tab key.Binding
}

// DefaultKeyMap is the default set of key bindings for the browse page.
var DefaultKeyMap = KeyMap{
	Esc: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Tab: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch panel")),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Tab, km.Esc}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}
