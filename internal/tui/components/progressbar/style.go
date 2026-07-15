package progressbar

import (
	"charm.land/lipgloss/v2"
	"image/color"
)

// Styles holds the visual styles for the ProgressBar component.
// Filled is the gradient start (bright) color, Empty is the gradient end (dim) color.
type Styles struct {
	Filled color.Color
	Empty  color.Color
}

// DefaultStyles returns the default ProgressBar styles.
func DefaultStyles() Styles {
	return Styles{
		Filled: lipgloss.Color("#FF7CCB"), // pink
		Empty:  lipgloss.Color("#FDFF8C"), // yellow
	}
}
