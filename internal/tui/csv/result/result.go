// Package result provides the table viewer TUI for CSV files.
// It renders a search input, data table, and status bar.
package result

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	csvparser "github.com/IrwantoCia/utility/internal/helper/parser/csv_parser"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// BackToCsvMenuMsg tells the coordinator to switch from result back to the CSV menu.
type BackToCsvMenuMsg struct{}

type Result struct {
	filePath string
	rows     [][]string
	headers  []string
	viewport viewport.Model
	ready    bool
	cursor   int

	lastWindow  tea.WindowSizeMsg
	err         error
	keys        KeyMap
	helpModel   help.Model
	searchInput textinput.Model
}

var _ common.Component = (*Result)(nil)

func New(filePath string) *Result {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 64
	ti.SetWidth(30)
	return &Result{
		filePath:    filePath,
		keys:        DefaultKeyMap,
		helpModel:   help.New(),
		searchInput: ti,
	}
}

func (r *Result) Init() tea.Cmd {
	return tea.Batch(
		r.loadCSV(),
		r.searchInput.Focus(),
	)
}

// loadCSV parses the CSV file using the generic parser.
func (r *Result) loadCSV() tea.Cmd {
	return func() tea.Msg {
		_, err := csvparser.Parse(r.filePath, func(headers []string, rows [][]string) struct{} {
			r.headers = headers
			r.rows = rows
			return struct{}{}
		})
		if err != nil {
			r.err = err
			return nil
		}
		r.buildContent()
		return nil
	}
}

func (r *Result) buildContent() {
	if len(r.headers) == 0 {
		return
	}

	t := table.New().
		Headers(r.headers...).
		Rows(r.rows...).
		Border(lipgloss.ThickBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				// Header row
				return style.Default.TableHeader
			}
			// Highlight cursor row (row is 1-indexed: row 1 = first data row)
			if row == r.cursor {
				return style.Default.RowHighlighted
			}
			// Alternate colors for non-cursor rows
			if row%2 == 0 {
				return style.Default.TableRowAlt
			}
			return lipgloss.NewStyle()
		})

	r.viewport.SetContent(t.String())
}

func (r *Result) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	r.lastWindow = ws
	r.helpModel, _ = r.helpModel.Update(ws)
	r.searchInput.SetWidth(max(20, ws.Width-10))

	headerHeight := 3 // "Search: ..." + blank
	footerHeight := 4 // blank + status + blank + help lines

	if !r.ready {
		r.viewport = viewport.New(
			viewport.WithWidth(ws.Width),
			viewport.WithHeight(ws.Height-headerHeight-footerHeight),
		)
		r.ready = true
	} else {
		r.viewport.SetWidth(ws.Width)
		r.viewport.SetHeight(ws.Height - headerHeight - footerHeight)
	}

	r.buildContent()
	// Force viewport to re-render at new dimensions
	r.viewport, _ = r.viewport.Update(ws)
	return nil
}

func (r *Result) View() string {
	if r.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress Esc to go back", r.err)
	}
	if len(r.headers) == 0 {
		return "No data\n\nPress Esc to go back"
	}

	helpStr := r.helpModel.View(r.keys)
	statusStr := fmt.Sprintf("Showing %d rows", len(r.rows))

	var s strings.Builder
	s.WriteString("Search: ")
	s.WriteString(r.searchInput.View())
	s.WriteString("\n\n")
	s.WriteString(r.viewport.View())
	s.WriteString("\n\n")
	s.WriteString(statusStr)
	s.WriteString("\n")

	// Pad to push help to bottom
	for i := lipgloss.Height(s.String()); i <= r.lastWindow.Height-lipgloss.Height(helpStr); i++ {
		s.WriteRune('\n')
	}

	s.WriteString(helpStr)
	return s.String()
}

func (r *Result) Update(msg tea.Msg) tea.Cmd {
	// If search input is focused, delegate to it
	if r.searchInput.Focused() {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(keyMsg, r.keys.Esc) {
				r.searchInput.Blur()
				return nil
			}
		}
		var cmd tea.Cmd
		r.searchInput, cmd = r.searchInput.Update(msg)
		return cmd
	}

	// Intercept keys BEFORE viewport
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, r.keys.Esc):
			return func() tea.Msg {
				return BackToCsvMenuMsg{}
			}
		case key.Matches(keyMsg, r.keys.Enter):
			r.searchInput.Focus()
			return nil
		case key.Matches(keyMsg, r.keys.Up):
			if r.cursor > 0 {
				r.cursor--
				r.buildContent()
				r.viewport.ScrollUp(1)
			}
			return nil
		case key.Matches(keyMsg, r.keys.Down):
			if r.cursor < len(r.rows)-1 {
				r.cursor++
				r.buildContent()
				r.viewport.ScrollDown(1)
			}
			return nil
		}
	}

	// Delegate everything else (Left/Right/PgUp/PgDn/Home/End/mouse) to viewport
	var cmd tea.Cmd
	r.viewport, cmd = r.viewport.Update(msg)
	return cmd
}
