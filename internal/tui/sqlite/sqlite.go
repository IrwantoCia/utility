// Package sqlite provides the TUI coordinator component for SQLite database
// interaction. It delegates to sub-pages: menu (file selection + connect).
//
// Routes:
//   - menu → (future: query/table browser pages)
//
// The sqlite coordinator owns the menu page and handles navigation between
// sub-pages via custom messages.
package sqlite

import (
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/menu"
)

// Sqlite coordinates the SQLite workflow.
type Sqlite struct {
	menuModel *menu.Menu
	lastWindow tea.WindowSizeMsg
}

var _ common.Component = (*Sqlite)(nil)

// Close implements common.Component.
func (s *Sqlite) Close() tea.Cmd { return nil }

func New() *Sqlite {
	return &Sqlite{
		menuModel: menu.New(),
	}
}

func (s *Sqlite) Init() tea.Cmd {
	return s.menuModel.Init()
}

func (s *Sqlite) View() string {
	return s.menuModel.View()
}

func (s *Sqlite) Update(msg tea.Msg) tea.Cmd {
	return s.menuModel.Update(msg)
}

func (s *Sqlite) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	s.lastWindow = ws
	return s.menuModel.Resize(ws)
}
