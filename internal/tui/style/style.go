// Package style provides shared lipgloss styles used across TUI components.
package style

import (
	"charm.land/lipgloss/v2"
)

// Styles holds all reusable styles for the TUI.
type Styles struct {
	Highlighted lipgloss.Style
}

// DefaultStyles returns a Styles struct with sensible defaults.
func DefaultStyles() Styles {
	return Styles{
		Highlighted: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1),
	}
}

// Default is the package-level default styles instance.
var Default = DefaultStyles()
