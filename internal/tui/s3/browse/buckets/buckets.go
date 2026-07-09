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
	items  []string
	cursor int
}

var _ common.Component = (*Buckets)(nil)

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
}

// Init is a no-op for the static bucket list.
func (b *Buckets) Init() tea.Cmd { return nil }

// Resize is a no-op; height is managed by wrapPanel.
func (b *Buckets) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	return nil
}

// View renders the bucket list with the cursor row highlighted.
func (b *Buckets) View() string {
	var sb strings.Builder

	for i, item := range b.items {
		cursor := "  "
		if i == b.cursor {
			cursor = "▸ "
		}

		line := cursor + item
		if i == b.cursor {
			line = style.Default.Highlighted.Render(line)
		}
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	return sb.String()
}

// MoveUp moves the cursor up, wrapping at the top.
func (b *Buckets) MoveUp() {
	if b.cursor > 0 {
		b.cursor--
	}
}

// MoveDown moves the cursor down, wrapping at the bottom.
func (b *Buckets) MoveDown() {
	if b.cursor < len(b.items)-1 {
		b.cursor++
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
