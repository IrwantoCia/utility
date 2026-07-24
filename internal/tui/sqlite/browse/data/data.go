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
// selected table in a tabular format, with vertical scrolling via viewport.
type Data struct {
	db        *sqlite.DB
	tableName string
	rows      [][]string
	colNames  []string
	errMsg    string
	viewport  viewport.Model
}

var _ common.Component = (*Data)(nil)

// New creates a Data panel backed by a live DB connection.
func New(db *sqlite.DB) *Data {
	return &Data{
		db:       db,
		viewport: viewport.New(),
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

// View renders the data preview through the scrollable viewport.
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
			if row%2 == 0 {
				return style.Default.TableRowAlt
			}
			return lipgloss.NewStyle()
		})

	result := t.String()

	footer := style.Default.BrowseMetaDim.Render(
		fmt.Sprintf("\n%d row(s) shown (limit %d)", len(d.rows), defaultLimit),
	)

	d.viewport.SetContent(result + footer)
	return d.viewport.View()
}

// Update delegates to the viewport model for built-in key handling.
func (d *Data) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return cmd
}

// ScrollUp scrolls the viewport up by one line.
func (d *Data) ScrollUp() {
	d.viewport.ScrollUp(1)
}

// ScrollDown scrolls the viewport down by one line.
func (d *Data) ScrollDown() {
	d.viewport.ScrollDown(1)
}

// PageUp scrolls the viewport up by one page.
func (d *Data) PageUp() {
	d.viewport.PageUp()
}

// PageDown scrolls the viewport down by one page.
func (d *Data) PageDown() {
	d.viewport.PageDown()
}

// SetTable queries the selected table's data and stores rows for display.
func (d *Data) SetTable(name string) {
	d.tableName = name
	rows, colNames, err := d.db.Query(name, defaultLimit, 0)
	if err != nil {
		d.errMsg = fmt.Sprintf("Error: %v", err)
		d.rows = nil
		d.colNames = nil
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
}
