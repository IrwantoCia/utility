package upload

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	s3helper "github.com/IrwantoCia/utility/internal/helper/s3"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/components/filepicker"
	"github.com/IrwantoCia/utility/internal/tui/components/listpicker"
	"github.com/IrwantoCia/utility/internal/tui/components/statusbar"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

type BackToS3MenuMsg struct{}

// bucketsLoadedMsg carries the result of loading bucket names from S3.
type bucketsLoadedMsg struct {
	names []string
	err   error
}

// loadBuckets fetches bucket names from the S3 client asynchronously.
func loadBuckets(client *s3helper.S3) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		buckets, err := client.ListBuckets(ctx)
		names := make([]string, 0, len(buckets))
		for _, b := range buckets {
			names = append(names, b.Name)
		}
		return bucketsLoadedMsg{names: names, err: err}
	}
}

type OptionType int

const (
	TypeInput OptionType = iota
	TypeAction
)

type cursorPos int

const (
	cursorSelectFile cursorPos = iota
	cursorBucket
	cursorUpload
)

type Option struct {
	Label       string
	Description string
	Icon        string
	Type        OptionType
}

type Upload struct {
	options      []Option
	cursor       cursorPos
	keys         KeyMap
	helpModel    help.Model
	lastWindow   tea.WindowSizeMsg
	picker       *filepicker.FilePicker
	pickerOpen   bool
	listPicker   *listpicker.ListPicker
	listOpen     bool
	selectedFile string
	bucket       string
	buckets      []string
	status       *statusbar.StatusBar

	client      *s3helper.S3 // S3 client (nil if not configured)
	clientError error        // client init error
}

var _ common.Component = (*Upload)(nil)

func New(client *s3helper.S3, clientErr error) *Upload {
	return &Upload{
		options: []Option{
			{Label: "Select File", Description: "Choose a file to upload", Icon: "📂", Type: TypeInput},
			{Label: "Bucket", Description: "Select destination bucket", Icon: "🪣", Type: TypeInput},
			{Label: "Upload", Description: "Start upload to S3", Icon: "⬆ ", Type: TypeAction},
		},
		keys:        DefaultKeyMap,
		helpModel:   help.New(),
		picker:      filepicker.New(),
		listPicker:  listpicker.New(),
		status:      statusbar.New(),
		bucket:      "",
		buckets:     []string{},
		client:      client,
		clientError: clientErr,
	}
}

func (u *Upload) Init() tea.Cmd { return nil }

func (u *Upload) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	u.lastWindow = ws
	u.helpModel, _ = u.helpModel.Update(ws)
	return tea.Batch(
		u.picker.Resize(ws),
		u.listPicker.Resize(ws),
	)
}

func (u *Upload) View() string {
	if u.pickerOpen {
		return u.picker.View()
	}

	if u.listOpen {
		return u.listPicker.View()
	}

	if u.client == nil {
		if u.clientError != nil {
			errLine := style.Default.Highlighted.Background(lipgloss.Color("1")).Render("Error: " + u.clientError.Error())
			centered := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Width(u.lastWindow.Width).Render(errLine)
			return centered + "\n\nPress Esc to go back"
		}
		hint := style.Default.CardDesc.Render("No .env file configured. Go back to S3 menu and select Set .env.")
		centered := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Width(u.lastWindow.Width).Render(hint)
		return centered + "\n\nPress Esc to go back"
	}

	w, h := u.lastWindow.Width, u.lastWindow.Height
	if w <= 0 || h <= 0 {
		return ""
	}

	cardWidth := max(40, w*60/100)
	cardWidth = min(cardWidth, 60)

	var cards []string
	for i, opt := range u.options {
		isSelected := i == int(u.cursor)

		iconStyle := style.Default.CardIcon
		if isSelected {
			switch opt.Type {
			case TypeInput:
				iconStyle = style.Default.CardIconInput
			case TypeAction:
				iconStyle = style.Default.CardIconAction
			}
		}

		titleStyle := style.Default.CardTitle
		if isSelected {
			titleStyle = style.Default.CardTitleSelected
		}

		descStyle := style.Default.CardDesc

		label := opt.Label
		switch i {
		case int(cursorSelectFile):
			if u.selectedFile != "" {
				label = "Select File (" + u.selectedFile + ")"
			}
		case int(cursorBucket):
			if isSelected && u.bucket != "" {
				label = "Bucket (" + u.bucket + ")"
			}
		}

		// Truncate label from beginning to keep filename visible
		maxLabelWidth := cardWidth - 8
		if len(label) > maxLabelWidth {
			label = "…" + label[len(label)-(maxLabelWidth-3):]
		}

		titleLine := lipgloss.JoinHorizontal(lipgloss.Left,
			iconStyle.Render(opt.Icon+"  "),
			titleStyle.Render(label),
		)

		descLine := "   " + descStyle.Render(opt.Description)

		cardContent := lipgloss.JoinVertical(lipgloss.Left,
			titleLine,
			descLine,
		)

		borderColor := lipgloss.Color("240")
		if isSelected {
			switch opt.Type {
			case TypeInput:
				borderColor = lipgloss.Color("75")
			case TypeAction:
				borderColor = lipgloss.Color("46")
			}
		}

		card := style.Default.CardContainer.
			BorderForeground(borderColor).
			Width(cardWidth).
			Render(cardContent)

		cards = append(cards, card)
	}

	cardStack := lipgloss.JoinVertical(lipgloss.Left, cards...)

	helpStr := u.helpModel.View(u.keys)
	helpHeight := lipgloss.Height(helpStr)

	cardStack = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		Width(w).
		Render(cardStack)

	// Status message box
	msgBox := u.status.View(cardWidth)
	if msgBox != "" {
		msgBox = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Center).
			Width(w).
			Render(msgBox)
	}

	banner := style.Default.MenuTitle.
		Width(w).
		Render(Banner)

	content := lipgloss.JoinVertical(lipgloss.Center,
		banner,
		"",
		cardStack,
		"",
		msgBox,
	)

	contentHeight := lipgloss.Height(content)
	availableHeight := h - helpHeight
	topPad := max(0, (availableHeight-contentHeight)/2)

	var sb strings.Builder
	for i := 0; i < topPad; i++ {
		sb.WriteRune('\n')
	}
	sb.WriteString(content)

	for i := lipgloss.Height(sb.String()); i <= h-helpHeight; i++ {
		sb.WriteRune('\n')
	}

	sb.WriteString(helpStr)
	return sb.String()
}

