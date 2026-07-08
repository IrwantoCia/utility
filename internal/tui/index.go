// Package tui implements the terminal-based user interface main menu and
// navigation.
package tui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/csv"
	"github.com/IrwantoCia/utility/internal/tui/transcribe"
	"github.com/IrwantoCia/utility/internal/tui/s3"
)

type menu struct {
	name      string
	component common.Component
}

type model struct {
	menus  []menu
	cursor int
	active int // -1 = menu, >= 0 = active page index
	help   help.Model
	keys   KeyMap
}

func (m model) activeComponent() common.Component {
	if m.active < 0 || m.active >= len(m.menus) {
		return nil
	}
	return m.menus[m.active].component
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(common.BackToMenuMsg); ok {
		m.active = -1
		return m, nil
	}

	if comp := m.activeComponent(); comp != nil {
		// Route WindowSizeMsg to Resize so sub-components can re-layout
		if ws, ok := msg.(tea.WindowSizeMsg); ok {
			cmd := comp.Resize(ws)
			return m, cmd
		}
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Quit):
				return m, tea.Quit
			}
		}
		cmd := comp.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		var cmds []tea.Cmd
		for _, menu := range m.menus {
			cmd := menu.component.Resize(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.menus) - 1
			}
		case key.Matches(msg, m.keys.Down):
			m.cursor++
			if m.cursor >= len(m.menus) {
				m.cursor = 0
			}
		case key.Matches(msg, m.keys.Select):
			m.active = m.cursor
			return m, m.menus[m.cursor].component.Init()
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	var content string

	if comp := m.activeComponent(); comp != nil {
		content = comp.View()
	} else {
		content = view(m)
		content += "\n" + m.help.View(m.keys)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func view(m model) string {
	var sb strings.Builder
	sb.WriteString("Menu\n")

	for i, v := range m.menus {
		if m.cursor == i {
			sb.WriteString("> ")
		}
		sb.WriteString(v.name)
		sb.WriteString("\n")
	}

	return sb.String()
}

func Run() {
	m := model{
		cursor: 0,
		active: -1,
		help:   help.New(),
		keys:   DefaultKeyMap,
		menus: []menu{
			{name: "CSV", component: csv.New()},
			{name: "S3", component: s3.New()},
			{name: "Transcribe", component: transcribe.New()},
		},
	}
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
