// Package data provides the middle panel of the SQLite browser, showing
// query results (data preview) for the selected table — no tab switching.
package data

import (
	"fmt"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/IrwantoCia/utility/internal/helper/sqlite"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// defaultLimit is the number of rows fetched per table query.
const defaultLimit = 100

// Data renders the middle panel showing actual data rows from the currently
// selected table in a tabular format, with vertical scrolling via viewport
// and cursor-based row highlighting.
type Data struct {
	db        *sqlite.DB
	tableName string
	filters   []sqlite.Filter
	rows      [][]string
	colNames  []string
	errMsg    string
	cursor    int
	selectedStyle lipgloss.Style
	viewport  viewport.Model
}

var _ common.Component = (*Data)(nil)

// New creates a Data panel backed by a live DB connection.
func New(db *sqlite.DB) *Data {
	return &Data{
		db:            db,
		cursor:        0,
		selectedStyle: style.Default.RowHighlighted,
		viewport:      viewport.New(),
	}
}

// Init implements common.Component.
func (d *Data) Init() tea.Cmd { return nil }

// Close implements common.Component.
func (d *Data) Close() tea.Cmd { return nil }

// Resize sets the viewport dimensions from the window-size message.
func (d *Data) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	d.viewport.SetWidth(ws.Width)
	d.viewport.SetHeight(ws.Height)
	return nil
}

// View renders the data preview through the scrollable viewport with cursor
// highlighting and auto-scroll.
func (d *Data) View() string {
	if d.errMsg != "" {
		d.viewport.SetContent(style.Default.BrowseEmpty.Render(d.errMsg))
		return d.viewport.View()
	}

	if d.tableName == "" {
		d.viewport.SetContent(style.Default.BrowseEmpty.Render("Select a table to view data"))
		return d.viewport.View()
	}

	t := table.New().
		Headers(d.colNames...).
		Rows(d.rows...).
		Width(d.viewport.Width()).
		Border(lipgloss.NormalBorder()).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return style.Default.TableHeader
			}
			if row == d.cursor {
				return d.selectedStyle.MaxHeight(1)
			}
			if row%2 == 0 {
				return style.Default.TableRowAlt.MaxHeight(1)
			}
			return lipgloss.NewStyle().MaxHeight(1)
		})

	result := t.String()

	footer := style.Default.BrowseMetaDim.Render(
		fmt.Sprintf("\n%d row(s) shown (limit %d)", len(d.rows), defaultLimit),
	)

	content := result + footer
	d.viewport.SetContent(content)
	d.viewport.SetYOffset(d.cursor)
	return d.viewport.View()
}

// Update delegates to the viewport model for built-in key handling.
func (d *Data) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return cmd
}

// MoveUp decrements the cursor (clamped to 0) and scrolls the viewport.
func (d *Data) MoveUp() {
	if d.cursor > 0 {
		d.cursor--
		d.viewport.SetYOffset(d.cursor)
	}
}

// MoveDown increments the cursor (clamped to max row) and scrolls the viewport.
func (d *Data) MoveDown() {
	if d.cursor < len(d.rows)-1 {
		d.cursor++
		d.viewport.SetYOffset(d.cursor)
	}
}

// PageUp moves the cursor up by one viewport height (clamped to 0).
func (d *Data) PageUp() {
	d.cursor = max(0, d.cursor-d.viewport.Height())
	d.viewport.SetYOffset(d.cursor)
}

// PageDown moves the cursor down by one viewport height (clamped to max row).
func (d *Data) PageDown() {
	maxRow := max(0, len(d.rows)-1)
	d.cursor = min(maxRow, d.cursor+d.viewport.Height())
	d.viewport.SetYOffset(d.cursor)
}

// SetTable queries the selected table's data with optional filters and stores
// rows for display. Resets the cursor to the first row.
func (d *Data) SetTable(name string, filters []sqlite.Filter) {
	d.tableName = name
	d.filters = filters
	rows, colNames, err := d.db.QueryFiltered(name, filters, defaultLimit, 0)
	if err != nil {
		d.errMsg = fmt.Sprintf("Error: %v", err)
		d.rows = nil
		d.colNames = nil
		d.cursor = 0
		return
	}
	d.errMsg = ""
	d.colNames = colNames
	d.rows = make([][]string, len(rows))
	for i, row := range rows {
		strRow := make([]string, len(row))
		for j, val := range row {
			strRow[j] = fmt.Sprintf("%v", val)
		}
		d.rows[i] = strRow
	}
	d.cursor = 0
}
