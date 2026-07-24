// Package data provides the middle panel of the SQLite browser, showing
// query results (data preview) for the selected table — no tab switching.
package data

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/IrwantoCia/utility/internal/helper/sqlite"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// defaultLimit is the number of rows fetched per table query.
const defaultLimit = 100

// Data renders the middle panel showing actual data rows from the currently
// selected table in a tabular format.
type Data struct {
	db        *sqlite.DB
	tableName string
	rows      [][]string
	colNames  []string
	errMsg    string
}

// New creates a Data panel backed by a live DB connection.
func New(db *sqlite.DB) *Data {
	return &Data{db: db}
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

// View renders the data preview as a lipgloss table with column headers,
// alternating row colors, and a row-count footer.
func (d *Data) View(width int) string {
	if d.errMsg != "" {
		return style.Default.BrowseEmpty.Render(d.errMsg)
	}

	if d.tableName == "" {
		return style.Default.BrowseEmpty.Render("Select a table to view data")
	}

	t := table.New().
		Headers(d.colNames...).
		Rows(d.rows...).
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
			if row%2 == 0 {
				return style.Default.TableRowAlt
			}
			return lipgloss.NewStyle()
		})

	result := t.String()

	footer := style.Default.BrowseMetaDim.Render(
		fmt.Sprintf("\n%d row(s) shown (limit %d)", len(d.rows), defaultLimit),
	)

	return result + footer
}