func (u *Upload) Update(msg tea.Msg) tea.Cmd {
	if u.pickerOpen {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if key.Matches(msg, u.keys.Esc) {
				u.pickerOpen = false
				return nil
			}
		}

		cmd := u.picker.Update(msg)

		if u.picker.SelectedFile != "" {
			u.selectedFile = u.picker.SelectedFile
			u.pickerOpen = false
		}

		return cmd
	}

	if u.listOpen {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if key.Matches(msg, u.keys.Esc) {
				u.listOpen = false
				return nil
			}
		}

		cmd := u.listPicker.Update(msg)
		if u.listPicker.Selected != "" {
			u.bucket = u.listPicker.Selected
			u.listOpen = false
		}
		return cmd
	}

	// Handle async bucket loading result
	if msg, ok := msg.(bucketsLoadedMsg); ok {
		u.status.Clear()
		if msg.err == nil && len(msg.names) > 0 {
			u.buckets = msg.names
			u.bucket = u.buckets[0]
			// Auto-open listpicker with loaded buckets
			u.listPicker.SetItems(u.buckets)
			u.listPicker.SetTitle("Select Bucket")
			u.listOpen = true
			return u.listPicker.Init()
		}
		if msg.err != nil {
			u.status.SetError("Failed: " + msg.err.Error())
		} else {
			u.status.SetError("No buckets found")
		}
		return nil
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(keyMsg, u.keys.Esc):
		return func() tea.Msg {
			return BackToS3MenuMsg{}
		}
	case key.Matches(keyMsg, u.keys.Up):
		u.cursor = cursorPos((int(u.cursor) - 1 + len(u.options)) % len(u.options))
	case key.Matches(keyMsg, u.keys.Down):
		u.cursor = cursorPos((int(u.cursor) + 1) % len(u.options))
	case key.Matches(keyMsg, u.keys.Enter):
		switch u.cursor {
		case cursorSelectFile:
			u.picker.SelectedFile = ""
			u.pickerOpen = true
			return u.picker.Init()
		case cursorBucket:
			if u.client == nil {
				u.status.SetError("No S3 client configured")
				return nil
			}
			if len(u.buckets) > 0 {
				// Already loaded, open listpicker immediately
				u.listPicker.SetItems(u.buckets)
				u.listPicker.SetTitle("Select Bucket")
				u.listOpen = true
				return u.listPicker.Init()
			}
			u.status.SetInfo("Loading buckets...")
			return loadBuckets(u.client)
		case cursorUpload:
			if u.selectedFile == "" {
				u.status.SetError("No file selected")
				return nil
			}
			if u.bucket == "" {
				u.status.SetError("No bucket selected")
				return nil
			}
			// TODO: actual upload
			return nil
		}
	}

	return nil
}
