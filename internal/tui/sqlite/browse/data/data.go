// Package data provides the middle panel of the SQLite browser, showing
// query results (data preview) for the selected table — no tab switching.
package data

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/IrwantoCia/utility/internal/helper/sqlite"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// RowChangedMsg is emitted when the cursor (highlighted row) changes, carrying
// the full column names and row values so the coordinator can propagate them
// to the Info (details) panel.
type RowChangedMsg struct {
	ColNames []string
	Values   []string
}

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
//
// Each cell is truncated to a computed character cap so no cell can wrap to
// multiple visual lines — guaranteeing 1-line-per-row invariant required by
// cursor-based YOffset scrolling.
func (d *Data) View() string {
	if d.errMsg != "" {
		d.viewport.SetContent(style.Default.BrowseEmpty.Render(d.errMsg))
		return d.viewport.View()
	}

	if d.tableName == "" {
		d.viewport.SetContent(style.Default.BrowseEmpty.Render("Select a table to view data"))
		return d.viewport.View()
	}

	width := d.viewport.Width()

	// Compute per-cell character cap to prevent word-wrap.
	// Total width = (colCount-1) column separators + 2*colCount cell padding + content.
	// Assign the average content width per column, clamped to a readable minimum.
	colCount := len(d.colNames)
	if colCount == 0 {
		colCount = 1
	}
	maxCell := max(8, width/colCount-3)

	// Build truncated copy of rows so no cell text can exceed its column width.
	truncatedRows := make([][]string, len(d.rows))
	for i, row := range d.rows {
		tr := make([]string, len(row))
		for j, cell := range row {
			tr[j] = truncateCell(cell, maxCell)
		}
		truncatedRows[i] = tr
	}

	t := table.New().
		Headers(d.colNames...).
		Rows(truncatedRows...).
		Width(width).
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
			return style.Default.TableRow.MaxHeight(1)
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

// truncateCell guarantees s is at most maxLen Unicode code points, with an
// ellipsis (…) appended when the string exceeds the limit. This prevents
// lipgloss word-wrapping within table cells.
func truncateCell(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	// Fast ASCII path: no allocation for short strings.
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

// Update delegates to the viewport model for built-in key handling.
func (d *Data) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return cmd
}

// Rows returns the full (untruncated) data rows.
func (d *Data) Rows() [][]string { return d.rows }

// ColNames returns the column names for the current table.
func (d *Data) ColNames() []string { return d.colNames }

// rowChangedCmd returns a tea.Cmd that emits a RowChangedMsg for the current
// cursor position. It guards against out-of-range cursor values.
func (d *Data) rowChangedCmd() tea.Cmd {
	return func() tea.Msg {
		if d.cursor >= 0 && d.cursor < len(d.rows) {
			return RowChangedMsg{
				ColNames: d.colNames,
				Values:   d.rows[d.cursor],
			}
		}
		return nil
	}
}

// MoveUp decrements the cursor (clamped to 0) and scrolls the viewport.
// Returns a command that emits RowChangedMsg for the new cursor position.
func (d *Data) MoveUp() tea.Cmd {
	if d.cursor > 0 {
		d.cursor--
		d.viewport.SetYOffset(d.cursor)
	}
	return d.rowChangedCmd()
}

// MoveDown increments the cursor (clamped to max row) and scrolls the viewport.
// Returns a command that emits RowChangedMsg for the new cursor position.
func (d *Data) MoveDown() tea.Cmd {
	if d.cursor < len(d.rows)-1 {
		d.cursor++
		d.viewport.SetYOffset(d.cursor)
	}
	return d.rowChangedCmd()
}

// PageUp moves the cursor up by one viewport height (clamped to 0).
// Returns a command that emits RowChangedMsg for the new cursor position.
func (d *Data) PageUp() tea.Cmd {
	d.cursor = max(0, d.cursor-d.viewport.Height())
	d.viewport.SetYOffset(d.cursor)
	return d.rowChangedCmd()
}

// PageDown moves the cursor down by one viewport height (clamped to max row).
// Returns a command that emits RowChangedMsg for the new cursor position.
func (d *Data) PageDown() tea.Cmd {
	maxRow := max(0, len(d.rows)-1)
	d.cursor = min(maxRow, d.cursor+d.viewport.Height())
	d.viewport.SetYOffset(d.cursor)
	return d.rowChangedCmd()
}

// SetTable queries the selected table's data with optional filters and stores
// rows for display. Resets the cursor to the first row.
// sanitizeCell strips vertical whitespace from a cell value to guarantee each
// logical row renders as exactly 1 visual line — required for cursor-based
// YOffset scrolling to stay in sync with the viewport.
func sanitizeCell(s string) string {
	r := strings.NewReplacer("\n", " ", "\r", " ", "\t", " ")
	return r.Replace(s)
}

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
			strRow[j] = sanitizeCell(fmt.Sprintf("%v", val))
		}
		d.rows[i] = strRow
	}
	d.cursor = 0
}
