// Package menu provides the file-selection TUI for the CSV viewer.
// It renders a cursor-driven menu and an optional file picker.
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

// ShowResultMsg tells the coordinator to navigate to the result page.
type ShowResultMsg struct {
	FilePath string
}

type OptionType int

const (
	TypeInput OptionType = iota
	TypeAction
)

type cursorPos int

const (
	cursorSelectFile cursorPos = iota
	cursorShowCSV
)

type Option struct {
	Label string
	Type  OptionType
}

type Menu struct {
	options       []Option
	cursor        cursorPos
	keys          KeyMap
	helpModel     help.Model
	picker        *filepicker.FilePicker
	pickerOpen    bool
	selectedFile  string
	width, height int
}

var _ common.Component = (*Menu)(nil)

func New() *Menu {
	return &Menu{
		options: []Option{
			{Label: "Select File", Type: TypeInput},
			{Label: "Show CSV", Type: TypeAction},
		},
		keys:      DefaultKeyMap,
		helpModel: help.New(),
		picker:    filepicker.New(),
	}
}

func (m *Menu) Init() tea.Cmd { return nil }

func (m *Menu) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	m.width = ws.Width
	m.height = ws.Height
	m.helpModel, _ = m.helpModel.Update(ws)
	return m.picker.Resize(ws)
}

// View renders the menu; shows the file picker if open.
func (m *Menu) View() string {
	if m.pickerOpen {
		return m.picker.View()
	}

	var content strings.Builder
	content.WriteString("CSV\n\n")

	for i, opt := range m.options {
		cursor := "  "
		if cursorPos(i) == m.cursor {
			cursor = "> "
		}

		display := opt.Label
		if opt.Label == "Select File" && m.selectedFile != "" {
			display = "Select File (" + m.selectedFile + ")"
		}

		if cursorPos(i) == m.cursor {
			content.WriteString(cursor)
			st := style.Default.Highlighted
			if opt.Type == TypeAction {
				st = style.Default.Action
			}
			content.WriteString(st.Render(display))
		} else {
			content.WriteString(cursor)
			content.WriteString(display)
		}
		content.WriteString("\n")
	}

	helpStr := m.helpModel.View(m.keys)

	var s strings.Builder
	s.WriteString(content.String())

	// Pad to fill remaining height before help
	for i := lipgloss.Height(s.String()); i <= m.height-lipgloss.Height(helpStr); i++ {
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
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(keyMsg, m.keys.Down):
			if m.cursor < cursorPos(len(m.options)-1) {
				m.cursor++
			}
		case key.Matches(keyMsg, m.keys.Enter):
			switch m.cursor {
			case cursorSelectFile:
				m.picker.SelectedFile = ""
				m.pickerOpen = true
				return m.picker.Init()
			case cursorShowCSV:
				if m.selectedFile != "" {
					return func() tea.Msg {
						return ShowResultMsg{FilePath: m.selectedFile}
					}
				}
			}
		}
	}

	return nil
}
