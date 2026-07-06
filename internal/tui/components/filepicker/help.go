package filepicker

import (
	fp "charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Esc   key.Binding
	Enter key.Binding
}

var DefaultKeyMap = KeyMap{
	Esc:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Enter: key.NewBinding(key.WithKeys("enter")),
}

func (km KeyMap) ShortHelp() []key.Binding {
	fpk := fp.DefaultKeyMap()
	return []key.Binding{
		km.Esc,
		fpk.Up,
		fpk.Down,
		fpk.GoToTop,
		fpk.GoToLast,
		fpk.Back,
		fpk.Open,
	}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}
