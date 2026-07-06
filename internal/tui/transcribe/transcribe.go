// Package transcribe provides the TUI component for transcription viewing
// and selection.
package transcribe

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
)

type Transcribe struct {
	Messages []string
	Selected int
	keys     KeyMap
}

var _ common.Component = (*Transcribe)(nil)

func New() *Transcribe {
	return &Transcribe{
		Messages: []string{"hello from transcribe", "this is the second message"},
		Selected: 1,
		keys:     DefaultKeyMap,
	}
}

func (t *Transcribe) View() string {
	return t.Messages[t.Selected-1]
}

func (t *Transcribe) Update(msg tea.Msg) tea.Cmd {
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
		}
	}

	return nil
}

func (t *Transcribe) KeyMap() help.KeyMap { return t.keys }
