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

// OptionType distinguishes input options from action options.
type OptionType int

const (
	TypeInput OptionType = iota
	TypeAction
)

// Option represents a single configurable action/input in the CSV workflow.
type Option struct {
	Label string
	Value string
	Type  OptionType
}

type Model struct {
	options    []Option
	cursor     int
	keys       KeyMap
	helpModel  help.Model
	picker     *filepicker.FilePicker
	isInPicker bool
}

var _ common.Component = (*Model)(nil)

func New() *Model {
	return &Model{
		options: []Option{
			{Label: "Select File"},
			{Label: "Show CSV", Type: TypeAction},
		},
		cursor:    0,
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

	var prevType OptionType
	for i, opt := range m.options {
		if i > 0 && opt.Type != prevType {
			s.WriteString("  ──────────\n")
		}
		prevType = opt.Type

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		display := opt.Label
		if opt.Value != "" {
			display = opt.Value
		}

		if i == m.cursor {
			s.WriteString(cursor)
			st := style.Default.Highlighted
			if opt.Type == TypeAction {
				st = style.Default.Action
			}
			s.WriteString(st.Render(display))
		} else {
			s.WriteString(cursor)
			s.WriteString(display)
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
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
			m.options[m.cursor].Value = m.picker.SelectedFile
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
		case key.Matches(keyMsg, m.keys.Up):
			m.cursor = max(m.cursor-1, 0)
		case key.Matches(keyMsg, m.keys.Down):
			m.cursor = min(m.cursor+1, len(m.options)-1)
		case key.Matches(keyMsg, m.keys.Enter):
			m.picker.SelectedFile = ""
			m.isInPicker = true
			return m.picker.Init()
		}
	}

	return nil
}
