package listpicker

import "charm.land/lipgloss/v2"

// Styles holds the visual styles for the ListPicker component.
type Styles struct {
	Box      lipgloss.Style // rounded border container
	Title    lipgloss.Style // colored title bar
	Item     lipgloss.Style
	Selected lipgloss.Style
}

// DefaultStyles returns the default ListPicker styles.
func DefaultStyles() Styles {
	accent := lipgloss.Color("62")
	return Styles{
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")).
			Background(accent).
			Padding(0, 1),
		Item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("248")),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")).
			Background(accent),
	}
}
