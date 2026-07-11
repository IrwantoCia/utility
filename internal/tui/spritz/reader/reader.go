// Package reader provides the Spritz RSVP reader TUI page.
package reader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/helper/spritz"
	"github.com/IrwantoCia/utility/internal/helper/spritz/helper"
	spritzreader "github.com/IrwantoCia/utility/internal/helper/spritz/reader"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/components/listpicker"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// BackToSpritzMenuMsg tells the Spritz coordinator to return to the Spritz sub-menu.
type BackToSpritzMenuMsg struct{}

type pageState int

const (
	stateMenu    pageState = iota // card menu (Select File / Start)
	statePicker                   // listpicker open
	stateReady                    // tokens loaded, showing first word preview
	statePlaying                  // RSVP tick loop active
	statePaused                   // paused, can resume
	stateDone                     // all tokens displayed
)

type tickMsg struct{}

type OptionType int

const (
	TypeInput  OptionType = iota
	TypeAction
)

type cursorPos int

const (
	cursorSelectFile cursorPos = iota
	cursorStart
)

type Option struct {
	Label       string
	Description string
	Icon        string
	Type        OptionType
}

// Reader is the RSVP display page with a card-style menu.
type Reader struct {
	state     pageState
	options   []Option
	cursor    cursorPos
	picker    *listpicker.ListPicker
	sdReader  *spritzreader.Reader
	wpm       int
	keys      KeyMap
	helpModel help.Model

	// RSVP rendering
	prefix string // text before ORP
	orp    string // the ORP character (single char or empty)
	suffix string // text after ORP

	// File info
	selectedFile string
	tokenCount   int
	tokenIndex   int

	// Menu state
	pickerOpen bool

	lastWindow tea.WindowSizeMsg
}

var _ common.Component = (*Reader)(nil)

// New creates a new Reader with card-style menu and default 300 WPM.
func New() *Reader {
	return &Reader{
		state: stateMenu,
		options: []Option{
			{Label: "Select File", Description: "Choose from .spritz/ cache", Icon: "📂", Type: TypeInput},
			{Label: "Start", Description: "Begin RSVP reading", Icon: "▶️", Type: TypeAction},
		},
		picker:    listpicker.New(),
		sdReader:  spritzreader.New(),
		wpm:       300,
		keys:      DefaultKeyMap,
		helpModel: help.New(),
	}
}

// Init is a no-op; the card menu is static until the user interacts.
func (r *Reader) Init() tea.Cmd { return nil }

// Resize stores window dimensions and propagates to picker and help.
func (r *Reader) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	r.lastWindow = ws
	r.helpModel, _ = r.helpModel.Update(ws)
	return r.picker.Resize(ws)
}

// View renders the RSVP display, picker overlay, or card menu depending on state.
func (r *Reader) View() string {
	if r.state == stateReady || r.state == statePlaying || r.state == statePaused || r.state == stateDone {
		return r.rsvpView()
	}

	if r.state == statePicker || r.pickerOpen {
		return r.picker.View()
	}

	if r.state == stateMenu {
		return r.menuView()
	}

	return ""
}

