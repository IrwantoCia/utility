// Package csv provides the TUI component for CSV viewing and editing.
package csv

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/components/filepicker"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

type input struct {
	file string
}

type Model struct {
	input      input
	keys       KeyMap
	helpModel  help.Model
	picker     *filepicker.FilePicker
	isInPicker bool
}

var _ common.Component = (*Model)(nil)

func New() *Model {
	return &Model{
		input:     input{},
		keys:      DefaultKeyMap,
		helpModel: help.New(),
		picker:    filepicker.New(),
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	m.helpModel, _ = m.helpModel.Update(ws)
	if m.picker != nil {
		return m.picker.Resize(ws)
	}
	return nil
}

func (m *Model) View() string {
	if m.isInPicker {
		return m.picker.View()
	}

	var s strings.Builder
	s.WriteString("CSV\n\n")

	if m.input.file != "" {
		s.WriteString("File: ")
		s.WriteString(m.input.file)
		s.WriteString("\n")
	} else {
		s.WriteString(style.Default.Highlighted.Render("> Select File"))
		s.WriteString("\n")
	}

	s.WriteString(m.helpModel.View(m.keys))
	return s.String()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if m.isInPicker {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(keyMsg, m.keys.Esc) {
				m.isInPicker = false
				return nil
			}
		}

		cmd := m.picker.Update(msg)
		if m.picker.SelectedFile != "" {
			m.input.file = m.picker.SelectedFile
			m.isInPicker = false
		}
		return cmd
	}

	m.helpModel, _ = m.helpModel.Update(msg)

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.Esc):
			return func() tea.Msg {
				return common.BackToMenuMsg{}
			}
		case key.Matches(keyMsg, m.keys.Enter):
			m.picker.SelectedFile = ""
			m.isInPicker = true
			return m.picker.Init()
		}
	}

	return nil
}
