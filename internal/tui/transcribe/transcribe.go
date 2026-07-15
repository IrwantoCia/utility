// Package transcribe provides the TUI component for speech-to-text
// transcription file selection and action.
package transcribe

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

type OptionType int

const (
	TypeInput OptionType = iota
	TypeAction
)

type cursorPos int

const (
	cursorSelectFile cursorPos = iota
	cursorTranscribe
)

type Option struct {
	Label       string
	Description string
	Icon        string
	Type        OptionType
}

type Transcribe struct {
	options      []Option
	cursor       cursorPos
	keys         KeyMap
	helpModel    help.Model
	picker       *filepicker.FilePicker
	pickerOpen   bool
	selectedFile string
	lastWindow   tea.WindowSizeMsg
}

var _ common.Component = (*Transcribe)(nil)

func New() *Transcribe {
	return &Transcribe{
		options: []Option{
			{
				Label:       "Select File",
				Description: "Choose an audio file to transcribe",
				Icon:        "📂",
				Type:        TypeInput,
			},
			{
				Label:       "Transcribe",
				Description: "Start speech-to-text transcription",
				Icon:        "♪",
				Type:        TypeAction,
			},
		},
		keys:      DefaultKeyMap,
		helpModel: help.New(),
		picker:    filepicker.New(),
	}
}

func (t *Transcribe) Init() tea.Cmd { return nil }

func (t *Transcribe) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	t.lastWindow = ws
	t.helpModel, _ = t.helpModel.Update(ws)
	return t.picker.Resize(ws)
}

// View renders the menu; shows the file picker if open.
func (t *Transcribe) View() string {
	if t.pickerOpen {
		return t.picker.View()
	}

	cardWidth := max(40, t.lastWindow.Width*60/100)
	cardWidth = min(cardWidth, 60)

	var cards []string
	for i, opt := range t.options {
		isSelected := cursorPos(i) == t.cursor

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

		if opt.Label == "Select File" && t.selectedFile != "" {
			display := "Select File (" + t.selectedFile + ")"
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

	helpStr := t.helpModel.View(t.keys)
	helpHeight := lipgloss.Height(helpStr)

	cardStack = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		Width(t.lastWindow.Width).
		Render(cardStack)

	// Render banner
	banner := style.Default.MenuTitle.
		Width(t.lastWindow.Width).
		Render(Banner)

	// Combine banner + cardStack
	content := lipgloss.JoinVertical(lipgloss.Center,
		banner,
		"",
		cardStack,
	)

	// Center vertically — compute top padding
	contentHeight := lipgloss.Height(content)
	availableHeight := t.lastWindow.Height - helpHeight
	topPad := max(0, (availableHeight-contentHeight)/2)

	var s strings.Builder
	for i := 0; i < topPad; i++ {
		s.WriteRune('\n')
	}
	s.WriteString(content)

	// Pad to fill remaining height before help
	for i := lipgloss.Height(s.String()); i <= t.lastWindow.Height-helpHeight; i++ {
		s.WriteRune('\n')
	}

	s.WriteString(helpStr)
	return s.String()
}

func (t *Transcribe) Update(msg tea.Msg) tea.Cmd {
	// When the file picker is open, delegate all input to it.
	if t.pickerOpen {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(keyMsg, t.keys.Esc) {
				t.pickerOpen = false
				return nil
			}
		}

		cmd := t.picker.Update(msg)
		if t.picker.SelectedFile != "" {
			t.selectedFile = t.picker.SelectedFile
			t.pickerOpen = false
		}
		return cmd
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, t.keys.Esc):
			return func() tea.Msg {
				return common.BackToMenuMsg{}
			}
		case key.Matches(keyMsg, t.keys.Up):
			t.cursor = (t.cursor - 1 + cursorPos(len(t.options))) % cursorPos(len(t.options))
		case key.Matches(keyMsg, t.keys.Down):
			t.cursor = (t.cursor + 1) % cursorPos(len(t.options))
		case key.Matches(keyMsg, t.keys.Enter):
			switch t.cursor {
			case cursorSelectFile:
				t.picker.SelectedFile = ""
				t.pickerOpen = true
				return t.picker.Init()
			case cursorTranscribe:
				if t.selectedFile != "" {
					return tea.Quit
				}
			}
		}
	}

	return nil
}