// rsvpView renders the RSVP word display with progress bar and info.
func (r *Reader) rsvpView() string {
	helpStr := r.helpModel.View(r.keys)
	helpH := lipgloss.Height(helpStr)
	availH := r.lastWindow.Height - helpH
	w := r.lastWindow.Width

	// Word display — prefix + ORP(red) + suffix, all centered
	orpStyle := style.Default.StatusError.Copy().Bold(true)
	wordStyle := style.Default.MenuTitle.Copy().Bold(true)

	wordLine := wordStyle.Render(r.prefix) + orpStyle.Render(r.orp) + wordStyle.Render(r.suffix)
	if r.orp == "" {
		wordLine = wordStyle.Render(r.prefix + r.suffix)
	}

	// Progress bar
	pct := r.sdReader.Progress()
	barW := min(40, w*50/100)
	filled := int(pct * float64(barW))
	prog := style.Default.StatusSuccess.Render(strings.Repeat("█", filled))
	empty := style.Default.StatusText.Render(strings.Repeat("░", barW-filled))
	progBar := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Width(w).Render(prog + empty)

	// Info line
	var info string
	switch r.state {
	case stateReady:
		info = fmt.Sprintf("WPM: %d  |  Press Space to start", r.wpm)
	case statePaused:
		info = fmt.Sprintf("WPM: %d  |  %d/%d  [PAUSED] Space to resume", r.wpm, r.tokenIndex+1, r.tokenCount)
	case stateDone:
		info = fmt.Sprintf("Done! %d words | WPM: %d", r.tokenCount, r.wpm)
	default: // playing
		info = fmt.Sprintf("WPM: %d  |  %d/%d", r.wpm, r.tokenIndex+1, r.tokenCount)
	}
	infoLine := style.Default.StatusText.Width(w).AlignHorizontal(lipgloss.Center).Render(info)

	// Vertical centering for the word block
	wordBlockH := lipgloss.Height(wordLine)
	infoBlockH := 3 // progress bar + spacing + info
	totalH := wordBlockH + infoBlockH
	topPad := (availH - totalH) / 2
	if topPad < 0 {
		topPad = 0
	}

	var s strings.Builder
	for range topPad {
		s.WriteRune('\n')
	}
	s.WriteString(lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Width(w).Render(wordLine))
	s.WriteRune('\n')
	s.WriteString(progBar)
	s.WriteRune('\n')
	s.WriteString(infoLine)
	for range r.lastWindow.Height - helpH - lipgloss.Height(s.String()) {
		s.WriteRune('\n')
	}
	s.WriteString(helpStr)
	return s.String()
}

// menuView renders the card-style menu matching the parser page pattern.
func (r *Reader) menuView() string {
	helpStr := r.helpModel.View(r.keys)
	helpH := lipgloss.Height(helpStr)
	w := r.lastWindow.Width

	cardWidth := max(40, w*60/100)
	cardWidth = min(cardWidth, 60)

	var cards []string
	for i, opt := range r.options {
		isSelected := cursorPos(i) == r.cursor

		iconStyle := style.Default.CardIcon
		if isSelected {
			if opt.Type == TypeAction {
				iconStyle = style.Default.CardIconAction
			} else {
				iconStyle = style.Default.CardIconInput
			}
		}

		titleStyle := style.Default.CardTitle
		if isSelected {
			titleStyle = style.Default.CardTitleSelected
		}
		descStyle := style.Default.CardDesc

		titleLine := lipgloss.JoinHorizontal(lipgloss.Left,
			iconStyle.Render(opt.Icon+"  "),
			titleStyle.Render(opt.Label),
		)

		if opt.Label == "Select File" && r.selectedFile != "" {
			display := "Select File (" + filepath.Base(r.selectedFile) + ")"
			titleLine = lipgloss.JoinHorizontal(lipgloss.Left,
				iconStyle.Render(opt.Icon+"  "),
				titleStyle.Render(display),
			)
		}

		descLine := "   " + descStyle.Render(opt.Description)
		cardContent := lipgloss.JoinVertical(lipgloss.Left, titleLine, descLine)

		borderColor := lipgloss.Color("240")
		if isSelected {
			if opt.Type == TypeAction {
				borderColor = lipgloss.Color("46")
			} else {
				borderColor = lipgloss.Color("75")
			}
		}

		card := style.Default.CardContainer.
			BorderForeground(borderColor).
			Width(cardWidth).
			Render(cardContent)
		cards = append(cards, card)
	}

	cardStack := lipgloss.JoinVertical(lipgloss.Left, cards...)
	cardStack = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Width(w).Render(cardStack)

	banner := style.Default.MenuTitle.Width(w).Render(Banner)

	var contentBuilder strings.Builder
	contentBuilder.WriteString(banner)
	contentBuilder.WriteString("\n")
	contentBuilder.WriteString(cardStack)

	content := contentBuilder.String()
	contentH := lipgloss.Height(content)
	availH := r.lastWindow.Height - helpH
	topPad := max(0, (availH-contentH)/2)

	var s strings.Builder
	for range topPad {
		s.WriteRune('\n')
	}
	s.WriteString(content)
	for range r.lastWindow.Height - helpH - lipgloss.Height(s.String()) {
		s.WriteRune('\n')
	}
	s.WriteString(helpStr)
	return s.String()
}

