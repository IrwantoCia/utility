// Package browse provides the SQLite table browser TUI with a 3-panel
// split-pane layout: Tables (left), Schema (middle), Details (right).
package browse

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/browse/details"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/browse/schema"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/browse/tables"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// BackToSqliteMenuMsg tells the SQLite coordinator to return to the menu.
type BackToSqliteMenuMsg struct{}

type panelFocus int

const (
	focusTables panelFocus = iota
	focusSchema
)

// Browse coordinates the 3-panel SQLite browser (Tables + Schema + Details).
type Browse struct {
	tables  *tables.Tables
	schema  *schema.Schema
	details *details.Details

	focus      panelFocus
	lastWindow tea.WindowSizeMsg
	keys       KeyMap
	helpModel  help.Model

	// Computed layout dimensions.
	leftW, midW, rightW int
	footerH             int
	innerH              int
}

var _ common.Component = (*Browse)(nil)

// Close implements common.Component.
func (b *Browse) Close() tea.Cmd { return nil }

// New creates a new Browse coordinator with placeholder data.
func New() *Browse {
	hm := help.New()
	hm.Styles = BrowseHelpStyles()
	return &Browse{
		tables:    tables.New(),
		schema:    schema.New(),
		details:   details.New(),
		focus:     focusTables,
		keys:      DefaultKeyMap,
		helpModel: hm,
	}
}

// Init sets the initial table selection on all panels.
func (b *Browse) Init() tea.Cmd {
	b.schema.SetTable("users")
	b.details.SetTable("users", 1204)
	return b.tables.Init()
}

// Resize computes the 3-column layout and forwards adjusted sizes to sub-panels.
func (b *Browse) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	b.lastWindow = ws
	b.helpModel, _ = b.helpModel.Update(ws)

	b.footerH = 1
	b.innerH = max(0, ws.Height-b.footerH-3) // 3 = top border + banner + bottom border

	b.leftW = ws.Width * 25 / 100
	b.midW = ws.Width * 45 / 100
	b.rightW = ws.Width - b.leftW - b.midW

	leftWS := tea.WindowSizeMsg{Width: b.leftW - 4, Height: b.innerH}

	return b.tables.Resize(leftWS)
}

// View renders the three bordered panels and help footer.
func (b *Browse) View() string {
	middle := b.renderPanels()
	helpStr := b.helpModel.View(b.keys)
	return lipgloss.JoinVertical(lipgloss.Left, middle, helpStr)
}

// renderPanels builds the three bordered panels joined horizontally.
func (b *Browse) renderPanels() string {
	tablesView  := b.wrapPanel(b.tables.View(), b.leftW, b.focus == focusTables, false, "Tables")
	schemaView  := b.wrapPanel(b.schema.View(b.midW-4), b.midW, b.focus == focusSchema, false, "Schema")
	detailsView := b.wrapPanel(b.details.View(b.rightW-4), b.rightW, false, false, "Details")
	return lipgloss.JoinHorizontal(lipgloss.Top, tablesView, schemaView, detailsView)
}

// wrapPanel surrounds content with a single bordered container with a title
// banner. When active, the panel uses bright cyan accent; when inactive,
// muted gray.
func (b *Browse) wrapPanel(content string, w int, active bool, filterActive bool, title string) string {
	innerW := w - 4

	bannerStyle := lipgloss.NewStyle()
	borderStyle := lipgloss.NewStyle()

	switch {
	case filterActive:
		borderStyle = style.Default.BrowseFilterBorder
		bannerStyle = lipgloss.NewStyle().
			Width(innerW).
			Background(lipgloss.Color("206")).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1)
	case active:
		borderStyle = style.Default.BrowseBorderActive
		bannerStyle = lipgloss.NewStyle().
			Width(innerW).
			Background(lipgloss.Color("45")).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1)
	default:
		borderStyle = style.Default.BrowseBorderInactive
		bannerStyle = lipgloss.NewStyle().
			Width(innerW).
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("250")).
			Padding(0, 1)
	}

	prefix := ""
	if active {
		prefix = "◈ "
	}
	banner := bannerStyle.Render(prefix + title)

	contentStyle := lipgloss.NewStyle().
		Width(innerW).
		Height(b.innerH)

	paddedContent := contentStyle.Render(content)

	inner := lipgloss.JoinVertical(lipgloss.Left, banner, paddedContent)

	return lipgloss.NewStyle().
		Width(w).
		Inherit(borderStyle).
		Padding(0, 1).
		Render(inner)
}

// Update handles input: Esc returns to SQLite menu, arrow/vim keys navigate,
// Left/Right/Tab switch panels, Enter refreshes schema/details.
func (b *Browse) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, b.keys.Esc) {
			return func() tea.Msg { return BackToSqliteMenuMsg{} }
		}
		if key.Matches(msg, b.keys.Up) {
			if b.focus == focusTables {
				b.tables.MoveUp()
				b.syncPanels()
			}
			return nil
		}
		if key.Matches(msg, b.keys.Down) {
			if b.focus == focusTables {
				b.tables.MoveDown()
				b.syncPanels()
			}
			return nil
		}
		if key.Matches(msg, b.keys.PgUp) {
			if b.focus == focusTables {
				b.tables.PageUp()
				b.syncPanels()
			}
			return nil
		}
		if key.Matches(msg, b.keys.PgDown) {
			if b.focus == focusTables {
				b.tables.PageDown()
				b.syncPanels()
			}
			return nil
		}
		if key.Matches(msg, b.keys.Left) {
			b.focus = focusTables
			b.syncPanels()
			return nil
		}
		if key.Matches(msg, b.keys.Right) {
			b.focus = focusSchema
			b.syncPanels()
			return nil
		}
		if key.Matches(msg, b.keys.Tab) {
			if b.focus == focusTables {
				b.focus = focusSchema
			} else {
				b.focus = focusTables
			}
			b.syncPanels()
			return nil
		}
		if key.Matches(msg, b.keys.Enter) {
			if b.focus == focusTables {
				b.syncPanels()
			}
			return nil
		}
		return nil

	default:
		var cmds []tea.Cmd
		if cmd := b.tables.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if len(cmds) > 0 {
			return tea.Batch(cmds...)
		}
	}

	return nil
}

// syncPanels updates schema and details panels from the currently selected
// table.
func (b *Browse) syncPanels() {
	selected := b.tables.Selected()
	if selected == "" {
		return
	}

	b.schema.SetTable(selected)

	rowCount := 0
	switch selected {
	case "users":
		rowCount = 1204
	case "orders":
		rowCount = 892
	default:
		rowCount = 0
	}
	b.details.SetTable(selected, rowCount)
}
