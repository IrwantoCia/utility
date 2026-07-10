// Package style provides shared lipgloss styles used across TUI components.
package style

import (
	"charm.land/lipgloss/v2"
)

// Styles holds all reusable styles for the TUI.
type Styles struct {
	Highlighted      lipgloss.Style
	Action           lipgloss.Style
	RowHighlighted   lipgloss.Style
	TableHeader      lipgloss.Style
	TableRowAlt      lipgloss.Style
	MenuItem         lipgloss.Style
	MenuItemSelected lipgloss.Style
	MenuTitle        lipgloss.Style
	MenuHint         lipgloss.Style
	MenuContainer    lipgloss.Style
	CardIcon          lipgloss.Style
	CardIconInput     lipgloss.Style
	CardIconAction    lipgloss.Style
	CardTitle         lipgloss.Style
	CardTitleSelected lipgloss.Style
	CardDesc          lipgloss.Style
	CardContainer     lipgloss.Style
	EnvSectionBorder  lipgloss.Style
	EnvSectionTitle   lipgloss.Style
	EnvKey            lipgloss.Style
	EnvValue          lipgloss.Style
	EnvValueMasked    lipgloss.Style
	EnvValueExample   lipgloss.Style
	EnvStatusSet      lipgloss.Style
	EnvStatusMissing  lipgloss.Style
	HelpKey       lipgloss.Style
	StatusBox     lipgloss.Style
	StatusText    lipgloss.Style
	StatusError   lipgloss.Style
	StatusSuccess lipgloss.Style
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
		Action: lipgloss.NewStyle().
			Background(lipgloss.Color("2")).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1),
		RowHighlighted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")),
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("63")),
		TableRowAlt: lipgloss.NewStyle().
			Background(lipgloss.Color("235")),
		MenuItem: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2).
			Foreground(lipgloss.Color("240")),
		MenuItemSelected: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2).
			Bold(true).
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")),
		MenuTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69")).
			AlignHorizontal(lipgloss.Center),
		MenuHint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			AlignHorizontal(lipgloss.Center).
			Italic(true),
		MenuContainer: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 3),
		CardIcon: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		CardIconInput: lipgloss.NewStyle().
			Foreground(lipgloss.Color("75")).
			Bold(true),
		CardIconAction: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true),
		CardTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		CardTitleSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true),
		CardDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true),
		CardContainer: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1),
		EnvSectionBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 1),
		EnvSectionTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69")),
		EnvKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(15),
		EnvValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		EnvValueMasked: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true),
		EnvValueExample: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true),
		EnvStatusSet: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")),
		EnvStatusMissing: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		HelpKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("240")),
		StatusBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		StatusText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		StatusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		StatusSuccess: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true),
	}
}

// Default is the package-level default styles instance.
var Default = DefaultStyles()
