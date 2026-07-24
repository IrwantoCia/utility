// Package sqlite provides the TUI coordinator component for SQLite database
// interaction. It delegates to sub-pages: menu (file selection + connect) and
// browse (table browser).
//
// Routes:
//   - menu → browse (via ShowBrowseMsg)
//   - browse → menu (via BackToSqliteMenuMsg)
//
// The sqlite coordinator owns the sub-pages and handles navigation between
// them via custom messages.
package sqlite

import (
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/browse"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/menu"
)

// Sqlite coordinates the SQLite workflow.
type Sqlite struct {
	menuModel   *menu.Menu
	browseModel *browse.Browse
	activePage  string // "menu" | "browse"
	lastWindow  tea.WindowSizeMsg
}

var _ common.Component = (*Sqlite)(nil)

// Close implements common.Component.
func (s *Sqlite) Close() tea.Cmd { return nil }

func New() *Sqlite {
	return &Sqlite{
		menuModel:   menu.New(),
		browseModel: browse.New(),
		activePage:  "menu",
	}
}

func (s *Sqlite) Init() tea.Cmd {
	return s.menuModel.Init()
}

func (s *Sqlite) View() string {
	if s.activePage == "browse" {
		return s.browseModel.View()
	}
	return s.menuModel.View()
}

func (s *Sqlite) Update(msg tea.Msg) tea.Cmd {
	// Intercept navigation messages first.
	switch msg.(type) {
	case menu.ShowBrowseMsg:
		s.activePage = "browse"
		s.browseModel.Resize(s.lastWindow)
		return s.browseModel.Init()
	case browse.BackToSqliteMenuMsg:
		s.activePage = "menu"
		return nil
	}

	if s.activePage == "browse" {
		return s.browseModel.Update(msg)
	}
	return s.menuModel.Update(msg)
}

// Resize propagates window size changes to all sub-models
// (not just the active page), matching the S3/CSV pattern.
func (s *Sqlite) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	s.lastWindow = ws
	var cmds []tea.Cmd
	cmds = append(cmds, s.menuModel.Resize(ws))
	cmds = append(cmds, s.browseModel.Resize(ws))
	return tea.Batch(cmds...)
}
