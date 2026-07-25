// Package menu provides the SQLite database sub-menu TUI.
// It renders card-style options for selecting and connecting to a .db file.
package menu

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/components/filepicker"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// ShowBrowseMsg is sent when the user connects to a database, telling the
// coordinator to switch to the browse view.
type ShowBrowseMsg struct {
	DBPath string
}

type OptionType int

const (
	TypeInput OptionType = iota
	TypeAction
)

type cursorPos int

const (
	cursorSelectFile cursorPos = iota
	cursorConnect
)

type Option struct {
	Label       string
	Description string
	Icon        string
	Type        OptionType
}

type Menu struct {
	options      []Option
	cursor       cursorPos
	keys         KeyMap
	helpModel    help.Model
	picker       *filepicker.FilePicker
	pickerOpen   bool
	selectedFile string
	lastWindow   tea.WindowSizeMsg
}

var _ common.Component = (*Menu)(nil)

// Close implements common.Component.
func (m *Menu) Close() tea.Cmd { return nil }

func New() *Menu {
	return &Menu{
		options: []Option{
			{
				Label:       "Select Database",
				Description: "Choose a SQLite .db file to open",
				Icon:        "🗄",
				Type:        TypeInput,
			},
			{
				Label:       "Connect",
				Description: "Open the selected database",
				Icon:        "🔌",
				Type:        TypeAction,
			},
		},
		keys:      DefaultKeyMap,
		helpModel: help.New(),
		picker:    filepicker.New(".db", ".sqlite", ".sqlite3", ".db3"),
	}
}

func (m *Menu) Init() tea.Cmd { return nil }

func (m *Menu) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	m.lastWindow = ws
	m.helpModel, _ = m.helpModel.Update(ws)
	return m.picker.Resize(ws)
}

// View renders the menu as a set of centered option cards.
func (m *Menu) View() string {
	if m.pickerOpen {
		return m.picker.View()
	}

	cardWidth := max(40, m.lastWindow.Width*60/100)
	cardWidth = min(cardWidth, 60)

	var cards []string
	for i, opt := range m.options {
		isSelected := cursorPos(i) == m.cursor

		// Icon styling
		iconStyle := style.Default.CardIcon
		if isSelected {
			if opt.Type == TypeAction {
				iconStyle = style.Default.CardIconAction
			} else {
				iconStyle = style.Default.CardIconInput
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

		if opt.Label == "Select Database" && m.selectedFile != "" {
			display := "Select Database (" + m.selectedFile + ")"
			titleLine = lipgloss.JoinHorizontal(lipgloss.Left,
				iconStyle.Render(opt.Icon+"  "),
				titleStyle.Render(display),
			)
		}

		descLine := "   " + descStyle.Render(opt.Description)

		cardContent := lipgloss.JoinVertical(lipgloss.Left,
			titleLine,
			descLine,
		)

		// Card border styling
		borderColor := lipgloss.Color("240")
		if isSelected {
			if opt.Type == TypeAction {
				borderColor = lipgloss.Color("46") // green
			} else {
				borderColor = lipgloss.Color("75") // blue
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
	for i := lipgloss.Height(s.String()); i <= m.lastWindow.Height-helpHeight; i++ {
		s.WriteRune('\n')
	}

	s.WriteString(helpStr)
	return s.String()
}

func (m *Menu) Update(msg tea.Msg) tea.Cmd {
	// When the file picker is open, delegate all input to it.
	if m.pickerOpen {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(keyMsg, m.keys.Esc) {
				m.pickerOpen = false
				return nil
			}
		}

		cmd := m.picker.Update(msg)
		if m.picker.SelectedFile != "" {
			m.selectedFile = m.picker.SelectedFile
			m.pickerOpen = false
		}
		return cmd
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.Esc):
			return func() tea.Msg {
				return common.BackToMenuMsg{}
			}
		case key.Matches(keyMsg, m.keys.Up):
			m.cursor = (m.cursor - 1 + cursorPos(len(m.options))) % cursorPos(len(m.options))
		case key.Matches(keyMsg, m.keys.Down):
			m.cursor = (m.cursor + 1) % cursorPos(len(m.options))
		case key.Matches(keyMsg, m.keys.Enter):
			switch m.cursor {
			case cursorSelectFile:
				m.picker.SelectedFile = ""
				m.pickerOpen = true
				return m.picker.Init()
			case cursorConnect:
				if m.selectedFile != "" {
					return func() tea.Msg {
						return ShowBrowseMsg{DBPath: m.selectedFile}
					}
				}
			}
		}
	}

	return nil
}
