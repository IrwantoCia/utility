// Package filepicker provides the TUI component for file selection.
package filepicker

import (
	"errors"
	"os"
	"strings"
	"time"

	fp "charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
)

type clearErrorMsg struct{}

type FilePicker struct {
	Model        fp.Model
	SelectedFile string
	err          error
	keys         KeyMap
	helpModel    help.Model
}

var _ common.Component = (*FilePicker)(nil)

// Close implements common.Component.
func (f *FilePicker) Close() tea.Cmd { return nil }

func New() *FilePicker {
	m := fp.New()
	m.CurrentDirectory, _ = os.Getwd()
	m.ShowPermissions = false
	m.KeyMap.Back = key.NewBinding(
		key.WithKeys("h", "backspace", "left"),
		key.WithHelp("h", "back"),
	)

	return &FilePicker{
		Model:     m,
		keys:      DefaultKeyMap,
		helpModel: help.New(),
	}
}

func (f *FilePicker) Init() tea.Cmd {
	return f.Model.Init()
}

func (f *FilePicker) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	f.helpModel, _ = f.helpModel.Update(ws)
	var cmd tea.Cmd
	f.Model, cmd = f.Model.Update(ws)
	return cmd
}

func (f *FilePicker) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case clearErrorMsg:
		f.err = nil
	}

	var cmd tea.Cmd
	f.Model, cmd = f.Model.Update(msg)

	if didSelect, path := f.Model.DidSelectFile(msg); didSelect {
		f.SelectedFile = path
	}

	if didSelect, path := f.Model.DidSelectDisabledFile(msg); didSelect {
		f.err = errors.New(path + " is not valid.")
		f.SelectedFile = ""
		return tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return cmd
}

func (f *FilePicker) View() string {
	var s strings.Builder
	s.WriteString("Please select:\n\n")
	if f.err != nil {
		s.WriteString(f.Model.Styles.DisabledFile.Render(f.err.Error()))
		s.WriteString("\n\n")
	}
	s.WriteString(f.Model.View())
	s.WriteString("\n\n")
	s.WriteString(f.helpModel.View(f.keys))
	return s.String()
}

func clearErrorAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}
