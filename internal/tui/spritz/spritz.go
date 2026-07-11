// Package spritz provides the TUI coordinator component for the Spritz workflow.
// It delegates to sub-pages: menu (Parser/Reader selection), parser (stub),
// and reader (stub).
//
// Routes:
//   - menu   → parser  (Enter on "Parser")
//   - menu   → reader  (Enter on "Reader")
//   - parser → menu    (Esc → BackToSpritzMenuMsg)
//   - reader → menu    (Esc → BackToSpritzMenuMsg)
//
// The spritz coordinator owns the menu, parser, and reader pages as peer children.
// It handles navigation between them via custom messages.
package spritz

import (
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/spritz/menu"
	"github.com/IrwantoCia/utility/internal/tui/spritz/parser"
	"github.com/IrwantoCia/utility/internal/tui/spritz/reader"
)

// pageKind tracks which sub-page is active.
type pageKind int

const (
	pageMenu pageKind = iota
	pageParser
	pageReader
)

// Spritz coordinates the Spritz workflow.
type Spritz struct {
	currentPage pageKind
	menuModel   *menu.Menu
	parserModel *parser.Parser
	readerModel *reader.Reader
	lastWindow  tea.WindowSizeMsg
}

var _ common.Component = (*Spritz)(nil)

func New() *Spritz {
	return &Spritz{
		currentPage: pageMenu,
		menuModel:   menu.New(),
	}
}

func (c *Spritz) Init() tea.Cmd {
	if c.currentPage == pageParser && c.parserModel != nil {
		return c.parserModel.Init()
	}
	if c.currentPage == pageReader && c.readerModel != nil {
		return c.readerModel.Init()
	}
	return c.menuModel.Init()
}

func (c *Spritz) View() string {
	switch c.currentPage {
	case pageParser:
		if c.parserModel != nil {
			return c.parserModel.View()
		}
	case pageReader:
		if c.readerModel != nil {
			return c.readerModel.View()
		}
	}
	return c.menuModel.View()
}

// Update dispatches to the active sub-page, intercepting
// ShowParserMsg/ShowReaderMsg to switch pages and
// BackToSpritzMenuMsg to return to menu.
func (c *Spritz) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case menu.ShowParserMsg:
		c.currentPage = pageParser
		c.parserModel = parser.New()
		c.parserModel.Resize(c.lastWindow)
		return c.parserModel.Init()
	case menu.ShowReaderMsg:
		c.currentPage = pageReader
		c.readerModel = reader.New()
		c.readerModel.Resize(c.lastWindow)
		return c.readerModel.Init()
	case parser.BackToSpritzMenuMsg:
		c.currentPage = pageMenu
		return nil
	case reader.BackToSpritzMenuMsg:
		c.currentPage = pageMenu
		return nil
	}

	var cmd tea.Cmd
	switch c.currentPage {
	case pageMenu:
		cmd = c.menuModel.Update(msg)
	case pageParser:
		if c.parserModel != nil {
			cmd = c.parserModel.Update(msg)
		}
	case pageReader:
		if c.readerModel != nil {
			cmd = c.readerModel.Update(msg)
		}
	}

	return cmd
}

func (c *Spritz) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	c.lastWindow = ws
	var cmds []tea.Cmd
	if cmd := c.menuModel.Resize(ws); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if c.parserModel != nil {
		if cmd := c.parserModel.Resize(ws); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if c.readerModel != nil {
		if cmd := c.readerModel.Resize(ws); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}
