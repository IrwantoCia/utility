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
	"fmt"

	tea "charm.land/bubbletea/v2"
	sqhelper "github.com/IrwantoCia/utility/internal/helper/sqlite"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/browse"
	"github.com/IrwantoCia/utility/internal/tui/sqlite/menu"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// Sqlite coordinates the SQLite workflow.
type Sqlite struct {
	menuModel   *menu.Menu
	browseModel *browse.Browse
	activePage  string // "menu" | "browse"
	lastWindow  tea.WindowSizeMsg
	errorMsg    string
}

var _ common.Component = (*Sqlite)(nil)

// Close implements common.Component.
func (s *Sqlite) Close() tea.Cmd {
	if s.browseModel != nil {
		return s.browseModel.Close()
	}
	return nil
}

func New() *Sqlite {
	return &Sqlite{
		menuModel:  menu.New(),
		activePage: "menu",
	}
}

func (s *Sqlite) Init() tea.Cmd {
	return s.menuModel.Init()
}

func (s *Sqlite) View() string {
	if s.activePage == "browse" && s.browseModel != nil {
		return s.browseModel.View()
	}
	menuView := s.menuModel.View()
	if s.errorMsg != "" {
		errorLine := style.Default.StatusError.Render(s.errorMsg)
		menuView = fmt.Sprintf("%s\n\n%s", menuView, errorLine)
	}
	return menuView
}

func (s *Sqlite) Update(msg tea.Msg) tea.Cmd {
	// Intercept navigation messages first.
	switch msg := msg.(type) {
	case menu.ShowBrowseMsg:
		db, err := sqhelper.Open(msg.DBPath)
		if err != nil {
			s.errorMsg = fmt.Sprintf("Failed to open: %v", err)
			s.activePage = "menu"
			return nil
		}
		s.errorMsg = ""
		s.browseModel = browse.New(db)
		s.activePage = "browse"
		s.browseModel.Resize(s.lastWindow)
		return s.browseModel.Init()
	case browse.BackToSqliteMenuMsg:
		if s.browseModel != nil {
			s.browseModel.Close()
			s.browseModel = nil
		}
		s.activePage = "menu"
		s.errorMsg = ""
		return nil
	}

	if s.activePage == "browse" && s.browseModel != nil {
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
	if s.browseModel != nil {
		cmds = append(cmds, s.browseModel.Resize(ws))
	}
	return tea.Batch(cmds...)
}
