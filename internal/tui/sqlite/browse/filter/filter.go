// Package filter provides a WHERE-clause filter editor panel for the SQLite
// browse feature, allowing users to add, edit, and remove filter rows.
package filter

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	sqhelper "github.com/IrwantoCia/utility/internal/helper/sqlite"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/components/listpicker"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// PanelHeight is the fixed height of the filter panel in lines.
// 1 header + 3 filter rows + 1 add button + 1 padding.
const PanelHeight = 6

type filterMode int

const (
	modeNormal filterMode = iota
	modePickingColumn
	modePickingOp
	modeEditingValue
)

type filterRow struct {
	column string
	op     sqhelper.FilterOp
	value  string
}

// FilterPanel is a TUI component for editing WHERE-clause filters on a SQLite
// table. It displays a list of filter rows (column, operator, value) with
// a modal list-picker for selecting columns/operators and a text input for
// entering values.
type FilterPanel struct {
	db      *sqhelper.DB
	filters []filterRow
	cursor  int
	columns []string

	picker    *listpicker.ListPicker
	textInput textinput.Model

	mode       filterMode
	editingIdx int

	width  int
	height int

	keys     KeyMap
	viewport viewport.Model
}

var _ common.Component = (*FilterPanel)(nil)

// consumedCmd is a no-op sentinel command returned when the filter panel
// consumes a key event without producing a real command. It signals to the
// coordinator that the key was handled.
var consumedCmd tea.Cmd = func() tea.Msg { return nil }

// New creates a FilterPanel backed by a live DB connection.
func New(db *sqhelper.DB) *FilterPanel {
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 100
	ti.SetWidth(30)

	return &FilterPanel{
		db:        db,
		picker:    listpicker.New(),
		textInput: ti,
		keys:      DefaultKeyMap,
		viewport:  viewport.New(),
	}
}

// Init implements common.Component.
func (f *FilterPanel) Init() tea.Cmd { return nil }

// Close implements common.Component.
func (f *FilterPanel) Close() tea.Cmd { return nil }

// Filters converts the internal filter rows to a []sqhelper.Filter slice.
// Only rows with both column and operator set are included.
func (f *FilterPanel) Filters() []sqhelper.Filter {
	result := make([]sqhelper.Filter, 0, len(f.filters))
	for _, fr := range f.filters {
		if fr.column != "" && fr.op != "" {
			result = append(result, sqhelper.Filter{
				Column: fr.column,
				Op:     fr.op,
				Value:  fr.value,
			})
		}
	}
	return result
}

// SetColumns updates the available column names for the column picker.
func (f *FilterPanel) SetColumns(cols []sqhelper.Column) {
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name
	}
	f.columns = names
}

// Resize stores window dimensions, adjusts the viewport, and forwards resize
// to the picker and text input sub-components.
func (f *FilterPanel) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	f.width = ws.Width
	f.height = ws.Height
	f.viewport.SetWidth(ws.Width)
	f.viewport.SetHeight(ws.Height)
	f.picker.Resize(ws)
	f.textInput.SetWidth(max(ws.Width-12, 10))
	return nil
}

// View renders the filter panel. When a picker sub-mode is active, the
// full-screen picker modal is displayed instead of the normal panel.
// Normal panel content is rendered inside a scrollable viewport so the
// overall panel height stays fixed.
func (f *FilterPanel) View() string {
	// In picker modes, show the full-screen picker modal (bypass viewport).
	if f.mode == modePickingColumn || f.mode == modePickingOp {
		return f.picker.View()
	}

	var lines []string

	// Header banner.
	header := style.Default.BrowseBannerActive.
		Width(f.width).
		Render("Filter")
	lines = append(lines, header)

	// Filter rows.
	for i, row := range f.filters {
		col := row.column
		if col == "" {
			col = style.Default.BrowseMetaDim.Render("?")
		}
		op := string(row.op)
		if op == "" {
			op = style.Default.BrowseMetaDim.Render("?")
		}
		val := row.value
		if val == "" {
			val = style.Default.BrowseMetaDim.Render("?")
		}

		var line string
		if f.mode == modeEditingValue && f.editingIdx == i {
			line = fmt.Sprintf("  %s  %s  %s", col, op, f.textInput.View())
		} else {
			line = fmt.Sprintf("  %s  %s  \"%s\"", col, op, val)
		}

		if i == f.cursor && f.mode == modeNormal {
			line = style.Default.BrowseListCursor.Render("❯ ") + line
		} else {
			line = "  " + line
		}
		lines = append(lines, line)
	}

	// Add filter button.
	addBtn := "[+ Add Filter]"
	if f.cursor == len(f.filters) && f.mode == modeNormal {
		addBtn = style.Default.BrowseListCursor.Render("❯ ") + addBtn
	} else {
		addBtn = "  " + addBtn
	}
	lines = append(lines, addBtn)

	content := strings.Join(lines, "\n")
	f.viewport.SetContent(content)
	return f.viewport.View()
}

