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
	items      []string
	cursor     int
	maxVisible int
	offset     int
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
	o.offset = 0
}

// Init is a no-op for the static object list.
func (o *Objects) Init() tea.Cmd { return nil }

// Resize sets the maximum visible rows based on window height.
func (o *Objects) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	o.maxVisible = ws.Height - 2
	return nil
}

// View renders the object list with the cursor row highlighted.
func (o *Objects) View() string {
	if len(o.items) > 0 {
		var sb strings.Builder
		end := o.offset + o.maxVisible
		if end > len(o.items) {
			end = len(o.items)
		}
		for i := o.offset; i < end; i++ {
			if i > o.offset {
				sb.WriteByte('\n')
			}
			item := o.items[i]
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
		}
		return sb.String()
	}

	return style.Default.CardDesc.Render("Select a bucket")
}

// Items returns the current object list (useful for the status bar).
func (o *Objects) Items() []string {
	return o.items
}

// MoveUp moves the cursor up, scrolling the viewport if needed.
func (o *Objects) MoveUp() {
	if o.cursor > 0 {
		o.cursor--
	}
	if o.cursor < o.offset {
		o.offset = o.cursor
	}
}

// MoveDown moves the cursor down, scrolling the viewport if needed.
func (o *Objects) MoveDown() {
	if o.cursor < len(o.items)-1 {
		o.cursor++
	}
	if o.cursor >= o.offset+o.maxVisible {
		o.offset = o.cursor - o.maxVisible + 1
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
