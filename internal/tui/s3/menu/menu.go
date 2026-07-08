// Package menu provides the S3 sub-menu TUI with Upload and Browse options.
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

// ShowUploadMsg tells the S3 coordinator to navigate to the upload page.
type ShowUploadMsg struct{}

// ShowBrowseMsg tells the S3 coordinator to navigate to the browse page.
type ShowBrowseMsg struct{}

// OptionType categorises a menu option.
type OptionType int

const TypeAction OptionType = 0

// Option describes a single selectable card in the menu.
type Option struct {
	Label       string
	Description string
	Icon        string
	Type        OptionType
}

// Menu represents the S3 sub-menu allowing the user to choose Upload or Browse.
type Menu struct {
	options    []Option
	cursor     int
	keys       KeyMap
	helpModel  help.Model
	lastWindow tea.WindowSizeMsg
}

var _ common.Component = (*Menu)(nil)

// New creates a new S3 sub-menu with card-style options.
func New() *Menu {
	return &Menu{
		options: []Option{
			{Label: "Upload", Description: "Upload file to S3 bucket", Icon: "📤", Type: TypeAction},
			{Label: "Browse", Description: "Explore S3 buckets and objects", Icon: "📁", Type: TypeAction},
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
	var cards []string
	for i, opt := range m.options {
		isSelected := i == m.cursor

		// Icon styling
		iconStyle := style.Default.CardIcon
		if isSelected {
			iconStyle = style.Default.CardIconAction
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
			borderColor = lipgloss.Color("46") // green
		}

		cardWidth := max(40, m.lastWindow.Width*60/100)
		cardWidth = min(cardWidth, 60)

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

	// Center vertically — compute top padding
	cardStackHeight := lipgloss.Height(cardStack)
	availableHeight := m.lastWindow.Height - helpHeight
	topPad := max(0, (availableHeight-cardStackHeight)/2)

	var s strings.Builder
	for i := 0; i < topPad; i++ {
		s.WriteRune('\n')
	}
	s.WriteString(cardStack)

	// Pad to fill remaining height before help
	for i := lipgloss.Height(s.String()); i <= m.lastWindow.Height-helpHeight; i++ {
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
		return func() tea.Msg {
			return common.BackToMenuMsg{}
		}
	case key.Matches(keyMsg, m.keys.Up):
		m.cursor = (m.cursor - 1 + len(m.options)) % len(m.options)
	case key.Matches(keyMsg, m.keys.Down):
		m.cursor = (m.cursor + 1) % len(m.options)
	case key.Matches(keyMsg, m.keys.Enter):
		switch m.cursor {
		case 0:
			return func() tea.Msg { return ShowUploadMsg{} }
		case 1:
			return func() tea.Msg { return ShowBrowseMsg{} }
		}
	}

	return nil
}
