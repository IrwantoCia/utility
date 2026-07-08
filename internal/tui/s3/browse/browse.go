// Package browse provides the 2-panel S3 browser coordinator
// (buckets on the left, objects on the right).
package browse

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/s3/browse/buckets"
	"github.com/IrwantoCia/utility/internal/tui/s3/browse/objects"
)

// BackToS3MenuMsg tells the S3 coordinator to return to the S3 sub-menu.
type BackToS3MenuMsg struct{}

// Browse coordinates the 2-panel S3 browser (buckets + objects).
type Browse struct {
	buckets *buckets.Buckets
	objects *objects.Objects

	focus      int // 0=buckets, 1=objects
	lastWindow tea.WindowSizeMsg
	keys       KeyMap
	helpModel  help.Model

	// Computed layout dimensions.
	leftW, rightW int
	headerH, footerH, panelH int
}

var _ common.Component = (*Browse)(nil)

// New creates a new Browse coordinator.
func New() *Browse {
	return &Browse{
		buckets:   buckets.New(),
		objects:   objects.New(),
		focus:     0,
		keys:      DefaultKeyMap,
		helpModel: help.New(),
	}
}

// Init initialises both sub-panels concurrently.
func (b *Browse) Init() tea.Cmd {
	return tea.Batch(
		b.buckets.Init(),
		b.objects.Init(),
	)
}

// Resize computes the 2-column layout and forwards adjusted sizes to sub-panels.
func (b *Browse) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	b.lastWindow = ws
	b.helpModel, _ = b.helpModel.Update(ws)

	b.headerH = 2
	b.footerH = 2
	b.panelH = ws.Height - b.headerH - b.footerH

	b.leftW = ws.Width * 25 / 100
	b.rightW = ws.Width - b.leftW

	// Sub-panels receive content area (borders consume 2 chars per axis).
	innerH := b.panelH - 2

	leftWS := tea.WindowSizeMsg{Width: b.leftW - 2, Height: innerH}
	rightWS := tea.WindowSizeMsg{Width: b.rightW - 2, Height: innerH}

	return tea.Batch(
		b.buckets.Resize(leftWS),
		b.objects.Resize(rightWS),
	)
}

// View renders the header, two bordered panels, and footer.
func (b *Browse) View() string {
	header := b.renderHeader()
	middle := b.renderPanels()
	footer := b.renderFooter()
	return lipgloss.JoinVertical(lipgloss.Left, header, middle, footer)
}

// renderHeader returns a styled title bar.
func (b *Browse) renderHeader() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("63")).
		Width(b.lastWindow.Width).
		AlignHorizontal(lipgloss.Center).
		Padding(0, 1)
	return style.Render("S3 Browser")
}

// renderPanels builds the two bordered panels joined horizontally.
func (b *Browse) renderPanels() string {
	bucketView := b.wrapPanel(b.buckets.View(), b.leftW, b.panelH, b.focus == 0, "Buckets")
	objectView := b.wrapPanel(b.objects.View(), b.rightW, b.panelH, b.focus == 1, "Objects")
	return lipgloss.JoinHorizontal(lipgloss.Top, bucketView, objectView)
}

// wrapPanel surrounds content with a bordered box with a highlighted title.
func (b *Browse) wrapPanel(content string, w, h int, active bool, title string) string {
	activeColor := lipgloss.Color("63")
	inactiveColor := lipgloss.Color("240")

	borderColor := inactiveColor
	if active {
		borderColor = activeColor
	}

	// Account for the space consumed by top/bottom banner lines.
	innerH := h - 2
	innerW := w - 2

	// Build banner for the top of the panel.
	bannerStyle := lipgloss.NewStyle().
		Width(innerW).
		Background(borderColor).
		Foreground(lipgloss.Color("255")).
		Bold(true).
		Padding(0, 1)

	banner := bannerStyle.Render(title)

	// Build the inner content area.
	contentStyle := lipgloss.NewStyle().
		Width(innerW).
		Height(innerH - 1) // minus banner

	paddedContent := contentStyle.Render(content)

	// Assemble: top border + banner, content, bottom border.
	top := lipgloss.NewStyle().
		Width(w).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(borderColor).
		Render(banner)

	mid := lipgloss.NewStyle().
		Width(w).
		Border(lipgloss.NormalBorder(), true, false, true, false).
		BorderForeground(borderColor).
		Render(paddedContent)

	bot := lipgloss.NewStyle().
		Width(w).
		Border(lipgloss.NormalBorder(), true, true, true, false).
		BorderForeground(borderColor).
		Render("")

	return lipgloss.JoinVertical(lipgloss.Left, top, mid, bot)
}

// renderFooter returns the status bar with key help.
func (b *Browse) renderFooter() string {
	helpStr := b.helpModel.View(b.keys)
	status := strings.Repeat(" ", b.lastWindow.Width-lipgloss.Width(helpStr))
	paddedHelp := lipgloss.NewStyle().
		Width(lipgloss.Width(helpStr)).
		AlignHorizontal(lipgloss.Right).
		Render(helpStr)

	style := lipgloss.NewStyle().
		Width(b.lastWindow.Width).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("240"))

	return style.Render(status + paddedHelp)
}

// Update handles input: Esc returns to S3 menu, Tab cycles focus,
// KeyPressMsg is routed to the active panel, and everything else
// is broadcast to both panels.
func (b *Browse) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, b.keys.Esc) {
			return func() tea.Msg {
				return BackToS3MenuMsg{}
			}
		}
		if key.Matches(msg, b.keys.Tab) {
			b.focus = (b.focus + 1) % 2
			return nil
		}

		// Route to the focused panel.
		switch b.focus {
		case 0:
			return b.buckets.Update(msg)
		case 1:
			return b.objects.Update(msg)
		}

	default:
		// Broadcast non-key events to both panels.
		var cmds []tea.Cmd
		if cmd := b.buckets.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if cmd := b.objects.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if len(cmds) > 0 {
			return tea.Batch(cmds...)
		}
	}

	return nil
}
