// Package details provides the right panel of the SQLite browser, showing
// metadata about the selected table (row count, size, indexes).
package details

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// Details renders the right panel showing metadata about the selected table.
type Details struct {
	tableName string
	rowCount  int
}

// New creates a Details panel with no table selected.
func New() *Details {
	return &Details{}
}

// SetTable updates the panel to show details for the named table.
func (d *Details) SetTable(name string, rowCount int) {
	d.tableName = name
	d.rowCount = rowCount
}

// View renders the details panel content.
func (d *Details) View(width int) string {
	labelStyle := style.Default.BrowseMetaLabel
	valStyle := style.Default.BrowseMetaValue
	dimStyle := style.Default.BrowseMetaDim

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

	info := d.getInfo()
	lines = append(lines,
		row("Table:", d.tableName),
		row("Rows:", info.rows),
		row("Size:", info.size),
		"",
		sectionTitle("Indexes"), "",
	)

	if info.indexes != "" {
		lines = append(lines, dimStyle.Render("  "+info.indexes))
	} else {
		lines = append(lines, dimStyle.Render("  —"))
	}

	if info.created != "" {
		lines = append(lines, "", sectionTitle("Created"), "")
		lines = append(lines, valStyle.Render("  "+info.created))
	}

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().Width(width).Render(content)
}

type tableInfo struct {
	rows    string
	size    string
	indexes string
	created string
}

// getInfo returns hardcoded placeholder details for the selected table.
func (d *Details) getInfo() tableInfo {
	switch d.tableName {
	case "users":
		return tableInfo{
			rows:    "1,204",
			size:    "2.4 MB",
			indexes: "id_idx, email_idx",
			created: "2024-01-15",
		}
	case "orders":
		return tableInfo{
			rows:    "892",
			size:    "1.1 MB",
			indexes: "id_idx, user_id_idx",
			created: "",
		}
	default:
		return tableInfo{
			rows:    "—",
			size:    "—",
			indexes: "—",
			created: "",
		}
	}
}
