// Package objects provides the middle panel of the S3 browser, displaying
// a navigable list of S3 objects for the selected bucket.
package objects

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

var (
	navUp   = key.NewBinding(key.WithKeys("k", "up"))
	navDown = key.NewBinding(key.WithKeys("j", "down"))
)

// Objects renders a navigable list of S3 object keys.
type Objects struct {
	items      []string
	cursor     int
	lastWindow tea.WindowSizeMsg
}

var _ common.Component = (*Objects)(nil)

// New creates an Objects panel with hardcoded demo data.
func New() *Objects {
	return &Objects{
		items: []string{
			"report.csv",
			"data.json",
			"logs/",
			"config.yaml",
		},
	}
}

// Init is a no-op for the static object list.
func (o *Objects) Init() tea.Cmd { return nil }

// Resize stores the window dimensions forwarded by the S3 coordinator.
func (o *Objects) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	o.lastWindow = ws
	return nil
}

// View renders the object list with the cursor row highlighted.
func (o *Objects) View() string {
	var sb strings.Builder

	for i, item := range o.items {
		cursor := "  "
		if i == o.cursor {
			cursor = "▸ "
		}

		icon := "\U0001F4C4 " // 📄 file
		if strings.HasSuffix(item, "/") {
			icon = "\U0001F4C1 " // 📁 folder
		}

		line := cursor + icon + item
		if i == o.cursor {
			line = style.Default.Highlighted.Render(line)
		}
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	// Pad to fill the panel height.
	targetH := o.lastWindow.Height
	if targetH > 0 {
		currentH := lipgloss.Height(sb.String())
		for j := currentH; j < targetH; j++ {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// Items returns the current object list (useful for the status bar).
func (o *Objects) Items() []string {
	return o.items
}

// Update handles keyboard navigation: k/up and j/down.
func (o *Objects) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(keyMsg, navUp):
		if o.cursor > 0 {
			o.cursor--
		}
	case key.Matches(keyMsg, navDown):
		if o.cursor < len(o.items)-1 {
			o.cursor++
		}
	}

	return nil
}
