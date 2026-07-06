// Package common provides shared interfaces and types used across the TUI
// application components.
package common

import tea "charm.land/bubbletea/v2"

type Component interface {
	Init() tea.Cmd
	View() string
	Update(msg tea.Msg) tea.Cmd
	Resize(ws tea.WindowSizeMsg) tea.Cmd
}
