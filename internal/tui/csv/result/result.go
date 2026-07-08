// Package result provides the table viewer TUI for CSV files.
// It renders a search input, data table, and status bar.
package result

import (
	"fmt"
	"math"
	"strconv"
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

// csvLoadedMsg is sent after the CSV file has been parsed and the viewport updated.
type csvLoadedMsg struct{}

type Result struct {
	filePath       string
	rows           [][]string
	allRows        [][]string // unfiltered original data
	headers        []string
	viewport       viewport.Model
	ready          bool
	cursor         int
	viewportHeight int

	isFiltered bool

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
	return r.loadCSV()
}

// loadCSV parses the CSV file using the generic parser.
func (r *Result) loadCSV() tea.Cmd {
	return func() tea.Msg {
		_, err := csvparser.Parse(r.filePath, func(headers []string, rows [][]string) struct{} {
			r.headers = headers
			r.rows = rows
			r.allRows = rows
			return struct{}{}
		})
		if err != nil {
			r.err = err
			return csvLoadedMsg{}
		}
		r.buildContent()
		return csvLoadedMsg{}
	}
}

func (r *Result) buildContent() {
	if len(r.headers) == 0 {
		return
	}

	numCols := len(r.headers)
	numericCol := make([]bool, numCols)
	colWidth := make([]int, numCols)
	numericCount := make([]int, numCols)

	for col := range colWidth {
		colWidth[col] = len(r.headers[col])
	}

	for _, row := range r.rows {
		for col, cell := range row {
			if col >= numCols {
				continue
			}
			if isNumeric(cell) {
				numericCount[col]++
			}
			display := cell
			if isNumeric(cell) {
				display = formatThousand(cell)
			}
			if w := len(display); w > colWidth[col] {
				colWidth[col] = w
			}
		}
	}

	// >50% numeric = numeric column
	rowCount := len(r.rows)
	for col := range numericCol {
		if rowCount > 0 && numericCount[col]*2 > rowCount {
			numericCol[col] = true
		}
	}

	formatted := make([][]string, len(r.rows))
	for i, row := range r.rows {
		formatted[i] = make([]string, len(row))
		for j, cell := range row {
			if j < numCols && numericCol[j] {
				formatted[i][j] = formatCell(cell, colWidth[j])
			} else {
				formatted[i][j] = cell
			}
		}
	}

	t := table.New().
		Headers(r.headers...).
		Rows(formatted...).
		Border(lipgloss.ThickBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return style.Default.TableHeader
			}
			if row == r.cursor {
				return style.Default.RowHighlighted
			}
			if row%2 == 0 {
				return style.Default.TableRowAlt
			}
			return lipgloss.NewStyle()
		})

	r.viewport.SetContent(t.String())
}

// applyFilter filters r.allRows using token-based AND matching across all columns.
// Empty query restores all rows.
func (r *Result) applyFilter() {
	query := strings.TrimSpace(r.searchInput.Value())
	if query == "" {
		r.rows = r.allRows
		r.isFiltered = false
	} else {
		tokens := strings.Fields(strings.ToLower(query))
		filtered := make([][]string, 0, len(r.allRows))
		for _, row := range r.allRows {
			if r.rowMatchesTokens(row, tokens) {
				filtered = append(filtered, row)
			}
		}
		r.rows = filtered
		r.isFiltered = true
	}
	r.cursor = 0
	r.buildContent()
	r.viewport.GotoTop()
}

// rowMatchesTokens returns true if every token appears as a case-insensitive
// substring in at least one column of the row.
func (r *Result) rowMatchesTokens(row []string, tokens []string) bool {
	for _, token := range tokens {
		found := false
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), token) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// isNumeric reports whether s can be parsed as a number.
func isNumeric(s string) bool {
	s = strings.ReplaceAll(s, ",", "")
	_, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return err == nil && strings.TrimSpace(s) != ""
}

// formatThousand adds comma thousand separators to a numeric string.
func formatThousand(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s
	}
	neg := n < 0
	if neg {
		n = -n
	}
	intStr := fmt.Sprintf("%.0f", math.Trunc(n))

	// Detect fractional part from original string
	var frac string
	if idx := strings.IndexByte(s, '.'); idx >= 0 && idx+1 < len(s) {
		frac = s[idx:]
	}
	// Build with commas
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	for i, ch := range intStr {
		if i > 0 && (len(intStr)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(ch)
	}
	b.WriteString(frac)
	return b.String()
}

// formatCell right-pads a numeric cell with spaces to reach maxWidth.
func formatCell(cell string, maxWidth int) string {
	if !isNumeric(cell) {
		return cell
	}
	f := formatThousand(cell)
	pad := maxWidth - len(f)
	if pad > 0 {
		return strings.Repeat(" ", pad) + f
	}
	return f
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
		r.viewportHeight = ws.Height - headerHeight - footerHeight
		r.ready = true
	} else {
		r.viewport.SetWidth(ws.Width)
		r.viewport.SetHeight(ws.Height - headerHeight - footerHeight)
		r.viewportHeight = ws.Height - headerHeight - footerHeight
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
	var statusStr string
	if r.isFiltered {
		statusStr = fmt.Sprintf("Showing %d of %d rows (filtered)", len(r.rows), len(r.allRows))
	} else {
		statusStr = fmt.Sprintf("Showing %d rows", len(r.rows))
	}

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
	// Re-render on data loaded
	if _, ok := msg.(csvLoadedMsg); ok {
		return nil
	}

	// If search input is focused, delegate to it
	if r.searchInput.Focused() {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(keyMsg, r.keys.Esc) {
				r.searchInput.Blur()
				return nil
			}
			if key.Matches(keyMsg, r.keys.Tab) {
				r.searchInput.Blur()
				return nil
			}
		}
		var cmd tea.Cmd
		r.searchInput, cmd = r.searchInput.Update(msg)
		r.applyFilter()
		return cmd
	}

	// Intercept keys BEFORE viewport
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, r.keys.Esc):
			return func() tea.Msg {
				return BackToCsvMenuMsg{}
			}
		case key.Matches(keyMsg, r.keys.Tab):
			r.searchInput.Focus()
			return nil
		case key.Matches(keyMsg, r.keys.Up):
			r.cursor--
			if r.cursor < 0 {
				r.cursor = len(r.rows) - 1
				r.viewport.GotoBottom()
			} else {
				r.viewport.ScrollUp(1)
			}
			r.buildContent()
			return nil
		case key.Matches(keyMsg, r.keys.Down):
			r.cursor++
			if r.cursor >= len(r.rows) {
				r.cursor = 0
				r.viewport.GotoTop()
			} else {
				r.viewport.ScrollDown(1)
			}
			r.buildContent()
			return nil
		case key.Matches(keyMsg, r.keys.HalfUp):
			half := max(1, r.viewportHeight)
			r.cursor -= half
			if r.cursor < 0 {
				r.cursor = len(r.rows) - 1
				r.viewport.GotoBottom()
			} else {
				r.viewport.ScrollUp(half)
			}
			r.buildContent()
			return nil
		case key.Matches(keyMsg, r.keys.HalfDown):
			half := max(1, r.viewportHeight)
			r.cursor += half
			if r.cursor >= len(r.rows) {
				r.cursor = 0
				r.viewport.GotoTop()
			} else {
				r.viewport.ScrollDown(half)
			}
			r.buildContent()
			return nil
		}
	}

	// Delegate everything else (Left/Right/PgUp/PgDn/Home/End/mouse) to viewport
	var cmd tea.Cmd
	r.viewport, cmd = r.viewport.Update(msg)
	return cmd
}