// Update dispatches to the state-specific handler.
func (r *Reader) Update(msg tea.Msg) tea.Cmd {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		r.lastWindow = ws
		r.helpModel, _ = r.helpModel.Update(ws)
	}

	switch r.state {
	case stateMenu:
		return r.updateMenu(msg)
	case statePicker:
		return r.updatePicker(msg)
	case stateReady:
		return r.updateReady(msg)
	case statePlaying:
		return r.updatePlaying(msg)
	case statePaused:
		return r.updatePaused(msg)
	case stateDone:
		return r.updateDone(msg)
	}
	return nil
}

// updateMenu handles keyboard input for the card-style menu.
func (r *Reader) updateMenu(msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(km, r.keys.Esc):
		return func() tea.Msg { return BackToSpritzMenuMsg{} }
	case key.Matches(km, r.keys.Up):
		r.cursor = (r.cursor - 1 + cursorPos(len(r.options))) % cursorPos(len(r.options))
	case key.Matches(km, r.keys.Down):
		r.cursor = (r.cursor + 1) % cursorPos(len(r.options))
	case key.Matches(km, r.keys.Enter):
		switch r.cursor {
		case cursorSelectFile:
			cached, err := helper.ListCached()
			if err != nil || len(cached) == 0 {
				return nil // silently fail — no cached files
			}
			r.picker.SetTitle("Cached Files")
			r.picker.SetItems(cached)
			r.picker.Selected = ""
			r.state = statePicker
			r.pickerOpen = true
			return r.picker.Init()
		case cursorStart:
			if r.selectedFile == "" {
				return nil // silently fail — no file selected
			}
			return r.loadFile(r.selectedFile)
		}
	}
	return nil
}

// updatePicker handles the listpicker overlay for selecting a cached file.
func (r *Reader) updatePicker(msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyPressMsg); ok && key.Matches(km, r.keys.Esc) {
		r.state = stateMenu
		r.pickerOpen = false
		return nil
	}

	cmd := r.picker.Update(msg)
	if r.picker.Selected != "" {
		r.selectedFile = r.picker.Selected
		r.state = stateMenu
		r.pickerOpen = false
		return nil
	}
	return cmd
}

// loadFile reads the cached tokens for sourcePath and transitions to stateReady.
func (r *Reader) loadFile(sourcePath string) tea.Cmd {
	cachePath := helper.CachePath(sourcePath)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		r.picker.Selected = ""
		return nil
	}
	var tokens []spritz.Token
	if err := json.Unmarshal(data, &tokens); err != nil {
		r.picker.Selected = ""
		return nil
	}
	r.selectedFile = sourcePath
	r.tokenCount = len(tokens)
	r.sdReader.Load(tokens)
	r.sdReader.SetWPM(r.wpm)

	// Preview first token
	tok, idx, _ := r.sdReader.Next()
	r.setWordDisplay(tok)
	r.tokenIndex = idx

	// Reset so Space/Enter starts from the beginning
	r.sdReader.Reset()
	r.tokenIndex = -1
	r.state = stateReady
	return nil
}

// setWordDisplay splits the token word at the ORP index for RSVP rendering.
func (r *Reader) setWordDisplay(tok spritz.Token) {
	if tok.ORPIndex < 0 || tok.ORPIndex >= len(tok.Word) {
		r.prefix = tok.Word
		r.orp = ""
		r.suffix = ""
		return
	}
	r.prefix = tok.Word[:tok.ORPIndex]
	r.orp = string(tok.Word[tok.ORPIndex])
	r.suffix = tok.Word[tok.ORPIndex+1:]
}

