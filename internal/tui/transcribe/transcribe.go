// Package transcribe provides the TUI component for speech-to-text
// transcription file selection and action.
package transcribe

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ffmpeghelper "github.com/IrwantoCia/utility/internal/helper/ffmpeg"
	"github.com/IrwantoCia/utility/internal/helper/whisper"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/components/filepicker"
	"github.com/IrwantoCia/utility/internal/tui/components/progressbar"
	"github.com/IrwantoCia/utility/internal/tui/components/statusbar"
	"github.com/IrwantoCia/utility/internal/tui/style"
	"github.com/IrwantoCia/utility/internal/tui/transcribe/settings"
)

type OptionType int

const (
	TypeInput OptionType = iota
	TypeAction
)

type cursorPos int

const (
	cursorSelectFile cursorPos = iota
	cursorSettings
	cursorTranscribe
)

type Option struct {
	Label       string
	Description string
	Icon        string
	Type        OptionType
}

type convertProgressMsg struct{ percent float64 }

type convertDoneMsg struct {
	outputPath string
	err        error
}

type convertStartedMsg struct {
	progressCh <-chan float64
	doneCh     <-chan convertDoneMsg
}

type transcribeProgressMsg struct{ elapsed time.Duration }

type transcribeDoneMsg struct {
	outputPath string
	err        error
}

type transcribeStartedMsg struct {
	progressCh <-chan time.Duration
	doneCh     <-chan transcribeDoneMsg
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

	ffmpeg       *ffmpeghelper.FFmpeg // lazy init on first use
	status       *statusbar.StatusBar
	phase        string // "" = idle, "extract" = ffmpeg, "transcribe" = whisper
	convProgress float64
	convOutput   string        // temp WAV path
	convChannels convertStartedMsg // stored for re-chaining progress
	convCancel   context.CancelFunc

	progressBar *progressbar.ProgressBar

	whisper            *whisper.Whisper
	transcribing       bool
	transcribeProgress time.Duration
	transcribeOutput   string
	transChannels      transcribeStartedMsg
	transCancel        context.CancelFunc

	// Coordinator state for sub-pages
	activePage string // "" = menu, "settings" = settings sub-page

	// Whisper settings (persisted across page switches)
	whisperModels    []whisper.Model
	selectedModel    string
	selectedLanguage string

	// Settings sub-page (lazy-init on first navigation)
	settingsPage *settings.Settings
}

var _ common.Component = (*Transcribe)(nil)

func New() *Transcribe {
	// Scan whisper models on startup
	models, _ := whisper.ScanModels(whisper.DefaultModelDir)
	var selectedModel string
	if len(models) > 0 {
		selectedModel = models[0].Name
	}

	return &Transcribe{
		options: []Option{
			{
				Label:       "Select File",
				Description: "Choose an audio file to transcribe",
				Icon:        "📂",
				Type:        TypeInput,
			},
			{
				Label:       "Settings",
				Description: "Configure model and language",
				Icon:        "⚙",
				Type:        TypeAction,
			},
			{
				Label:       "Transcribe",
				Description: "Start speech-to-text transcription",
				Icon:        "♪",
				Type:        TypeAction,
			},
		},
		keys:              DefaultKeyMap,
		helpModel:         help.New(),
		picker:            filepicker.New(),
		status:            statusbar.New(),
		progressBar:       progressbar.New(),
		whisperModels:     models,
		selectedModel:     selectedModel,
		selectedLanguage:  "auto",
	}
}

func (t *Transcribe) Init() tea.Cmd { return nil }

func (t *Transcribe) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	t.lastWindow = ws
	t.helpModel, _ = t.helpModel.Update(ws)
	cardWidth := max(40, ws.Width*60/100)
	cardWidth = min(cardWidth, 60)
	t.progressBar.SetWidth(cardWidth - 4)
	if t.settingsPage != nil {
		t.settingsPage.Resize(ws)
	}
	return t.picker.Resize(ws)
}

