// Package buckets provides the left panel of the S3 browser, showing a
// navigable list of S3 bucket names.
package buckets

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// Buckets renders a navigable list of S3 buckets.
type Buckets struct {
	items      []string
	cursor     int
	status     string
	maxVisible int
	offset     int
}

var _ common.Component = (*Buckets)(nil)

// Close implements common.Component.
func (b *Buckets) Close() tea.Cmd { return nil }

// New creates an empty Buckets panel.
func New() *Buckets {
	return &Buckets{
		items: []string{},
	}
}

// SetItems replaces the bucket list and resets the cursor to the top.
func (b *Buckets) SetItems(items []string) {
	b.items = items
	b.cursor = 0
	b.offset = 0
	b.status = ""
}

// SetStatus sets a status message (shown when items list is empty).
func (b *Buckets) SetStatus(s string) {
	b.status = s
}

// Init is a no-op for the static bucket list.
func (b *Buckets) Init() tea.Cmd { return nil }

// Resize sets the maximum visible rows based on window height.
func (b *Buckets) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	b.maxVisible = ws.Height - 2
	return nil
}

// View renders the bucket list with the cursor row highlighted,
// or a status/loading message when the list is empty.
func (b *Buckets) View() string {
	if len(b.items) > 0 {
		var sb strings.Builder
		end := b.offset + b.maxVisible
		if end > len(b.items) {
			end = len(b.items)
		}
		for i := b.offset; i < end; i++ {
			if i > b.offset {
				sb.WriteByte('\n')
			}
			item := b.items[i]
			if i == b.cursor {
				cursor := style.Default.BrowseListCursor.Render("❯ ")
				name := style.Default.BrowseListSelected.Render(item)
				sb.WriteString(cursor + name)
			} else {
				prefix := style.Default.BrowseListNormal.Render("  ▪ ")
				name := style.Default.BrowseListNormal.Render(item)
				sb.WriteString(prefix + name)
			}
		}
		return sb.String()
	}

	msg := b.status
	if msg == "" {
		msg = "Loading…"
	}
	return style.Default.BrowseEmpty.Render(msg)
}

// PageUp moves the cursor up by one page worth of items.
func (b *Buckets) PageUp() {
	if b.maxVisible <= 0 {
		return
	}
	b.cursor -= b.maxVisible / 2
	if b.cursor < 0 {
		b.cursor = 0
	}
	b.offset = b.cursor
}

// PageDown moves the cursor down by one page worth of items.
func (b *Buckets) PageDown() {
	if b.maxVisible <= 0 || len(b.items) == 0 {
		return
	}
	b.cursor += b.maxVisible / 2
	if b.cursor >= len(b.items) {
		b.cursor = len(b.items) - 1
	}
	// Scroll viewport so cursor stays visible
	if b.cursor >= b.offset+b.maxVisible {
		b.offset = b.cursor - b.maxVisible + 1
	}
	if b.offset < 0 {
		b.offset = 0
	}
}

// MoveUp moves the cursor up, scrolling the viewport if needed.
func (b *Buckets) MoveUp() {
	if b.cursor > 0 {
		b.cursor--
	}
	if b.cursor < b.offset {
		b.offset = b.cursor
	}
}

// MoveDown moves the cursor down, scrolling the viewport if needed.
func (b *Buckets) MoveDown() {
	if b.cursor < len(b.items)-1 {
		b.cursor++
	}
	if b.cursor >= b.offset+b.maxVisible {
		b.offset = b.cursor - b.maxVisible + 1
	}
}

// Selected returns the currently selected bucket name, or "" if empty.
func (b *Buckets) Selected() string {
	if len(b.items) == 0 || b.cursor >= len(b.items) {
		return ""
	}
	return b.items[b.cursor]
}

// Update is a no-op; key handling is done by the coordinator.
func (b *Buckets) Update(msg tea.Msg) tea.Cmd { return nil }
