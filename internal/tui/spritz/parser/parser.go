// Package parser provides the Spritz parser TUI page.
// It renders a cursor-driven card menu for file selection and parsing.
package parser

import (
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/components/filepicker"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// BackToSpritzMenuMsg tells the Spritz coordinator to return to the Spritz sub-menu.
type BackToSpritzMenuMsg struct{}

type OptionType int

const (
	TypeInput OptionType = iota
	TypeAction
)

// StatusType classifies the parser status message.
type StatusType int

const (
	StatusNone    StatusType = iota // no status shown
	StatusInfo                      // informational (e.g., "Parsing...")
	StatusError                     // error (e.g., "File not found")
	StatusSuccess                   // success (e.g., "245 tokens parsed")
)

type cursorPos int

const (
	cursorSelectFile cursorPos = iota
	cursorStartParse
)

// Option describes a single card in the parser menu.
type Option struct {
	Label       string
	Description string
	Icon        string
	Type        OptionType
}

// Parser provides a card-style menu page for the Spritz parser.
type Parser struct {
	options      []Option
	cursor       cursorPos
	keys         KeyMap
	helpModel    help.Model
	picker       *filepicker.FilePicker
	pickerOpen   bool
	selectedFile string
	statusText   string
	statusType   StatusType
	lastWindow   tea.WindowSizeMsg
}

var _ common.Component = (*Parser)(nil)

// New creates a new Parser with card-style menu.
func New() *Parser {
	return &Parser{
		options: []Option{
			{
				Label:       "Select File",
				Description: "Choose a text or markdown file",
				Icon:        "📂",
				Type:        TypeInput,
			},
			{
				Label:       "Start Parse",
				Description: "Tokenize and preview the text",
				Icon:        "🔍",
				Type:        TypeAction,
			},
		},
		keys:      DefaultKeyMap,
		helpModel: help.New(),
		picker:    filepicker.New(),
	}
}

// Init is a no-op; the parser menu is static until user interacts.
func (p *Parser) Init() tea.Cmd { return nil }

// Resize stores window dimensions and propagates to help and picker.
func (p *Parser) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	p.lastWindow = ws
	p.helpModel, _ = p.helpModel.Update(ws)
	return p.picker.Resize(ws)
}

// View renders the card menu; shows the file picker if open.
func (p *Parser) View() string {
	if p.pickerOpen {
		return p.picker.View()
	}

	cardWidth := max(40, p.lastWindow.Width*60/100)
	cardWidth = min(cardWidth, 60)

	var cards []string
	for i, opt := range p.options {
		isSelected := cursorPos(i) == p.cursor

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

		if opt.Label == "Select File" && p.selectedFile != "" {
			display := "Select File (" + filepath.Base(p.selectedFile) + ")"
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

	// Center cards horizontally
	cardStack = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		Width(p.lastWindow.Width).
		Render(cardStack)

	// Render banner
	banner := style.Default.MenuTitle.
		Width(p.lastWindow.Width).
		Render(Banner)

	// Build content for vertical centering
	var contentBuilder strings.Builder
	contentBuilder.WriteString(banner)
	contentBuilder.WriteString("\n")
	contentBuilder.WriteString(cardStack)

	if p.statusType != StatusNone {
		var statusStyle lipgloss.Style
		switch p.statusType {
		case StatusError:
			statusStyle = style.Default.StatusError
		case StatusSuccess:
			statusStyle = style.Default.StatusSuccess
		default:
			statusStyle = style.Default.StatusText
		}

		contentBuilder.WriteString("\n\n")
		statusLine := statusStyle.
			AlignHorizontal(lipgloss.Center).
			Width(p.lastWindow.Width).
			Render(p.statusText)
		contentBuilder.WriteString(statusLine)
	}

	content := contentBuilder.String()
	helpStr := p.helpModel.View(p.keys)
	helpHeight := lipgloss.Height(helpStr)

	// Center vertically
	contentHeight := lipgloss.Height(content)
	availableHeight := p.lastWindow.Height - helpHeight
	topPad := max(0, (availableHeight-contentHeight)/2)

	var s strings.Builder
	for i := 0; i < topPad; i++ {
		s.WriteRune('\n')
	}
	s.WriteString(content)

	// Pad to fill remaining height before help
	for i := lipgloss.Height(s.String()); i <= p.lastWindow.Height-helpHeight; i++ {
		s.WriteRune('\n')
	}

	s.WriteString(helpStr)
	return s.String()
}

// Update handles keyboard input and picker delegation.
func (p *Parser) Update(msg tea.Msg) tea.Cmd {
	// When the file picker is open, delegate all input to it.
	if p.pickerOpen {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(keyMsg, p.keys.Esc) {
				p.pickerOpen = false
				return nil
			}
		}

		cmd := p.picker.Update(msg)
		if p.picker.SelectedFile != "" {
			p.selectedFile = p.picker.SelectedFile
			p.statusType = StatusNone
			p.pickerOpen = false
		}
		return cmd
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, p.keys.Esc):
			return func() tea.Msg {
				return BackToSpritzMenuMsg{}
			}
		case key.Matches(keyMsg, p.keys.Up):
			p.cursor = (p.cursor - 1 + cursorPos(len(p.options))) % cursorPos(len(p.options))
		case key.Matches(keyMsg, p.keys.Down):
			p.cursor = (p.cursor + 1) % cursorPos(len(p.options))
		case key.Matches(keyMsg, p.keys.Enter):
			switch p.cursor {
			case cursorSelectFile:
				p.picker.SelectedFile = ""
				p.pickerOpen = true
				return p.picker.Init()
			case cursorStartParse:
				if p.selectedFile == "" {
					p.statusText = "⚠ No file selected"
					p.statusType = StatusError
				} else {
					p.statusText = "📖 Parsing: " + filepath.Base(p.selectedFile)
					p.statusType = StatusInfo
				}
			}
		}
	}

	return nil
}
