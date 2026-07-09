package listpicker

import "charm.land/lipgloss/v2"

// Styles holds the visual styles for the ListPicker component.
type Styles struct {
	Title    lipgloss.Style
	Item     lipgloss.Style
	Selected lipgloss.Style
}

// DefaultStyles returns the default ListPicker styles.
func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69")).
			AlignHorizontal(lipgloss.Center),
		Item: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2).
			Foreground(lipgloss.Color("240")),
		Selected: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2).
			Bold(true).
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")),
	}
}
