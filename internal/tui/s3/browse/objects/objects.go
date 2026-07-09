// Package objects provides the middle panel of the S3 browser, displaying
// a navigable list of S3 objects for the selected bucket.
package objects

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// Objects renders a navigable list of S3 object keys.
type Objects struct {
	items  []string
	cursor int
}

var _ common.Component = (*Objects)(nil)

// New creates an empty Objects panel.
func New() *Objects {
	return &Objects{
		items: []string{},
	}
}

// SetItems replaces the object list and resets the cursor to the top.
func (o *Objects) SetItems(items []string) {
	o.items = items
	o.cursor = 0
}

// Init is a no-op for the static object list.
func (o *Objects) Init() tea.Cmd { return nil }

// Resize is a no-op; height is managed by wrapPanel.
func (o *Objects) Resize(ws tea.WindowSizeMsg) tea.Cmd {
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

	return sb.String()
}

// Items returns the current object list (useful for the status bar).
func (o *Objects) Items() []string {
	return o.items
}

// MoveUp moves the cursor up, wrapping at the top.
func (o *Objects) MoveUp() {
	if o.cursor > 0 {
		o.cursor--
	}
}

// MoveDown moves the cursor down, wrapping at the bottom.
func (o *Objects) MoveDown() {
	if o.cursor < len(o.items)-1 {
		o.cursor++
	}
}

// Selected returns the currently selected object key, or "" if empty.
func (o *Objects) Selected() string {
	if len(o.items) == 0 || o.cursor >= len(o.items) {
		return ""
	}
	return o.items[o.cursor]
}

// Update is a no-op; key handling is done by the coordinator.
func (o *Objects) Update(msg tea.Msg) tea.Cmd { return nil }
