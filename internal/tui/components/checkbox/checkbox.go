// Package checkbox provides a TUI component for multi-select from a
// dynamic string list.
package checkbox

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
)

// Checkbox is a reusable TUI component for selecting multiple items from a
// dynamic list. Callers supply the items via SetItems before opening.
type Checkbox struct {
	items      []string
	checked    map[int]bool
	cursor     int
	Selected   []string // populated on Enter, caller should clear before re-open
	title      string
	styles     Styles
	keys       KeyMap
	helpModel  help.Model
	lastWindow tea.WindowSizeMsg
}

var _ common.Component = (*Checkbox)(nil)

// Close implements common.Component.
func (c *Checkbox) Close() tea.Cmd { return nil }

// New creates a new Checkbox with no items and no selection.
func New() *Checkbox {
	return &Checkbox{
		checked:   make(map[int]bool),
		keys:      DefaultKeyMap,
		helpModel: help.New(),
		title:     "Select",
		styles:    DefaultStyles(),
	}
}

// SetItems replaces the item list and resets checked map, cursor, and selection.
func (c *Checkbox) SetItems(items []string) {
	c.items = items
	c.checked = make(map[int]bool)
	c.cursor = 0
	c.Selected = nil
}

// SetTitle sets the title text shown above the item list.
func (c *Checkbox) SetTitle(t string) { c.title = t }

// SetChecked pre-checks items that match (by string equality). Useful when
// re-opening with previous selections.
func (c *Checkbox) SetChecked(items []string) {
	c.checked = make(map[int]bool)
	for _, s := range items {
		for i, item := range c.items {
			if item == s {
				c.checked[i] = true
			}
		}
	}
}

// GetChecked returns all currently checked items.
func (c *Checkbox) GetChecked() []string {
	var out []string
	for i, item := range c.items {
		if c.checked[i] {
			out = append(out, item)
		}
	}
	return out
}

// Init is a no-op.
func (c *Checkbox) Init() tea.Cmd { return nil }

// Resize stores window dimensions and initialises the help model.
func (c *Checkbox) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	c.lastWindow = ws
	c.helpModel, _ = c.helpModel.Update(ws)
	return nil
}

// View renders the modern floating-modal checkbox list with title, items,
// and help bar.
func (c *Checkbox) View() string {
	if len(c.items) == 0 {
		return ""
	}

	termW := c.lastWindow.Width
	termH := c.lastWindow.Height

	helpStr := c.helpModel.View(c.keys)
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
	n := len(c.items)
	start, end := 0, n
	if n > maxVisible {
		half := maxVisible / 2
		start = max(c.cursor-half, 0)
		end = start + maxVisible
		if end > n {
			end = n
			start = max(0, end-maxVisible)
		}
	}

	// Title bar
	innerW := boxW - 4 // 2 border + 2 padding
	titleText := c.styles.Title.Width(innerW).
		AlignHorizontal(lipgloss.Left).
		Render("❯ " + c.title)

	// Item rows
	var itemRows []string
	for i := start; i < end; i++ {
		prefix := "☐ "
		s := c.styles.Item
		if c.checked[i] {
			prefix = "☑ "
		}
		if i == c.cursor {
			prefix = "❯ " + prefix
			s = c.styles.Selected
		} else {
			prefix = "  " + prefix
		}
		line := s.Width(innerW).Render(prefix + c.items[i])
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

	box := c.styles.Box.Width(boxW).Render(inner)

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

// Update handles keyboard input for navigation and toggling.
func (c *Checkbox) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(keyMsg, c.keys.Up):
		if c.cursor > 0 {
			c.cursor--
		}
		return nil
	case key.Matches(keyMsg, c.keys.Down):
		if c.cursor < len(c.items)-1 {
			c.cursor++
		}
		return nil
	case key.Matches(keyMsg, c.keys.Space):
		if c.cursor >= 0 && c.cursor < len(c.items) {
			c.checked[c.cursor] = !c.checked[c.cursor]
		}
		return nil
	case key.Matches(keyMsg, c.keys.Enter):
		c.Selected = c.GetChecked()
		return nil
	}

	return nil
}
