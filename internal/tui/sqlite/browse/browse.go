// Package browse provides the SQLite table browser TUI with a 4-panel
// split-pane layout: Tables (left), Filter+Data (middle), Info (right).
package browse

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	sqhelper "github.com/IrwantoCia/utility/internal/helper/sqlite"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/browse/data"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/browse/details"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/browse/filter"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/browse/tables"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// BackToSqliteMenuMsg tells the SQLite coordinator to return to the menu.
type BackToSqliteMenuMsg struct{}

type panelFocus int

const (
	focusTables panelFocus = iota
	focusFilter
	focusData
	focusDetails
)

// Browse coordinates the 4-panel SQLite browser (Tables + Filter + Data + Details).
type Browse struct {
	db      *sqhelper.DB
	tables  *tables.Tables
	filter  *filter.FilterPanel
	data    *data.Data
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
func (b *Browse) Close() tea.Cmd {
	b.db.Close()
	return nil
}

// New creates a new Browse coordinator with a live DB connection.
func New(db *sqhelper.DB) *Browse {
	hm := help.New()
	hm.Styles = BrowseHelpStyles()
	return &Browse{
		db:        db,
		tables:    tables.New(db),
		filter:    filter.New(db),
		data:      data.New(db),
		details:   details.New(db),
		focus:     focusTables,
		keys:      DefaultKeyMap,
		helpModel: hm,
	}
}

// Init loads table list and populates schema/details for the first table.
func (b *Browse) Init() tea.Cmd {
	b.tables.Init()
	b.syncPanels()
	return nil
}

// Resize computes the 3-column layout and forwards adjusted sizes to sub-panels.
// The middle column is split vertically into a fixed-height Filter panel (top)
// and a Data panel (bottom).
func (b *Browse) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	b.lastWindow = ws
	b.helpModel, _ = b.helpModel.Update(ws)

	b.footerH = 1
	b.innerH = max(0, ws.Height-b.footerH-3) // 3 = top border + banner + bottom border

	b.leftW = ws.Width * 15 / 100
	b.rightW = ws.Width * 25 / 100
	b.midW = ws.Width - b.leftW - b.rightW

	leftWS := tea.WindowSizeMsg{Width: b.leftW - 4, Height: b.innerH}

	// Split middle column: filter on top (fixed height), data below.
	filterH := filter.PanelHeight
	dataH := max(0, b.innerH-filterH-3)

	filterWS := tea.WindowSizeMsg{Width: b.midW - 4, Height: filterH}
	dataWS := tea.WindowSizeMsg{Width: b.midW - 4, Height: dataH}

	return tea.Batch(
		b.tables.Resize(leftWS),
		b.filter.Resize(filterWS),
		b.data.Resize(dataWS),
	)
}

// View renders the three bordered panels and help footer.
func (b *Browse) View() string {
	middle := b.renderPanels()
	helpStr := b.helpModel.View(b.keys)
	return lipgloss.JoinVertical(lipgloss.Left, middle, helpStr)
}

