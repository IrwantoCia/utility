// Package listpicker provides a TUI component for selecting an item from
// a dynamic string list.
package listpicker

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// ListPicker is a reusable TUI component for selecting an item from a
// dynamic list. Callers supply the items via SetItems before opening.
type ListPicker struct {
	items      []string
	cursor     int
	Selected   string // set on Enter, caller should clear before re-open
	title      string
	keys       KeyMap
	helpModel  help.Model
	lastWindow tea.WindowSizeMsg
}

var _ common.Component = (*ListPicker)(nil)

// New creates a new ListPicker with no items and no selection.
func New() *ListPicker {
	return &ListPicker{
		keys:      DefaultKeyMap,
		helpModel: help.New(),
		title:     "Select",
	}
}

// SetItems replaces the item list and resets cursor and selection.
func (l *ListPicker) SetItems(items []string) {
	l.items = items
	l.cursor = 0
	l.Selected = ""
}

// SetTitle sets the title text shown above the item list.
func (l *ListPicker) SetTitle(t string) { l.title = t }

// Init is a no-op.
func (l *ListPicker) Init() tea.Cmd { return nil }

// Resize stores window dimensions and initialises the help model.
func (l *ListPicker) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	l.lastWindow = ws
	l.helpModel, _ = l.helpModel.Update(ws)
	return nil
}

// View renders the full-screen list picker with title, items, and help bar.
func (l *ListPicker) View() string {
	h := l.lastWindow.Height

	helpStr := l.helpModel.View(l.keys)
	helpHeight := lipgloss.Height(helpStr)

	// Available rows for content
	available := h - helpHeight

	// Title
	titleLine := style.Default.MenuTitle.
		Width(l.lastWindow.Width).
		Render(l.title)

	var rows []string
	rows = append(rows, "")
	rows = append(rows, titleLine)
	rows = append(rows, "")

	// Visible range (cursor-centred scrolling)
	n := len(l.items)
	maxVisible := max(3, available-5)
	start, end := 0, n
	if n > maxVisible {
		half := maxVisible / 2
		start = l.cursor - half
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > n {
			end = n
			start = max(0, end-maxVisible)
		}
	}

	// Render visible items
	itemWidth := l.lastWindow.Width - 4
	for i := start; i < end; i++ {
		item := l.items[i]
		s := style.Default.MenuItem
		if i == l.cursor {
			s = style.Default.MenuItemSelected
		}
		itemStr := s.Width(itemWidth).Render("  " + item)
		rows = append(rows, itemStr)
	}

	// Join and pad to fill available height
	padded := lipgloss.JoinVertical(lipgloss.Left, rows...)
	contentHeight := lipgloss.Height(padded)
	var b strings.Builder
	b.WriteString(padded)
	for i := contentHeight; i < available; i++ {
		b.WriteRune('\n')
	}
	b.WriteString(helpStr)
	return b.String()
}

// Update handles keyboard input for navigation and selection.
func (l *ListPicker) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(keyMsg, l.keys.Up):
		if l.cursor > 0 {
			l.cursor--
		}
		return nil
	case key.Matches(keyMsg, l.keys.Down):
		if l.cursor < len(l.items)-1 {
			l.cursor++
		}
		return nil
	case key.Matches(keyMsg, l.keys.Enter):
		if l.cursor >= 0 && l.cursor < len(l.items) {
			l.Selected = l.items[l.cursor]
		}
		return nil
	}

	return nil
}
