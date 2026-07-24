// Package tables provides the left panel of the SQLite browser, showing a
// scrollable, navigable list of table names.
package tables

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// Tables renders a scrollable list of placeholder table names in the left
// panel of the SQLite browser.
type Tables struct {
	items      []string
	cursor     int
	maxVisible int
	offset     int
}

var _ common.Component = (*Tables)(nil)

// Close implements common.Component.
func (t *Tables) Close() tea.Cmd { return nil }

// New creates a Tables panel pre-populated with placeholder table names.
func New() *Tables {
	return &Tables{
		items: []string{
			"users",
			"orders",
			"products",
			"categories",
			"inventory",
			"reviews",
			"payments",
			"sessions",
			"settings",
			"audit_log",
		},
	}
}

// Init is a no-op for the static placeholder list.
func (t *Tables) Init() tea.Cmd { return nil }

// Resize sets the maximum visible rows based on window height.
func (t *Tables) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	t.maxVisible = ws.Height - 2
	return nil
}

// View renders the table list with the cursor row highlighted,
// or an empty-state message when the list is empty.
func (t *Tables) View() string {
	if len(t.items) > 0 {
		var sb strings.Builder
		end := t.offset + t.maxVisible
		if end > len(t.items) {
			end = len(t.items)
		}
		for i := t.offset; i < end; i++ {
			if i > t.offset {
				sb.WriteByte('\n')
			}
			item := t.items[i]
			if i == t.cursor {
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

	return style.Default.BrowseEmpty.Render("No tables found")
}

// MoveUp moves the cursor up, scrolling the viewport if needed.
func (t *Tables) MoveUp() {
	if t.cursor > 0 {
		t.cursor--
	}
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
}

// MoveDown moves the cursor down, scrolling the viewport if needed.
func (t *Tables) MoveDown() {
	if t.cursor < len(t.items)-1 {
		t.cursor++
	}
	if t.cursor >= t.offset+t.maxVisible {
		t.offset = t.cursor - t.maxVisible + 1
	}
}

// PageUp moves the cursor up by half a page.
func (t *Tables) PageUp() {
	if t.maxVisible <= 0 {
		return
	}
	t.cursor -= t.maxVisible / 2
	if t.cursor < 0 {
		t.cursor = 0
	}
	t.offset = t.cursor
}

// PageDown moves the cursor down by half a page.
func (t *Tables) PageDown() {
	if t.maxVisible <= 0 || len(t.items) == 0 {
		return
	}
	t.cursor += t.maxVisible / 2
	if t.cursor >= len(t.items) {
		t.cursor = len(t.items) - 1
	}
	if t.cursor >= t.offset+t.maxVisible {
		t.offset = t.cursor - t.maxVisible + 1
	}
	if t.offset < 0 {
		t.offset = 0
	}
}

// Selected returns the currently selected table name, or "" if empty.
func (t *Tables) Selected() string {
	if len(t.items) == 0 || t.cursor >= len(t.items) {
		return ""
	}
	return t.items[t.cursor]
}

// Update is a no-op; key handling is done by the coordinator.
func (t *Tables) Update(msg tea.Msg) tea.Cmd { return nil }