// renderPanels builds the four-panel layout joined horizontally: Tables (left),
// Filter+Data stacked (middle), and Info (right).
func (b *Browse) renderPanels() string {
	tablesView := b.wrapPanel(b.tables.View(), b.leftW, b.focus == focusTables, false, "Tables")

	// Middle column: filter on top (fixed height), data below.
	filterH := filter.PanelHeight
	dataH := b.innerH - filterH - 3

	filterView := b.renderSubPanel(b.filter.View(), b.midW, filterH, b.focus == focusFilter, "Filter")
	dataView := b.renderSubPanel(b.data.View(), b.midW, dataH, b.focus == focusData, "Data")
	middle := lipgloss.JoinVertical(lipgloss.Top, filterView, dataView)

	detailsView := b.wrapPanel(b.details.View(b.rightW-4), b.rightW, b.focus == focusDetails, false, "Info")
	return lipgloss.JoinHorizontal(lipgloss.Top, tablesView, middle, detailsView)
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

// renderSubPanel wraps a sub-panel (filter or data) inside the middle column,
// with its own banner and content area sized to the given height.
func (b *Browse) renderSubPanel(content string, w, h int, active bool, title string) string {
	innerW := w - 4

	bannerStyle := lipgloss.NewStyle()
	borderStyle := lipgloss.NewStyle()

	if active {
		borderStyle = style.Default.BrowseBorderActive
		bannerStyle = lipgloss.NewStyle().
			Width(innerW).
			Background(lipgloss.Color("45")).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1)
	} else {
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
		Height(h)
	paddedContent := contentStyle.Render(content)

	inner := lipgloss.JoinVertical(lipgloss.Left, banner, paddedContent)

	return lipgloss.NewStyle().
		Width(w).
		Inherit(borderStyle).
		Padding(0, 1).
		Render(inner)
}

// focusRight implements the custom right-arrow jump. From Tables → Data,
// Filter → Info, Data → Info, Info → Info (no-op).
func (b *Browse) focusRight() panelFocus {
	switch b.focus {
	case focusTables:
		return focusData
	case focusFilter:
		return focusDetails
	case focusData:
		return focusDetails
	case focusDetails:
		return focusDetails
	default:
		return focusTables
	}
}

// focusLeft implements the custom left-arrow jump. From Tables → Tables (no-op),
// Filter → Tables, Data → Tables, Info → Data.
func (b *Browse) focusLeft() panelFocus {
	switch b.focus {
	case focusTables:
		return focusTables
	case focusFilter:
		return focusTables
	case focusData:
		return focusTables
	case focusDetails:
		return focusData
	default:
		return focusTables
	}
}

// Update handles input: delegates to the active panel first, then processes
// global keys (focus switching, Esc to return to menu).
func (b *Browse) Update(msg tea.Msg) tea.Cmd {
	// Delegate resize to all panels.
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		return b.Resize(ws)
	}

	// ── FocusDataMsg: filter panel wants focus on Data panel ──────
	if _, ok := msg.(filter.FocusDataMsg); ok {
		b.focus = focusData
		b.syncPanels()
		return nil
	}

	keyMsg, isKey := msg.(tea.KeyPressMsg)

	// ── Active panel delegation ────────────────────────────────────
	// The filter panel handles keys internally (picker, text input).
	if b.focus == focusFilter && isKey {
		if cmd := b.filter.Update(msg); cmd != nil {
			b.syncPanels()
			return cmd
		}
		// If filter didn't consume the key, fall through to global keys.
	}
	if b.focus == focusData && isKey {
		switch {
		case key.Matches(keyMsg, b.keys.Up):
			b.data.MoveUp()
			return nil
		case key.Matches(keyMsg, b.keys.Down):
			b.data.MoveDown()
			return nil
		case key.Matches(keyMsg, b.keys.PgUp):
			b.data.PageUp()
			return nil
		case key.Matches(keyMsg, b.keys.PgDown):
			b.data.PageDown()
			return nil
		}
	}

	// ── Global key handling ────────────────────────────────────────
	if isKey {
		if key.Matches(keyMsg, b.keys.Esc) {
			return func() tea.Msg { return BackToSqliteMenuMsg{} }
		}
		if key.Matches(keyMsg, b.keys.Filter) {
			b.focus = focusFilter
			b.syncPanels()
			return nil
		}
		if key.Matches(keyMsg, b.keys.Right) {
			b.focus = b.focusRight()
			b.syncPanels()
			return nil
		}
		if key.Matches(keyMsg, b.keys.Left) {
			b.focus = b.focusLeft()
			b.syncPanels()
			return nil
		}

		// Panel-specific movement (not filter — it handles its own).
		switch b.focus {
		case focusTables:
			switch {
			case key.Matches(keyMsg, b.keys.Up):
				b.tables.MoveUp()
				b.syncPanels()
			case key.Matches(keyMsg, b.keys.Down):
				b.tables.MoveDown()
				b.syncPanels()
			case key.Matches(keyMsg, b.keys.PgUp):
				b.tables.PageUp()
				b.syncPanels()
			case key.Matches(keyMsg, b.keys.PgDown):
				b.tables.PageDown()
				b.syncPanels()
			case key.Matches(keyMsg, b.keys.Enter):
				b.syncPanels()
			}
		case focusFilter:
			// Filter handles keys via delegation above; this is a no-op for
			// global keys that filter's Update returned nil for (e.g. unknown keys).
			return nil
		case focusData:
			// Data movement is handled above.
		case focusDetails:
			// Details is read-only; no movement keys.
		}

		return nil
	}

	// ── Non-key messages (delegate to all panels) ─────────────────
	var cmds []tea.Cmd
	for _, comp := range []common.Component{b.tables, b.filter, b.data} {
		if cmd := comp.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}

	return nil
}

// syncPanels updates data, filter, and details panels from the currently
// selected table. It passes the current filter state to the data panel and
// updates the filter panel's column list.
func (b *Browse) syncPanels() {
	selected := b.tables.Selected()
	if selected == "" {
		return
	}

	filters := b.filter.Filters()
	b.data.SetTable(selected, filters)
	b.details.SetTable(selected)

	cols, err := b.db.Columns(selected)
	if err == nil {
		b.filter.SetColumns(cols)
	}
}
