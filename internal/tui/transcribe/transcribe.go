// Package transcribe provides the TUI component for transcription viewing
// and selection.
package transcribe

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/components/filepicker"
)

type Transcribe struct {
	Messages   []string
	Selected   int
	keys       KeyMap
	picker     *filepicker.FilePicker
	isInPicker bool
	filePath   string
	helpModel  help.Model
}

var _ common.Component = (*Transcribe)(nil)

func New() *Transcribe {
	return &Transcribe{
		Messages:  []string{"hello from transcribe", "this is the second message"},
		Selected:  1,
		keys:      DefaultKeyMap,
		picker:    filepicker.New(),
		helpModel: help.New(),
	}
}

func (t *Transcribe) Init() tea.Cmd { return nil }

func (t *Transcribe) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	t.helpModel, _ = t.helpModel.Update(ws)
	if t.picker != nil {
		return t.picker.Resize(ws)
	}
	return nil
}

func (t *Transcribe) View() string {
	if t.isInPicker {
		return t.picker.View()
	}

	if t.filePath != "" {
		return "selected file: " + t.filePath + "\n\n" + t.helpModel.View(t.keys)
	}

	return t.Messages[t.Selected-1] + "\n\n" + t.helpModel.View(t.keys)
}

func (t *Transcribe) Update(msg tea.Msg) tea.Cmd {
	if t.isInPicker {
		cmd := t.picker.Update(msg)
		if t.picker.SelectedFile != "" {
			t.filePath = t.picker.SelectedFile
			t.isInPicker = false
		}
		return cmd
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, t.keys.Up):
			t.Selected -= 1
			if t.Selected <= 0 {
				t.Selected += len(t.Messages)
			}
		case key.Matches(msg, t.keys.Down):
			t.Selected += 1
			if t.Selected > len(t.Messages) {
				t.Selected -= len(t.Messages)
			}
		case key.Matches(msg, t.keys.Enter):
			if t.Selected == 1 {
				t.filePath = ""
				t.isInPicker = true
				return t.picker.Init()
			}
		}
	}

	return nil
}
