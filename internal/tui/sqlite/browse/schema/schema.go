// Package schema provides the middle panel of the SQLite browser, showing
// column definitions for the selected table with tab toggling.
package schema

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// Schema renders the middle panel showing columns of the currently selected
// table, with a tab to toggle between Schema and Data preview views.
type Schema struct {
	tableName string
	tab       int // 0 = Schema, 1 = Data preview
}

// New creates a Schema panel with no table selected.
func New() *Schema {
	return &Schema{}
}

// SetTable switches the displayed columns to the given table.
func (s *Schema) SetTable(name string) {
	s.tableName = name
}

// View renders the schema or data preview panel content.
func (s *Schema) View(width int) string {
	labelStyle := style.Default.BrowseMetaLabel
	valStyle := style.Default.BrowseMetaValue
	dimStyle := style.Default.BrowseMetaDim

	if s.tableName == "" {
		return style.Default.BrowseEmpty.Render("Select a table to view schema")
	}

	// ── Tab headers ──
	schemaTab := "Schema"
	dataTab := "Data Preview"
	if s.tab == 0 {
		schemaTab = lipgloss.NewStyle().Underline(true).Render("Schema")
		dataTab = dimStyle.Render("Data Preview")
	} else {
		schemaTab = dimStyle.Render("Schema")
		dataTab = lipgloss.NewStyle().Underline(true).Render("Data Preview")
	}
	tabs := lipgloss.JoinHorizontal(lipgloss.Left,
		style.Default.BrowseMetaSection.Render("["),
		schemaTab,
		style.Default.BrowseMetaSection.Render("]"),
		style.Default.BrowseMetaSection.Render("  ["),
		dataTab,
		style.Default.BrowseMetaSection.Render("]"),
	)

	// ── Section separator ──
	sectionTitle := func(title string) string {
		return style.Default.BrowseMetaSection.Render("── " + title + " ──")
	}

	if s.tab == 1 {
		// Data preview tab
		content := lipgloss.JoinVertical(lipgloss.Left,
			tabs,
			"",
			style.Default.BrowseEmpty.Render("📋 No data loaded yet"),
			"",
			dimStyle.Render("  Load data from within the app"),
		)
		return lipgloss.NewStyle().Width(width).Render(content)
	}

	// Schema tab
	var rows []string
	rows = append(rows, tabs, "")
	rows = append(rows, sectionTitle("Columns for "+s.tableName), "")

	columns := s.getColumns()
	for _, col := range columns {
		lbl := labelStyle.Render(col.name)
		val := valStyle.Render(col.typ)
		row := fmt.Sprintf("  %s  %s", lbl, val)
		rows = append(rows, row)
	}

	content := strings.Join(rows, "\n")
	return lipgloss.NewStyle().Width(width).Render(content)
}

type columnDef struct {
	name string
	typ  string
}

// getColumns returns hardcoded placeholder column definitions for the current
// table name.
func (s *Schema) getColumns() []columnDef {
	switch s.tableName {
	case "users":
		return []columnDef{
			{"id", "INTEGER PK NOT NULL"},
			{"name", "TEXT NOT NULL"},
			{"email", "TEXT UNIQUE NOT NULL"},
			{"created_at", "DATETIME DEFAULT CURRENT_TIMESTAMP"},
			{"active", "INTEGER DEFAULT 1"},
		}
	case "orders":
		return []columnDef{
			{"id", "INTEGER PK"},
			{"user_id", "INTEGER FK"},
			{"total", "REAL"},
			{"status", "TEXT"},
			{"created_at", "DATETIME"},
		}
	default:
		return []columnDef{
			{"id", "INTEGER PRIMARY KEY"},
			{"name", "TEXT"},
			{"created_at", "DATETIME"},
		}
	}
}
