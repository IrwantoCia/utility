// Package schema provides the middle panel of the SQLite browser, showing
// query results (data preview) for the selected table — no tab switching.
package schema

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/helper/sqlite"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// defaultLimit is the number of rows fetched per table query.
const defaultLimit = 100

// Schema renders the middle panel showing actual data rows from the currently
// selected table in a tabular format.
type Schema struct {
	db        *sqlite.DB
	tableName string
	rows      []sqlite.Row
	colNames  []string
	errMsg    string
}

// New creates a Schema panel backed by a live DB connection.
func New(db *sqlite.DB) *Schema {
	return &Schema{db: db}
}

// SetTable queries the selected table's data and stores rows for display.
func (s *Schema) SetTable(name string) {
	s.tableName = name
	rows, colNames, err := s.db.Query(name, defaultLimit, 0)
	if err != nil {
		s.errMsg = fmt.Sprintf("Error: %v", err)
		s.rows = nil
		s.colNames = nil
		return
	}
	s.errMsg = ""
	s.rows = rows
	s.colNames = colNames
}

// View renders the data preview as a table with column headers and row values.
func (s *Schema) View(width int) string {
	dimStyle := style.Default.BrowseMetaDim
	labelStyle := style.Default.BrowseMetaLabel
	valStyle := style.Default.BrowseMetaValue

	if s.errMsg != "" {
		return style.Default.BrowseEmpty.Render(s.errMsg)
	}

	if s.tableName == "" {
		return style.Default.BrowseEmpty.Render("Select a table to view data")
	}

	var b strings.Builder

	// ── Header row ──
	for i, name := range s.colNames {
		if i > 0 {
			b.WriteString(" │ ")
		}
		b.WriteString(labelStyle.Render(name))
	}
	b.WriteByte('\n')

	// ── Separator ──
	sep := strings.Repeat("─", width-2)
	b.WriteString(dimStyle.Render(sep))
	b.WriteByte('\n')

	// ── Data rows ──
	for i, row := range s.rows {
		for j, val := range row {
			if j > 0 {
				b.WriteString(" │ ")
			}
			str := fmt.Sprintf("%v", val)
			if i%2 == 1 {
				b.WriteString(dimStyle.Render(str))
			} else {
				b.WriteString(valStyle.Render(str))
			}
		}
		b.WriteByte('\n')
	}

	// ── Footer ──
	footer := dimStyle.Render(fmt.Sprintf("\n%d row(s) shown (limit %d)", len(s.rows), defaultLimit))
	b.WriteString(footer)

	content := b.String()
	return lipgloss.NewStyle().Width(width).Render(content)
}
