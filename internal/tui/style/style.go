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
	TableBorder      lipgloss.Style // dark teal border for 3D inset table depth
	TableHeader      lipgloss.Style // bright cyan bg — raised/embossed header
	TableRow         lipgloss.Style // dark bg — solid table surface
	TableRowAlt      lipgloss.Style // darker bg — alternating depth contrast
	MenuItem         lipgloss.Style
	MenuItemSelected lipgloss.Style
	MenuTitle        lipgloss.Style
	MenuHint         lipgloss.Style
	MenuDesc         lipgloss.Style
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

	// ── Browse panel styles ───────────────────────────────────────
	BrowseBorderActive   lipgloss.Style
	BrowseBorderInactive lipgloss.Style
	BrowseBannerActive   lipgloss.Style
	BrowseBannerInactive lipgloss.Style
	BrowseListCursor     lipgloss.Style
	BrowseListSelected   lipgloss.Style
	BrowseListNormal     lipgloss.Style
	BrowseMetaLabel      lipgloss.Style
	BrowseMetaValue      lipgloss.Style
	BrowseMetaSection    lipgloss.Style
	BrowseMetaDim        lipgloss.Style
	BrowseEmpty          lipgloss.Style
	BrowseFilterBorder   lipgloss.Style
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
			Bold(true).
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("27")).
			Padding(0, 1),
		// TableBorder colors the table's NormalBorder in a dark teal ("30"),
		// creating a recessed/inset depth step against the panel's bright cyan ("45") border.
		TableBorder: lipgloss.NewStyle().
			Foreground(lipgloss.Color("30")),
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("39")).
			Padding(0, 1),
		// TableRow is the base style for data rows (non-alternating, non-highlighted).
		// A dark background grounds the table as a solid surface for the 3D effect.
		TableRow: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Padding(0, 1),
		// TableRowAlt uses a darker bg to create alternating depth contrast
		// that complements the raised-table appearance.
		TableRowAlt: lipgloss.NewStyle().
			Background(lipgloss.Color("234")).
			Padding(0, 1),
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
		MenuDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
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

		// ── Browse panel styles ───────────────────────────────────
		BrowseBorderActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("45")),
		BrowseBorderInactive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),
		BrowseBannerActive: lipgloss.NewStyle().
			Background(lipgloss.Color("45")).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1),
		BrowseBannerInactive: lipgloss.NewStyle().
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("250")).
			Padding(0, 1),
		BrowseListCursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("45")).
			Bold(true),
		BrowseListSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true),
		BrowseListNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		BrowseMetaLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("45")).
			Bold(true),
		BrowseMetaValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")),
		BrowseMetaSection: lipgloss.NewStyle().
			Foreground(lipgloss.Color("45")).
			Bold(true),
		BrowseMetaDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		BrowseEmpty: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true),
		BrowseFilterBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("206")),
	}
}

// Default is the package-level default styles instance.
var Default = DefaultStyles()
