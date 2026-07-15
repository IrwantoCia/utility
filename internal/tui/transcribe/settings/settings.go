package settings

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/helper/whisper"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/components/listpicker"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// BackMsg tells the coordinator to return to the transcribe menu.
type BackMsg struct{}

type cursorPos int

const (
	cursorModel cursorPos = iota
	cursorLanguage
)

// languageItems is the fixed list of supported language selections.
var languageItems = []string{"auto", "en (English)", "id (Indonesian)", "ja (Japanese)", "tl (Tagalog)"}

// Settings holds the sub-page state for configuring whisper model and language.
type Settings struct {
	keys       KeyMap
	helpModel  help.Model
	cursor     cursorPos
	lastWindow tea.WindowSizeMsg

	// Listpickers for interactive selection
	modelPicker  *listpicker.ListPicker
	langPicker   *listpicker.ListPicker
	activePicker *listpicker.ListPicker // non-nil when a picker is open
	activeCursor cursorPos             // which cursor opened the picker

	// Scan results
	models []whisper.Model

	// Options (option-like entries for the card menu)
	options []optionEntry

	// Selected values (set by caller before showing, updated on picker select)
	SelectedModel    string
	SelectedLanguage string
}

type optionEntry struct {
	Label       string
	Description string
	Icon        string
}

var _ common.Component = (*Settings)(nil)

// New creates a new Settings sub-page with the given models and selections.
func New(models []whisper.Model, selectedModel, selectedLanguage string) *Settings {
	s := &Settings{
		keys:             DefaultKeyMap,
		helpModel:        help.New(),
		models:           models,
		modelPicker:      listpicker.New(),
		langPicker:       listpicker.New(),
		SelectedModel:    selectedModel,
		SelectedLanguage: selectedLanguage,
		options: []optionEntry{
			{
				Label:       "Model",
				Description: "Select whisper.cpp model",
				Icon:        "🧠",
			},
			{
				Label:       "Language",
				Description: "Transcription language",
				Icon:        "🌐",
			},
		},
	}

	// Pre-populate language picker items
	s.langPicker.SetItems(languageItems)

	return s
}

// Init is a no-op.
func (s *Settings) Init() tea.Cmd { return nil }

// Resize stores the window size and propagates to help and pickers.
func (s *Settings) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	s.lastWindow = ws
	s.helpModel, _ = s.helpModel.Update(ws)
	s.modelPicker.Resize(ws)
	s.langPicker.Resize(ws)
	return nil
}

// View renders the settings menu or the active picker if one is open.
func (s *Settings) View() string {
	// If a picker is active, show it full-screen
	if s.activePicker != nil {
		return s.activePicker.View()
	}

	cardWidth := max(40, s.lastWindow.Width*60/100)
	cardWidth = min(cardWidth, 60)

	var cards []string
	for i, opt := range s.options {
		isSelected := cursorPos(i) == s.cursor

		// Icon styling
		iconStyle := style.Default.CardIcon
		if isSelected {
			iconStyle = style.Default.CardIconAction
		}

		// Title styling
		titleStyle := style.Default.CardTitle
		if isSelected {
			titleStyle = style.Default.CardTitleSelected
		}

		// Description styling
		descStyle := style.Default.CardDesc

		// Build current value subtitle
		var valueLine string
		switch opt.Label {
		case "Model":
			if s.SelectedModel != "" {
				valueLine = "   " + descStyle.Render("Current: "+s.SelectedModel)
			} else {
				valueLine = "   " + descStyle.Render("No model selected")
			}
		case "Language":
			valueLine = "   " + descStyle.Render("Current: "+s.SelectedLanguage)
		}

		// Build card content
		titleLine := lipgloss.JoinHorizontal(lipgloss.Left,
			iconStyle.Render(opt.Icon+"  "),
			titleStyle.Render(opt.Label),
		)

		descLine := "   " + descStyle.Render(opt.Description)

		cardContent := lipgloss.JoinVertical(lipgloss.Left,
			titleLine,
			descLine,
			valueLine,
		)

		// Card border styling
		borderColor := lipgloss.Color("240")
		if isSelected {
			borderColor = lipgloss.Color("75")
		}

		card := style.Default.CardContainer.
			BorderForeground(borderColor).
			Width(cardWidth).
			Render(cardContent)

		cards = append(cards, card)
	}

	cardStack := lipgloss.JoinVertical(lipgloss.Left, cards...)

	helpStr := s.helpModel.View(s.keys)
	helpHeight := lipgloss.Height(helpStr)

	cardStack = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		Width(s.lastWindow.Width).
		Render(cardStack)

	// Render banner
	banner := style.Default.MenuTitle.
		Width(s.lastWindow.Width).
		Render(Banner)

	// Combine
	content := lipgloss.JoinVertical(lipgloss.Center,
		banner,
		"",
		cardStack,
	)

	// Center vertically
	contentHeight := lipgloss.Height(content)
	availableHeight := s.lastWindow.Height - helpHeight
	topPad := max(0, (availableHeight-contentHeight)/2)

	var sb strings.Builder
	for i := 0; i < topPad; i++ {
		sb.WriteRune('\n')
	}
	sb.WriteString(content)

	for i := lipgloss.Height(sb.String()); i <= s.lastWindow.Height-helpHeight; i++ {
		sb.WriteRune('\n')
	}

	sb.WriteString(helpStr)
	return sb.String()
}

