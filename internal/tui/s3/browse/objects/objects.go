// Package objects provides the middle panel of the S3 browser, displaying
// a navigable list of S3 objects for the selected bucket.
package objects

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// Objects renders a navigable list of S3 object keys.
type Objects struct {
	items        []string
	allItems     []string // original unfiltered list from S3
	filter       string   // current filter query
	filterActive bool     // true when editing filter (typing mode)
	cursor       int
	maxVisible   int
	offset       int
	maxWidth     int // available content width from Resize
}

var _ common.Component = (*Objects)(nil)

// Close implements common.Component.
func (o *Objects) Close() tea.Cmd { return nil }

// New creates an empty Objects panel.
func New() *Objects {
	return &Objects{
		items: []string{},
	}
}

// SetItems replaces the object list and resets the cursor to the top.
func (o *Objects) SetItems(items []string) {
	o.allItems = items
	o.items = items
	o.cursor = 0
	o.offset = 0
}

// EnterFilter activates filter editing mode.
func (o *Objects) EnterFilter() {
	o.filterActive = true
	o.filter = ""
	o.cursor = 0
	o.offset = 0
}

// ExitFilterMode exits filter editing but keeps the filter and filtered items applied.
func (o *Objects) ExitFilterMode() {
	o.filterActive = false
}

// AppendFilter adds a rune to the filter query and re-filters.
func (o *Objects) AppendFilter(r rune) {
	o.filter += string(r)
	o.applyFilter()
}

// DeleteFilter removes the last rune from the filter query and re-filters.
// Handle multi-byte runes: convert to []rune, slice, convert back.
func (o *Objects) DeleteFilter() {
	runes := []rune(o.filter)
	if len(runes) == 0 {
		return
	}
	o.filter = string(runes[:len(runes)-1])
	o.applyFilter()
}

// ClearFilter clears the filter and restores all items.
func (o *Objects) ClearFilter() {
	o.filterActive = false
	o.filter = ""
	o.items = o.allItems
	o.cursor = 0
	o.offset = 0
}

// Filter returns the current filter query string.
func (o *Objects) Filter() string { return o.filter }

// FilterActive returns whether filter editing mode is active.
func (o *Objects) FilterActive() bool { return o.filterActive }

// TotalItems returns the count of all (unfiltered) items.
func (o *Objects) TotalItems() int { return len(o.allItems) }

// applyFilter rebuilds o.items by case-insensitive substring matching against o.allItems.
// If filter is empty, restores o.items = o.allItems.
// Resets cursor and offset to 0.
func (o *Objects) applyFilter() {
	if o.filter == "" {
		o.items = o.allItems
		o.cursor = 0
		o.offset = 0
		return
	}
	q := strings.ToLower(o.filter)
	filtered := make([]string, 0)
	for _, item := range o.allItems {
		if strings.Contains(strings.ToLower(item), q) {
			filtered = append(filtered, item)
		}
	}
	o.items = filtered
	o.cursor = 0
	o.offset = 0
}

// Init is a no-op for the static object list.
func (o *Objects) Init() tea.Cmd { return nil }

// Resize sets the maximum visible rows based on window height.
func (o *Objects) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	o.maxVisible = ws.Height - 2
	o.maxWidth = ws.Width
	return nil
}

// View renders the object list with the cursor row highlighted.
func (o *Objects) View() string {
	if len(o.items) > 0 {
		var sb strings.Builder

		// Filter bar when filter is active
		if o.filterActive {
			filterPrefix := style.Default.BrowseListCursor.Render("🔍")
			filterText := style.Default.BrowseListSelected.Render(" " + o.filter + " ")
			filterCount := style.Default.BrowseListNormal.Render(
				fmt.Sprintf("  %d/%d", len(o.items), len(o.allItems)),
			)
			filterBar := filterPrefix + filterText + filterCount
			sb.WriteString(filterBar)
			sb.WriteByte('\n')

			// Separator line
			sep := lipgloss.NewStyle().
				Width(o.maxWidth).
				Foreground(lipgloss.Color("206")).
				Render(strings.Repeat("─", o.maxWidth))
			sb.WriteString(sep)
			sb.WriteByte('\n')
		}

		end := o.offset + o.maxVisible
		if o.filterActive {
			// Account for the filter bar + separator (2 lines)
			end = o.offset + o.maxVisible - 2
			if end < o.offset {
				end = o.offset
			}
		}
		if end > len(o.items) {
			end = len(o.items)
		}
		for i := o.offset; i < end; i++ {
			if i > o.offset {
				sb.WriteByte('\n')
			}
			item := o.items[i]

			icon := "📄 "
			if strings.HasSuffix(item, "/") {
				icon = "📁 "
			}

			if i == o.cursor {
				cursorStr := "❯ "
				iconStr := icon
				displayName := truncate(item, o.maxWidth-len("  ")-len(cursorStr)-len(iconStr))
				line := "  " + style.Default.BrowseListCursor.Render(cursorStr) + style.Default.BrowseListSelected.Render(iconStr+displayName)
				sb.WriteString(line)
			} else {
				iconStr := icon
				displayName := truncate(item, o.maxWidth-len("  ")-len(iconStr))
				line := "  " + style.Default.BrowseListNormal.Render(iconStr + displayName)
				sb.WriteString(line)
			}
		}
		return sb.String()
	}

	return style.Default.BrowseEmpty.Render("Select a bucket")
}

// truncate shortens s to fit within maxWidth visual columns, appending "…"
// when truncated. Uses len(s) as a proxy for visual width (accurate for ASCII).
func truncate(s string, maxWidth int) string {
	if len(s) <= maxWidth {
		return s
	}
	// Leave room for "…" (3 bytes in UTF-8, 1 visual column)
	limit := maxWidth - 1
	if limit <= 0 {
		return "…"
	}
	// Cut at byte boundary — safe for ASCII, may break multi-byte runes.
	// For filenames this is acceptable (keys are typically ASCII).
	if limit > len(s) {
		limit = len(s)
	}
	return s[:limit] + "…"
}

// Items returns the current object list (useful for the status bar).
func (o *Objects) Items() []string {
	return o.items
}

// PageUp moves the cursor up by one page worth of items.
func (o *Objects) PageUp() {
	if o.maxVisible <= 0 {
		return
	}
	o.cursor -= o.maxVisible / 2
	if o.cursor < 0 {
		o.cursor = 0
	}
	o.offset = o.cursor
}

// PageDown moves the cursor down by one page worth of items.
func (o *Objects) PageDown() {
	if o.maxVisible <= 0 || len(o.items) == 0 {
		return
	}
	o.cursor += o.maxVisible / 2
	if o.cursor >= len(o.items) {
		o.cursor = len(o.items) - 1
	}
	// Scroll viewport so cursor stays visible
	if o.cursor >= o.offset+o.maxVisible {
		o.offset = o.cursor - o.maxVisible + 1
	}
	if o.offset < 0 {
		o.offset = 0
	}
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
