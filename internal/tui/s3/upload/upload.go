package upload

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

type BackToS3MenuMsg struct{}

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
	selectedFile string
	bucket       string
	buckets      []string
}

var _ common.Component = (*Upload)(nil)

func New() *Upload {
	return &Upload{
		options: []Option{
			{Label: "Select File", Description: "Choose a file to upload", Icon: "📂", Type: TypeInput},
			{Label: "Bucket", Description: "Select destination bucket", Icon: "🪣", Type: TypeInput},
			{Label: "Upload", Description: "Start upload to S3", Icon: "⬆ ", Type: TypeAction},
		},
		keys:      DefaultKeyMap,
		helpModel: help.New(),
		picker:    filepicker.New(),
		bucket:    "prod",
		buckets:   []string{"prod", "staging", "dev"},
	}
}

func (u *Upload) Init() tea.Cmd { return nil }

func (u *Upload) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	u.lastWindow = ws
	u.helpModel, _ = u.helpModel.Update(ws)
	return u.picker.Resize(ws)
}

func (u *Upload) View() string {
	if u.pickerOpen {
		return u.picker.View()
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
			if isSelected {
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

	banner := style.Default.MenuTitle.
		Width(w).
		Render(Banner)

	content := lipgloss.JoinVertical(lipgloss.Center,
		banner,
		"",
		cardStack,
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
			idx := 0
			for i, b := range u.buckets {
				if b == u.bucket {
					idx = i
					break
				}
			}
			idx = (idx + 1) % len(u.buckets)
			u.bucket = u.buckets[idx]
		case cursorUpload:
			return func() tea.Msg {
				return BackToS3MenuMsg{}
			}
		}
	}

	return nil
}
