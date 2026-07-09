// Package listpicker provides a TUI component for selecting an item from
// a dynamic string list.
package listpicker

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
)

// ListPicker is a reusable TUI component for selecting an item from a
// dynamic list. Callers supply the items via SetItems before opening.
type ListPicker struct {
	items      []string
	cursor     int
	Selected   string // set on Enter, caller should clear before re-open
	title      string
	styles     Styles
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
		styles:    DefaultStyles(),
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

// View renders the modern floating-modal list picker with title, items,
// and help bar.
func (l *ListPicker) View() string {
	if len(l.items) == 0 {
		return ""
	}

	termW := l.lastWindow.Width
	termH := l.lastWindow.Height

	helpStr := l.helpModel.View(l.keys)
	helpHeight := lipgloss.Height(helpStr)

	// Box width: at most 60% of terminal, at least 40, capped at 60
	boxW := min(max(termW*60/100, 40), 60)

	// Available rows for items inside box
	// title(1) + spacer(1) + items + help outside box
	// Box overhead: border(2) + padding top(1) + padding bottom(1) = 4
	// Content inside box: title(1) + spacer(1) + items
	// So: boxHeight = 4 + 2 + visibleItems
	// And: termH = boxHeight + helpHeight
	// → visibleItems = termH - helpHeight - 6
	maxVisible := termH - helpHeight - 6
	if maxVisible < 3 {
		maxVisible = 3
	}

	// Scroll logic (cursor-centered)
	n := len(l.items)
	start, end := 0, n
	if n > maxVisible {
		half := maxVisible / 2
		start = max(l.cursor-half, 0)
		end = start + maxVisible
		if end > n {
			end = n
			start = max(0, end-maxVisible)
		}
	}

	// Title bar
	innerW := boxW - 4 // 2 border + 2 padding
	titleText := l.styles.Title.Width(innerW).
		AlignHorizontal(lipgloss.Left).
		Render("❯ " + l.title)

	// Item rows
	var itemRows []string
	for i := start; i < end; i++ {
		prefix := "  "
		s := l.styles.Item
		if i == l.cursor {
			prefix = "❯ "
			s = l.styles.Selected
		}
		line := s.Width(innerW).Render(prefix + l.items[i])
		itemRows = append(itemRows, line)
	}

	// Pad to maxVisible so box height is consistent
	for len(itemRows) < maxVisible {
		itemRows = append(itemRows, "")
	}

	// Assemble inside the box
	inner := lipgloss.JoinVertical(lipgloss.Left,
		titleText,
		"",
		lipgloss.JoinVertical(lipgloss.Left, itemRows...),
	)

	box := l.styles.Box.Width(boxW).Render(inner)

	// Center box vertically above help bar
	availH := termH - helpHeight
	centered := lipgloss.NewStyle().
		Width(termW).
		Height(availH).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(box)

	return lipgloss.JoinVertical(lipgloss.Left, centered, helpStr)
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
