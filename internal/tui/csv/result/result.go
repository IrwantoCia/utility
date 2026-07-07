package result

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
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
	table    *table.Table
	cursor   int

	width, height int
	err           error
	keys          KeyMap
	helpModel     help.Model
	searchInput   textinput.Model
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
	_, err := csvparser.Parse(r.filePath, func(headers []string, rows [][]string) struct{} {
		r.headers = headers
		r.rows = rows
		return struct{}{}
	})
	if err != nil {
		r.err = err
		return nil
	}

	r.buildTable()
	return nil
}

func (r *Result) buildTable() {
	if len(r.headers) == 0 {
		return
	}
	r.table = table.New().
		Headers(r.headers...).
		Rows(r.rows...).
		Border(lipgloss.NormalBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == r.cursor {
				return style.Default.RowHighlighted
			}
			return lipgloss.NewStyle()
		})
	if r.width > 0 {
		r.table = r.table.Width(r.width)
	}
	// Table height = total minus search(1) + blank(1) + status(1) + blank(1) + help(~2)
	tableHeight := max(3, r.height-6)
	r.table = r.table.Height(tableHeight)
}

func (r *Result) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	r.width = ws.Width
	r.height = ws.Height
	r.helpModel, _ = r.helpModel.Update(ws)
	r.searchInput.SetWidth(max(20, ws.Width-10))
	if r.table != nil {
		tableHeight := max(3, r.height-6) // -6 to account for search input
		r.table = r.table.Width(ws.Width).Height(tableHeight)
	}
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
	s.WriteString(r.table.String())
	s.WriteString("\n")
	s.WriteString(statusStr)
	s.WriteString("\n")

	// Pad to fill remaining height before help
	for i := lipgloss.Height(s.String()); i <= r.height-lipgloss.Height(helpStr); i++ {
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

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, r.keys.Esc):
			return func() tea.Msg {
				return BackToCsvMenuMsg{}
			}
		case key.Matches(keyMsg, r.keys.Up):
			if len(r.rows) > 0 {
				r.cursor = (r.cursor - 1 + len(r.rows)) % len(r.rows)
				r.buildTable()
			}
		case key.Matches(keyMsg, r.keys.Down):
			if len(r.rows) > 0 {
				r.cursor = (r.cursor + 1) % len(r.rows)
				r.buildTable()
			}
		case key.Matches(keyMsg, r.keys.Enter):
			r.searchInput.Focus()
			return nil
		}
	}
	return nil
}
