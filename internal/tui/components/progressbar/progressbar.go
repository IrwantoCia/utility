// Package progressbar provides a simple progress bar display component.
// It wraps bubbles/v2/progress.Model with a clean API.
//
// This component is NOT a common.Component — it has no lifecycle methods.
// Callers set the percent and width and call View in their own View method.
package progressbar

import (
	"charm.land/bubbles/v2/progress"
)

// ProgressBar wraps a progress.Model for simple rendering.
type ProgressBar struct {
	model   progress.Model
	percent float64 // 0.0 to 100.0
	styles  Styles
}

// New creates a ProgressBar with default gradient styling.
func New() *ProgressBar {
	s := DefaultStyles()
	return &ProgressBar{
		model:  progress.New(progress.WithScaled(true), progress.WithColors(s.Filled, s.Empty)),
		styles: s,
	}
}

// SetPercent sets the progress value (0.0 to 100.0).
func (p *ProgressBar) SetPercent(percent float64) {
	p.percent = percent
}

// SetWidth sets the width of the progress bar in characters.
func (p *ProgressBar) SetWidth(width int) {
	p.model.SetWidth(width)
}

// Width returns the current width.
func (p *ProgressBar) Width() int {
	return p.model.Width()
}

// SetStyles allows customizing colors. Call before first render.
func (p *ProgressBar) SetStyles(s Styles) {
	p.styles = s
	p.model = progress.New(progress.WithScaled(true), progress.WithColors(s.Filled, s.Empty))
}

// View renders the progress bar. Returns empty string if percent is 0.
func (p *ProgressBar) View() string {
	if p.percent <= 0 {
		return ""
	}
	// ViewAs takes 0.0 to 1.0
	return p.model.ViewAs(p.percent / 100.0)
}
