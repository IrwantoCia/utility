// Package csv provides the TUI coordinator component for CSV viewing.
// It delegates to sub-pages: menu (file selection) and result (filter + table).
//
// Routes:
//   - menu   → result (Enter on "Show CSV")
//   - result → menu   (Esc → BackToCsvMenuMsg)
//
// The csv coordinator owns the menu and result pages as peer children.
// It handles navigation between them via custom messages.
package csv

import (
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/csv/menu"
	"github.com/IrwantoCia/utility/internal/tui/csv/result"
)

// pageKind tracks which sub-page is active.
type pageKind int

const (
	pageMenu pageKind = iota
	pageResult
)

// Csv coordinates the CSV workflow.
type Csv struct {
	currentPage pageKind
	menuModel   *menu.Menu
	resultModel *result.Result
	lastWindow  tea.WindowSizeMsg
}

var _ common.Component = (*Csv)(nil)

func New() *Csv {
	return &Csv{
		currentPage: pageMenu,
		menuModel:   menu.New(),
	}
}

func (c *Csv) Init() tea.Cmd {
	if c.currentPage == pageResult && c.resultModel != nil {
		return c.resultModel.Init()
	}
	return c.menuModel.Init()
}

func (c *Csv) View() string {
	if c.currentPage == pageResult && c.resultModel != nil {
		return c.resultModel.View()
	}
	return c.menuModel.View()
}

// Update dispatches to the active sub-page, intercepting
// ShowResultMsg to switch pages and BackToCsvMenuMsg to return to menu.
func (c *Csv) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case menu.ShowResultMsg:
		c.currentPage = pageResult
		c.resultModel = result.New(msg.FilePath)
		c.resultModel.Resize(c.lastWindow)
		return c.resultModel.Init()
	case result.BackToCsvMenuMsg:
		c.currentPage = pageMenu
		return nil
	}

	var cmd tea.Cmd
	switch c.currentPage {
	case pageMenu:
		cmd = c.menuModel.Update(msg)
	case pageResult:
		if c.resultModel != nil {
			cmd = c.resultModel.Update(msg)
		}
	}

	return cmd
}

func (c *Csv) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	c.lastWindow = ws
	var cmds []tea.Cmd
	if cmd := c.menuModel.Resize(ws); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if c.resultModel != nil {
		if cmd := c.resultModel.Resize(ws); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}