// Update handles keyboard input. If a picker is active, it delegates to it.
// Otherwise, it navigates the settings menu.
func (s *Settings) Update(msg tea.Msg) tea.Cmd {
	// If a picker is active, delegate to it
	if s.activePicker != nil {
		// Handle Esc to close the picker without selecting
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(keyMsg, s.keys.Esc) {
				s.activePicker = nil
				return nil
			}
		}

		cmd := s.activePicker.Update(msg)

		// Check if the picker made a selection
		if s.activePicker.Selected != "" {
			selected := s.activePicker.Selected
			s.activePicker = nil

			switch s.activeCursor {
			case cursorModel:
				// Parse model name from display: "small (466MB)" -> "small"
				if idx := strings.Index(selected, " ("); idx >= 0 {
					s.SelectedModel = selected[:idx]
				} else {
					s.SelectedModel = selected
				}
			case cursorLanguage:
				// Parse language code from display: "id (Indonesian)" -> "id"
				parts := strings.SplitN(selected, " ", 2)
				s.SelectedLanguage = parts[0]
			}
			return nil
		}

		return cmd
	}

	// Handle key presses for the settings menu
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, s.keys.Esc):
			return func() tea.Msg {
				return BackMsg{}
			}
		case key.Matches(keyMsg, s.keys.Up):
			s.cursor = (s.cursor - 1 + cursorPos(len(s.options))) % cursorPos(len(s.options))
		case key.Matches(keyMsg, s.keys.Down):
			s.cursor = (s.cursor + 1) % cursorPos(len(s.options))
		case key.Matches(keyMsg, s.keys.Enter):
			switch s.cursor {
			case cursorModel:
				if len(s.models) == 0 {
					return nil
				}
				items := make([]string, len(s.models))
				for i, m := range s.models {
					items[i] = fmt.Sprintf("%s (%s)", m.Name, formatSize(m.Size))
				}
				s.modelPicker.SetItems(items)
				s.modelPicker.SetTitle("Select Model")
				s.activePicker = s.modelPicker
				s.activeCursor = cursorModel
				return s.modelPicker.Init()
			case cursorLanguage:
				s.langPicker.SetItems(languageItems)
				s.langPicker.SetTitle("Select Language")
				s.activePicker = s.langPicker
				s.activeCursor = cursorLanguage
				return s.langPicker.Init()
			}
		}
	}

	return nil
}

// formatSize returns a human-readable file size string.
func formatSize(bytes int64) string {
	const unit = 1024
	units := []string{"B", "KB", "MB", "GB", "TB"}

	if bytes == 0 {
		return "0 B"
	}

	size := float64(bytes)
	exp := 0
	for size >= unit && exp < len(units)-1 {
		size /= unit
		exp++
	}

	if exp == 0 {
		return fmt.Sprintf("%d B", bytes)
	}
	return fmt.Sprintf("%.1f %s", size, units[exp])
}
