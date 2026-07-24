// Package schema provides the middle panel of the SQLite browser, showing
// column definitions for the selected table with tab toggling.
package schema

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/helper/sqlite"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// Schema renders the middle panel showing columns of the currently selected
// table, with a tab to toggle between Schema and Data preview views.
type Schema struct {
	db        *sqlite.DB
	tableName string
	columns   []sqlite.Column
	errMsg    string
	tab       int // 0 = Schema, 1 = Data preview
}

// New creates a Schema panel backed by a live DB connection.
func New(db *sqlite.DB) *Schema {
	return &Schema{db: db}
}

// SetTable switches the displayed columns to the given table by querying the DB.
func (s *Schema) SetTable(name string) {
	s.tableName = name
	cols, err := s.db.Columns(name)
	if err != nil {
		s.errMsg = fmt.Sprintf("Error: %v", err)
		s.columns = nil
		return
	}
	s.errMsg = ""
	s.columns = cols
}

// View renders the schema or data preview panel content.
func (s *Schema) View(width int) string {
	labelStyle := style.Default.BrowseMetaLabel
	valStyle := style.Default.BrowseMetaValue
	dimStyle := style.Default.BrowseMetaDim

	if s.errMsg != "" {
		return style.Default.BrowseEmpty.Render(s.errMsg)
	}

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
			style.Default.BrowseEmpty.Render("No data loaded yet"),
			"",
			dimStyle.Render("  Load data from within the app"),
		)
		return lipgloss.NewStyle().Width(width).Render(content)
	}

	// Schema tab
	var rows []string
	rows = append(rows, tabs, "")
	rows = append(rows, sectionTitle("Columns for "+s.tableName), "")

	for _, col := range s.columns {
		lbl := labelStyle.Render(col.Name)
		val := valStyle.Render(col.Type)
		line := fmt.Sprintf("  %s  %s", lbl, val)
		if col.NotNull {
			line += "  " + dimStyle.Render("NOT NULL")
		}
		if col.PK {
			line += "  " + dimStyle.Render("PK")
		}
		if col.Default.Valid {
			line += "  " + dimStyle.Render("DEFAULT: "+col.Default.String)
		}
		rows = append(rows, line)
	}

	content := strings.Join(rows, "\n")
	return lipgloss.NewStyle().Width(width).Render(content)
}
