// Package menu provides the Spritz sub-menu TUI with Parser and Reader options.
package menu

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// ShowParserMsg tells the Spritz coordinator to navigate to the parser page.
type ShowParserMsg struct{}

// ShowReaderMsg tells the Spritz coordinator to navigate to the reader page.
type ShowReaderMsg struct{}

// OptionType categorises a menu option.
type OptionType int

const (
	TypeAction OptionType = iota
	TypeInput
)

// Option describes a single selectable card in the menu.
type Option struct {
	Label       string
	Description string
	Icon        string
	Type        OptionType
}

// Menu represents the Spritz sub-menu allowing the user to choose Parser or Reader.
type Menu struct {
	options    []Option
	cursor     int
	keys       KeyMap
	helpModel  help.Model
	lastWindow tea.WindowSizeMsg
}

var _ common.Component = (*Menu)(nil)

// New creates a new Spritz sub-menu with card-style options.
func New() *Menu {
	return &Menu{
		options: []Option{
			{Label: "Parser", Description: "Preview tokenized text", Icon: "🔍", Type: TypeAction},
			{Label: "Reader", Description: "RSVP speed reading", Icon: "🎯", Type: TypeAction},
		},
		keys:      DefaultKeyMap,
		helpModel: help.New(),
	}
}

// Init is a no-op for the static menu.
func (m *Menu) Init() tea.Cmd { return nil }

// Resize stores the window dimensions and initialises the help model.
func (m *Menu) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	m.lastWindow = ws
	m.helpModel, _ = m.helpModel.Update(ws)
	return nil
}

// View renders the menu as a set of horizontally-centered option cards.
func (m *Menu) View() string {
	w := m.lastWindow.Width
	cardWidth := max(40, w*60/100)
	cardWidth = min(cardWidth, 60)

	var cards []string
	for i, opt := range m.options {
		isSelected := i == m.cursor

		// Icon styling
		iconStyle := style.Default.CardIcon
		if isSelected {
			switch opt.Type {
			case TypeInput:
				iconStyle = style.Default.CardIconInput
			case TypeAction:
				iconStyle = style.Default.CardIconAction
			}
		}

		// Title styling
		titleStyle := style.Default.CardTitle
		if isSelected {
			titleStyle = style.Default.CardTitleSelected
		}

		// Description styling
		descStyle := style.Default.CardDesc

		// Build card content
		titleLine := lipgloss.JoinHorizontal(lipgloss.Left,
			iconStyle.Render(opt.Icon+"  "),
			titleStyle.Render(opt.Label),
		)

		descLine := "   " + descStyle.Render(opt.Description)

		cardContent := lipgloss.JoinVertical(lipgloss.Left,
			titleLine,
			descLine,
		)

		// Card border styling
		borderColor := lipgloss.Color("240")
		if isSelected {
			switch opt.Type {
			case TypeInput:
				borderColor = lipgloss.Color("75") // blue
			case TypeAction:
				borderColor = lipgloss.Color("46") // green
			}
		}

		card := style.Default.CardContainer.
			BorderForeground(borderColor).
			Width(cardWidth).
			Render(cardContent)

		cards = append(cards, card)
	}

	cardStack := lipgloss.JoinVertical(lipgloss.Left, cards...)

	helpStr := m.helpModel.View(m.keys)
	helpHeight := lipgloss.Height(helpStr)

	// Center horizontally — wrap cardStack in full-width container
	cardStack = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		Width(m.lastWindow.Width).
		Render(cardStack)

	// Render banner
	banner := style.Default.MenuTitle.
		Width(m.lastWindow.Width).
		Render(Banner)

	// Combine banner + cardStack
	content := lipgloss.JoinVertical(lipgloss.Center,
		banner,
		"",
		cardStack,
	)

	// Center vertically — compute top padding
	contentHeight := lipgloss.Height(content)
	availableHeight := m.lastWindow.Height - helpHeight
	topPad := max(0, (availableHeight-contentHeight)/2)

	var s strings.Builder
	for range topPad {
		s.WriteRune('\n')
	}
	s.WriteString(content)

	// Pad to fill remaining height before help
	for range m.lastWindow.Height - helpHeight - lipgloss.Height(s.String()) {
		s.WriteRune('\n')
	}

	s.WriteString(helpStr)
	return s.String()
}

// Update handles keyboard navigation: k/↑, j/↓ cycle, Enter selects, Esc goes back.
func (m *Menu) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(keyMsg, m.keys.Esc):
		return func() tea.Msg { return common.BackToMenuMsg{} }
	case key.Matches(keyMsg, m.keys.Up):
		m.cursor = (m.cursor - 1 + len(m.options)) % len(m.options)
	case key.Matches(keyMsg, m.keys.Down):
		m.cursor = (m.cursor + 1) % len(m.options)
	case key.Matches(keyMsg, m.keys.Enter):
		switch m.cursor {
		case 0:
			return func() tea.Msg { return ShowParserMsg{} }
		case 1:
			return func() tea.Msg { return ShowReaderMsg{} }
		}
	}

	return nil
}
