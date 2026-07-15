// Package tui implements the terminal-based user interface main menu and
// navigation.
package tui

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/csv"
	"github.com/IrwantoCia/utility/internal/tui/s3"
	"github.com/IrwantoCia/utility/internal/tui/spritz"
	"github.com/IrwantoCia/utility/internal/tui/style"
	"github.com/IrwantoCia/utility/internal/tui/transcribe"
)

type menu struct {
	name        string
	description string
	component   common.Component
}

type model struct {
	menus      []menu
	cursor     int
	active     int // -1 = menu, >= 0 = active page index
	lastWindow tea.WindowSizeMsg
	help       help.Model
	keys       KeyMap
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
				// Give the active component a chance to clean up
				// (e.g. cancel background ffmpeg/whisper processes).
				if closer, ok := comp.(interface{ Close() }); ok {
					closer.Close()
				}
				return m, tea.Quit
			}
		}
		cmd := comp.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.lastWindow = msg
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
	banner := style.Default.MenuTitle.Render(Banner)

	var items []string
	for i, v := range m.menus {
		if m.cursor == i {
			items = append(items, style.Default.MenuItemSelected.Render("▸ "+v.name))
		} else {
			items = append(items, style.Default.MenuItem.Render("  "+v.name))
		}
	}

	menuContent := lipgloss.JoinVertical(lipgloss.Left, items...)
	menuBox := style.Default.MenuContainer.Render(menuContent)

	// Description box below menu
	var descBox string
	if m.cursor >= 0 && m.cursor < len(m.menus) {
		desc := m.menus[m.cursor].description
		if desc != "" {
			descBox = style.Default.StatusBox.Render(style.Default.MenuDesc.Render(desc))
		}
	}

	parts := []string{banner, "", menuBox}
	if descBox != "" {
		parts = append(parts, "", descBox)
	}
	content := lipgloss.JoinVertical(lipgloss.Center, parts...)

	if m.lastWindow.Width > 0 {
		content = lipgloss.Place(
			m.lastWindow.Width,
			m.lastWindow.Height-3,
			lipgloss.Center, lipgloss.Center,
			content,
		)
	}

	return content
}

func Run() {
	m := model{
		cursor: 0,
		active: -1,
		help:   help.New(),
		keys:   DefaultKeyMap,
		menus: []menu{
			{name: "CSV", description: "View and search CSV files with table display", component: csv.New()},
			{name: "S3", description: "Browse and manage S3 buckets and objects", component: s3.New()},
			{name: "Spritz", description: "RSVP speed reading — word by word", component: spritz.New()},
			{name: "Transcribe", description: "Speech-to-text transcription", component: transcribe.New()},
		},
	}
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