// Update handles input for the filter panel. Key events are interpreted
// based on the current mode (normal listing, column picker, operator picker,
// or value editor).
//
// Returns consumedCmd for keys that are handled internally (prevents the
// coordinator from processing them as global keys), and nil for keys that
// should fall through to the coordinator (e.g. Tab/Left/Right/Esc in normal
// mode for focus switching or menu navigation).
func (f *FilterPanel) Update(msg tea.Msg) tea.Cmd {
	// ── Picker sub-mode: column or operator selection ──────────────
	if f.mode == modePickingColumn || f.mode == modePickingOp {
		// Intercept Esc to cancel picker.
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(keyMsg, f.keys.Esc) {
				f.mode = modeNormal
				if f.editingIdx < len(f.filters) {
					row := f.filters[f.editingIdx]
					if row.column == "" && row.op == "" && row.value == "" {
						f.deleteRow(f.editingIdx)
					}
				}
				return consumedCmd
			}
		}
		_ = f.picker.Update(msg)
		if f.picker.Selected != "" {
			switch f.mode {
			case modePickingColumn:
				f.filters[f.editingIdx].column = f.picker.Selected
				f.picker.Selected = ""
				f.mode = modePickingOp
				f.picker.SetItems(filterOps())
				f.picker.SetTitle("Select operator")
				return consumedCmd
			case modePickingOp:
				f.filters[f.editingIdx].op = sqhelper.FilterOp(f.picker.Selected)
				f.picker.Selected = ""
				f.mode = modeEditingValue
				f.textInput.SetValue("")
				f.textInput.Focus()
				return textinput.Blink
			}
		}
		// All keys in picker mode are consumed (prevents focus switch).
		return consumedCmd
	}

	// ── Value editing sub-mode ─────────────────────────────────────
	if f.mode == modeEditingValue {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(keyMsg, f.keys.Esc) {
				f.mode = modeNormal
				if f.editingIdx < len(f.filters) {
					row := f.filters[f.editingIdx]
					if row.column == "" && row.op == "" && row.value == "" {
						f.deleteRow(f.editingIdx)
					}
				}
				f.textInput.Blur()
				return consumedCmd
			}
			if key.Matches(keyMsg, f.keys.Enter) {
				f.filters[f.editingIdx].value = f.textInput.Value()
				f.mode = modeNormal
				f.textInput.Blur()
				f.cursor = len(f.filters)
				return consumedCmd
			}
		}
		var cmd tea.Cmd
		f.textInput, cmd = f.textInput.Update(msg)
		// Always consume keys in editing mode.
		if cmd != nil {
			return cmd
		}
		return consumedCmd
	}

	// ── Normal mode: navigate and edit filter rows ─────────────────
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, f.keys.Up):
			if f.cursor > 0 {
				f.cursor--
			}
			return consumedCmd

		case key.Matches(keyMsg, f.keys.Down):
			if f.cursor < len(f.filters) {
				f.cursor++
			}
			return consumedCmd

		case key.Matches(keyMsg, f.keys.Enter):
			if f.cursor < len(f.filters) {
				return f.startEditingRow(f.cursor)
			}
			return f.addNewFilter()

		case key.Matches(keyMsg, f.keys.Delete):
			if f.cursor < len(f.filters) {
				f.filters = append(f.filters[:f.cursor], f.filters[f.cursor+1:]...)
				if f.cursor >= len(f.filters) {
					f.cursor = max(0, len(f.filters)-1)
				}
				return consumedCmd
			}

		case key.Matches(keyMsg, f.keys.Esc):
			// Let the coordinator handle Esc (return to menu).
			return nil
		}
	}

	return nil
}

// startEditingRow sets up the filter panel to edit the first empty field
// (column → op → value) of the row at index i.
func (f *FilterPanel) startEditingRow(i int) tea.Cmd {
	f.editingIdx = i
	row := &f.filters[i]

	switch {
	case row.column == "":
		f.mode = modePickingColumn
		f.picker.SetItems(f.columns)
		f.picker.SetTitle("Select column")
		f.picker.Selected = ""
		return consumedCmd

	case row.op == "":
		f.mode = modePickingOp
		f.picker.SetItems(filterOps())
		f.picker.SetTitle("Select operator")
		f.picker.Selected = ""
		return consumedCmd

	case row.value == "":
		f.mode = modeEditingValue
		f.textInput.SetValue("")
		f.textInput.Focus()
		return textinput.Blink
	}

	return consumedCmd
}

// addNewFilter appends an empty filter row and starts column picking.
func (f *FilterPanel) addNewFilter() tea.Cmd {
	f.filters = append(f.filters, filterRow{})
	f.cursor = len(f.filters) - 1
	f.editingIdx = len(f.filters) - 1

	if len(f.columns) > 0 {
		f.mode = modePickingColumn
		f.picker.SetItems(f.columns)
		f.picker.SetTitle("Select column")
		f.picker.Selected = ""
		return consumedCmd
	}

	// No columns — skip directly to operator picker.
	f.mode = modePickingOp
	f.picker.SetItems(filterOps())
	f.picker.SetTitle("Select operator")
	return consumedCmd
}

// filterOps returns the list of available operator strings for the picker.
// deleteRow removes the filter row at index i and adjusts the cursor.
func (f *FilterPanel) deleteRow(i int) {
	f.filters = append(f.filters[:i], f.filters[i+1:]...)
	if f.cursor >= len(f.filters) {
		f.cursor = max(0, len(f.filters)-1)
	}
}

func filterOps() []string {
	return []string{
		string(sqhelper.OpEQ),
		string(sqhelper.OpNE),
		string(sqhelper.OpGT),
		string(sqhelper.OpLT),
		string(sqhelper.OpGTE),
		string(sqhelper.OpLTE),
		string(sqhelper.OpLIKE),
		string(sqhelper.OpNOTLIKE),
	}
}