// View renders the active page: settings sub-page or main menu.
func (t *Transcribe) View() string {
	if t.activePage == "settings" && t.settingsPage != nil {
		return t.settingsPage.View()
	}

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

		// Dynamic description for Settings card
		desc := opt.Description
		if opt.Label == "Settings" {
			if t.selectedModel != "" {
				desc = fmt.Sprintf("Model: %s / Language: %s", t.selectedModel, t.selectedLanguage)
			} else {
				desc = "No models found — add .bin files to ~/.config/utility/whisper/models"
			}
		}

		descLine := "   " + descStyle.Render(desc)

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

	// Progress bar during conversion/transcription
	var progressLine string
	if t.phase == "extract" {
		progressLine = style.Default.CardTitle.Render("  " + t.progressBar.View())
	} else if t.phase == "transcribe" {
		elapsed := t.transcribeProgress.Truncate(time.Second).String()
		progressLine = style.Default.CardTitle.Render(fmt.Sprintf("  Transcribing... (%s)", elapsed))
	}

	// Status message box
	msgBox := t.status.View(cardWidth)
	if msgBox != "" {
		msgBox = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Center).
			Width(t.lastWindow.Width).
			Render(msgBox)
	}

	// Combine banner + cardStack + progress + status
	content := lipgloss.JoinVertical(lipgloss.Center,
		banner,
		"",
		cardStack,
		"",
		progressLine,
		msgBox,
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
	// Intercept page-switching messages regardless of current state.
	switch msg.(type) {
	case settings.BackMsg:
		if t.settingsPage != nil {
			t.selectedModel = t.settingsPage.SelectedModel
			t.selectedLanguage = t.settingsPage.SelectedLanguage
		}
		t.activePage = ""
		return nil
	}

	// Route to settings sub-page when active.
	if t.activePage == "settings" && t.settingsPage != nil {
		return t.settingsPage.Update(msg)
	}

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

	switch msg := msg.(type) {
	case convertStartedMsg:
		t.convChannels = msg
		return listenConvert(msg)
	case convertProgressMsg:
		t.convProgress = msg.percent
		t.progressBar.SetPercent(msg.percent)
		return listenConvert(t.convChannels)
	case convertDoneMsg:
		t.phase = ""
		t.convChannels = convertStartedMsg{}
		if msg.err != nil {
			t.status.SetError("Conversion failed: " + msg.err.Error())
			return nil
		}
		t.convOutput = msg.outputPath
		t.status.SetSuccess(fmt.Sprintf("Audio extracted: %s", msg.outputPath))

		// Auto-chain: start whisper transcription after successful extraction
		if t.selectedModel == "" {
			t.status.SetError("No whisper model selected. Configure in Settings.")
			return nil
		}

		if t.whisper == nil {
			var err error
			t.whisper, err = whisper.New()
			if err != nil {
				t.status.SetError("Whisper not found: " + err.Error())
				return nil
			}
		}

		baseName := filepath.Base(t.selectedFile)
		ext := filepath.Ext(baseName)
		outputBase := strings.TrimSuffix(baseName, ext)

		modelPath := filepath.Join(whisper.DefaultModelDir, "ggml-"+t.selectedModel+".bin")
		home, _ := os.UserHomeDir()
		modelPath = strings.Replace(modelPath, "~", home, 1)

		ctx, cancel := context.WithCancel(context.Background())
		t.transCancel = cancel
		t.phase = "transcribe"
		t.transcribeProgress = 0
		t.transcribeOutput = ""
		t.transcribing = true
		t.status.SetInfo("Transcribing...")
		return startTranscribe(ctx, t.whisper, modelPath, msg.outputPath, outputBase, t.selectedLanguage)
	case transcribeStartedMsg:
		t.transChannels = msg
		return listenTranscribe(msg)
	case transcribeProgressMsg:
		t.transcribeProgress = msg.elapsed
		return listenTranscribe(t.transChannels)
	case transcribeDoneMsg:
		t.transcribing = false
		t.phase = ""
		t.transChannels = transcribeStartedMsg{}
		if msg.err != nil {
			t.status.SetError("Transcription failed: " + msg.err.Error())
		} else {
			t.transcribeOutput = msg.outputPath
			t.status.SetSuccess(fmt.Sprintf("Transcription saved: %s", msg.outputPath))
		}
		return nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		// Lock navigation during conversion/transcription — only Esc allowed
		if t.phase != "" {
			if key.Matches(keyMsg, t.keys.Esc) {
				if t.convCancel != nil {
					t.convCancel()
				}
				if t.transCancel != nil {
					t.transCancel()
				}
				t.phase = ""
				t.status.SetInfo("Cancelled")
				return nil
			}
			return nil
		}

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
			case cursorSettings:
				t.settingsPage = settings.New(t.whisperModels, t.selectedModel, t.selectedLanguage)
				t.settingsPage.Resize(t.lastWindow)
				t.activePage = "settings"
				return t.settingsPage.Init()
			case cursorTranscribe:
				if t.selectedFile == "" {
					t.status.SetError("No file selected")
					return nil
				}
				// Lazy init ffmpeg
				if t.ffmpeg == nil {
					var err error
					t.ffmpeg, err = ffmpeghelper.New()
					if err != nil {
						t.status.SetError("FFmpeg not found: " + err.Error())
						return nil
					}
				}
				// Build output path in CWD
				baseName := filepath.Base(t.selectedFile)
				ext := filepath.Ext(baseName)
				outputPath := strings.TrimSuffix(baseName, ext) + ".wav"

				ctx, cancel := context.WithCancel(context.Background())
				t.convCancel = cancel
				t.phase = "extract"
				t.convProgress = 0
				t.convOutput = ""
				t.status.SetInfo("Extracting audio...")
				return startConvert(ctx, t.ffmpeg, t.selectedFile, outputPath)
			}
		}
	}

	return nil
}