// doTick schedules the next RSVP frame.
func (r *Reader) doTick() tea.Cmd {
	dur := time.Duration(r.sdReader.Interval()) * time.Millisecond
	return tea.Tick(dur, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

// adjustWPM changes the reading speed, clamped to [60, 1000].
func (r *Reader) adjustWPM(delta int) {
	r.wpm += delta
	if r.wpm < 60 {
		r.wpm = 60
	}
	if r.wpm > 1000 {
		r.wpm = 1000
	}
	r.sdReader.SetWPM(r.wpm)
}

// updateReady waits for user input (Space to start, +/- for WPM).
func (r *Reader) updateReady(msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	switch {
	case key.Matches(km, r.keys.Esc):
		r.state = stateMenu
		return nil
	case key.Matches(km, r.keys.Space):
		tok, idx, more := r.sdReader.Next()
		if !more {
			r.state = stateDone
			return nil
		}
		r.setWordDisplay(tok)
		r.tokenIndex = idx
		r.state = statePlaying
		return r.doTick()
	case key.Matches(km, r.keys.Next):
		tok, idx, more := r.sdReader.Next()
		if !more {
			return nil
		}
		r.setWordDisplay(tok)
		r.tokenIndex = idx
	case key.Matches(km, r.keys.Prev):
		tok, idx, _ := r.sdReader.Prev()
		r.setWordDisplay(tok)
		r.tokenIndex = idx
	case key.Matches(km, r.keys.Plus):
		r.adjustWPM(+10)
	case key.Matches(km, r.keys.Minus):
		r.adjustWPM(-10)
	}
	return nil
}

// updatePlaying handles the RSVP tick loop and user input.
func (r *Reader) updatePlaying(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tickMsg:
		tok, idx, more := r.sdReader.Next()
		if !more {
			r.state = stateDone
			return nil
		}
		r.setWordDisplay(tok)
		r.tokenIndex = idx
		return r.doTick()
	case tea.KeyPressMsg:
		switch {
		case key.Matches(m, r.keys.Esc):
			r.state = stateMenu
			return nil
		case key.Matches(m, r.keys.Space):
			r.state = statePaused
			return nil
		case key.Matches(m, r.keys.Next):
			tok, idx, more := r.sdReader.Next()
			if !more {
				return nil
			}
			r.setWordDisplay(tok)
			r.tokenIndex = idx
		case key.Matches(m, r.keys.Prev):
			tok, idx, _ := r.sdReader.Prev()
			r.setWordDisplay(tok)
			r.tokenIndex = idx
		case key.Matches(m, r.keys.Plus):
			r.adjustWPM(+10)
		case key.Matches(m, r.keys.Minus):
			r.adjustWPM(-10)
		}
	}
	return nil
}

// updatePaused handles user input while paused.
func (r *Reader) updatePaused(msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	switch {
	case key.Matches(km, r.keys.Esc):
		r.state = stateMenu
		return nil
	case key.Matches(km, r.keys.Space):
		r.state = statePlaying
		return r.doTick()
	case key.Matches(km, r.keys.Next):
		tok, idx, more := r.sdReader.Next()
		if !more {
			return nil
		}
		r.setWordDisplay(tok)
		r.tokenIndex = idx
	case key.Matches(km, r.keys.Prev):
		tok, idx, _ := r.sdReader.Prev()
		r.setWordDisplay(tok)
		r.tokenIndex = idx
	case key.Matches(km, r.keys.Plus):
		r.adjustWPM(+10)
	case key.Matches(km, r.keys.Minus):
		r.adjustWPM(-10)
	}
	return nil
}

// updateDone handles input after all tokens have been read.
func (r *Reader) updateDone(msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	switch {
	case key.Matches(km, r.keys.Esc):
		r.state = stateMenu
		return nil
	case key.Matches(km, r.keys.Space):
		r.sdReader.Reset()
		r.prefix = ""
		r.orp = ""
		r.suffix = ""
		r.tokenIndex = -1
		r.state = stateReady
		return nil
	}
	return nil
}
