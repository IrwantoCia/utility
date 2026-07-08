// Package menu provides the S3 sub-menu TUI with Upload and Browse options.
package menu

import (
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/components/filepicker"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// ShowUploadMsg tells the S3 coordinator to navigate to the upload page.
type ShowUploadMsg struct{ EnvFile string }

// ShowBrowseMsg tells the S3 coordinator to navigate to the browse page.
type ShowBrowseMsg struct{ EnvFile string }

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

// Menu represents the S3 sub-menu allowing the user to choose Upload or Browse.
type Menu struct {
	options    []Option
	cursor     int
	keys       KeyMap
	helpModel  help.Model
	lastWindow tea.WindowSizeMsg
	picker     *filepicker.FilePicker
	pickerOpen bool
	envFile    string   // selected .env file path
	envInfo    EnvInfo  // parsed S3_* vars from .env
}

var _ common.Component = (*Menu)(nil)

// New creates a new S3 sub-menu with card-style options.
func New() *Menu {
	m := &Menu{
		options: []Option{
			{Label: "Upload", Description: "Upload file to S3 bucket", Icon: "📤", Type: TypeAction},
			{Label: "Browse", Description: "Explore S3 buckets and objects", Icon: "📁", Type: TypeAction},
			{Label: "Set .env", Description: "Choose .env file with S3 credentials", Icon: "📂", Type: TypeInput},
		},
		keys:      DefaultKeyMap,
		helpModel: help.New(),
		picker:    filepicker.New(),
	}

	// Auto-detect .env in current directory
	if _, err := os.Stat(".env"); err == nil {
		m.envFile = ".env"
	}
	m.envInfo.Load(m.envFile)

	return m
}

// Init is a no-op for the static menu.
func (m *Menu) Init() tea.Cmd { return nil }

// Resize stores the window dimensions and initialises the help model.
func (m *Menu) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	m.lastWindow = ws
	m.helpModel, _ = m.helpModel.Update(ws)
	return m.picker.Resize(ws)
}

// View renders the menu as a set of horizontally-centered option cards.
func (m *Menu) View() string {
	if m.pickerOpen {
		return m.picker.View()
	}

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

		// Build label, appending envFile if selected
		label := opt.Label
		if i == 2 && m.envFile != "" { // "Set .env" card
			label = "Set .env (" + m.envFile + ")"
		}

		// Truncate label if needed
		maxLabelWidth := cardWidth - 8
		if len(label) > maxLabelWidth {
			label = "…" + label[len(label)-(maxLabelWidth-3):]
		}

		// Build card content
		titleLine := lipgloss.JoinHorizontal(lipgloss.Left,
			iconStyle.Render(opt.Icon+"  "),
			titleStyle.Render(label),
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

	envSection := m.envInfo.View(cardWidth)
	envSection = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		Width(m.lastWindow.Width).
		Render(envSection)

	content := lipgloss.JoinVertical(lipgloss.Center,
		banner,
		"",
		cardStack,
		"",
		envSection,
	)

	// Center vertically — compute top padding
	contentHeight := lipgloss.Height(content)
	availableHeight := m.lastWindow.Height - helpHeight
	topPad := max(0, (availableHeight-contentHeight)/2)

	var s strings.Builder
	for i := 0; i < topPad; i++ {
		s.WriteRune('\n')
	}
	s.WriteString(content)

	// Pad to fill remaining height before help
	for i := lipgloss.Height(s.String()); i <= m.lastWindow.Height-helpHeight; i++ {
		s.WriteRune('\n')
	}

	s.WriteString(helpStr)
	return s.String()
}

// Update handles keyboard navigation: k/↑, j/↓ cycle, Enter selects, Esc goes back.
func (m *Menu) Update(msg tea.Msg) tea.Cmd {
	if m.pickerOpen {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(keyMsg, m.keys.Esc) {
				m.pickerOpen = false
				return nil
			}
		}
		cmd := m.picker.Update(msg)
		if m.picker.SelectedFile != "" {
			m.envFile = m.picker.SelectedFile
			m.envInfo.Load(m.picker.SelectedFile)
			m.pickerOpen = false
		}
		return cmd
	}

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
			return func() tea.Msg { return ShowUploadMsg{EnvFile: m.envFile} }
		case 1:
			return func() tea.Msg { return ShowBrowseMsg{EnvFile: m.envFile} }
		case 2:
			m.picker.SelectedFile = ""
			m.pickerOpen = true
			return m.picker.Init()
		}
	}

	return nil
}