func startConvert(ctx context.Context, ff *ffmpeghelper.FFmpeg, inputPath, outputPath string) tea.Cmd {
	return func() tea.Msg {
		progressCh := make(chan float64, 50)
		doneCh := make(chan convertDoneMsg, 1)

		go func() {
			err := ff.Convert(ctx, inputPath, outputPath, func(pct float64) {
				progressCh <- pct
			})
			doneCh <- convertDoneMsg{outputPath: outputPath, err: err}
			close(progressCh)
		}()

		return convertStartedMsg{progressCh: progressCh, doneCh: doneCh}
	}
}

func listenConvert(started convertStartedMsg) tea.Cmd {
	return func() tea.Msg {
		select {
		case pct, ok := <-started.progressCh:
			if ok {
				return convertProgressMsg{percent: pct}
			}
			select {
			case done := <-started.doneCh:
				return done
			default:
				return nil
			}
		case done := <-started.doneCh:
			return done
		}
	}
}

func startTranscribe(ctx context.Context, w *whisper.Whisper, modelPath, inputPath, outputBase, language string) tea.Cmd {
	return func() tea.Msg {
		progressCh := make(chan time.Duration, 100)
		doneCh := make(chan transcribeDoneMsg, 1)

		go func() {
			err := w.Transcribe(ctx, modelPath, inputPath, outputBase, language, func(elapsed time.Duration) {
				select {
				case progressCh <- elapsed:
				default:
				}
			})
			doneCh <- transcribeDoneMsg{outputPath: outputBase + ".txt", err: err}
			close(progressCh)
		}()

		return transcribeStartedMsg{progressCh: progressCh, doneCh: doneCh}
	}
}

func listenTranscribe(started transcribeStartedMsg) tea.Cmd {
	return func() tea.Msg {
		select {
		case elapsed, ok := <-started.progressCh:
			if ok {
				return transcribeProgressMsg{elapsed: elapsed}
			}
			select {
			case done := <-started.doneCh:
				return done
			default:
				return nil
			}
		case done := <-started.doneCh:
			return done
		}
	}
}


