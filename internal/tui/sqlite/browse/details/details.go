// Package details provides the right panel of the SQLite browser, showing
// metadata about the selected table (row count, indexes).
package details

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/helper/sqlite"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// Details renders the right panel showing metadata about the selected table.
type Details struct {
	db        *sqlite.DB
	tableName string
	rowCount  int64
	indexes   []sqlite.Index
	errMsg    string
}

// New creates a Details panel backed by a live DB connection.
func New(db *sqlite.DB) *Details {
	return &Details{db: db}
}

// SetTable updates the panel by querying row count and indexes for the named table.
func (d *Details) SetTable(name string) {
	d.tableName = name
	d.errMsg = ""

	count, err := d.db.RowCount(name)
	if err != nil {
		d.errMsg = fmt.Sprintf("Error: %v", err)
		return
	}
	d.rowCount = count

	indexes, err := d.db.Indexes(name)
	if err != nil {
		d.errMsg = fmt.Sprintf("Error: %v", err)
		return
	}
	d.indexes = indexes
}

// formatRowCount formats an int64 with comma separators.
func formatRowCount(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for i := len(s); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		parts = append([]string{s[start:i]}, parts...)
	}
	return strings.Join(parts, ",")
}

// View renders the details panel content.
func (d *Details) View(width int) string {
	labelStyle := style.Default.BrowseMetaLabel
	valStyle := style.Default.BrowseMetaValue
	dimStyle := style.Default.BrowseMetaDim

	if d.errMsg != "" {
		return style.Default.BrowseEmpty.Render(d.errMsg)
	}

	if d.tableName == "" {
		return style.Default.BrowseEmpty.Render("Select a table to view details")
	}

	sectionTitle := func(title string) string {
		return style.Default.BrowseMetaSection.Render("── " + title + " ──")
	}

	row := func(label, value string) string {
		lbl := labelStyle.Render(label)
		val := valStyle.Render(value)
		return fmt.Sprintf("  %s %s", lbl, val)
	}

	var lines []string
	lines = append(lines, sectionTitle("Table Info"), "")
	lines = append(lines,
		row("Table:", d.tableName),
		row("Rows:", formatRowCount(d.rowCount)),
		"",
		sectionTitle("Indexes"), "",
	)

	if len(d.indexes) > 0 {
		for _, idx := range d.indexes {
			colList := strings.Join(idx.Columns, ", ")
			unique := ""
			if idx.Unique {
				unique = " (UNIQUE)"
			}
			lines = append(lines, dimStyle.Render(fmt.Sprintf("  %s%s ON (%s)", idx.Name, unique, colList)))
		}
	} else {
		lines = append(lines, dimStyle.Render("  —"))
	}

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().Width(width).Render(content)
}
