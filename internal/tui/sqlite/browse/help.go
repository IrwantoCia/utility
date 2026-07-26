// Package browse provides the SQLite table browser TUI with a 3-panel
// split-pane layout.
package browse

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
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
	Filter key.Binding
	Copy   key.Binding
	Reload key.Binding
}

// DefaultKeyMap is the default set of key bindings for the browse page.
var DefaultKeyMap = KeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Left:   key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "quick left")),
	Right:  key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "quick right")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Esc:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	PgUp:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "pg up")),
	PgDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "pg down")),
	Filter: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	Copy:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy")),
	Reload: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reload")),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Up, km.Down, km.PgUp, km.PgDown, km.Enter, km.Left, km.Filter, km.Copy, km.Reload, km.Esc}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{km.ShortHelp()}
}

var _ help.KeyMap = KeyMap{}

// BrowseHelpStyles returns a help.Styles configured with the browse theme:
// keys in amber, descriptions in slate, separators in dark gray.
func BrowseHelpStyles() help.Styles {
	return help.Styles{
		Ellipsis: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		ShortKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("45")).
			Bold(true),
		ShortDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")),
		ShortSeparator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		FullKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("45")).
			Bold(true),
		FullDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")),
		FullSeparator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}
