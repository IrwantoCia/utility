// Package browse provides the 2-panel S3 browser coordinator
// (buckets on the left, objects on the right).
package browse

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	s3helper "github.com/IrwantoCia/utility/internal/helper/s3"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/s3/browse/buckets"
	"github.com/IrwantoCia/utility/internal/tui/s3/browse/objects"
)

// BackToS3MenuMsg tells the S3 coordinator to return to the S3 sub-menu.
type BackToS3MenuMsg struct{}

type bucketsLoadedMsg struct {
	names []string
	err   error
}

type objectsLoadedMsg struct {
	keys   []string
	err    error
	bucket string
}

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
	footerH       int
	innerH        int // content height for sub-panels (inside borders)

	client      *s3helper.S3 // S3 client (nil if not configured)
	clientError error        // client init error
}

var _ common.Component = (*Browse)(nil)

// New creates a new Browse coordinator.
func New(client *s3helper.S3, clientErr error) *Browse {
	return &Browse{
		buckets:     buckets.New(),
		objects:     objects.New(),
		focus:       0,
		keys:        DefaultKeyMap,
		helpModel:   help.New(),
		client:      client,
		clientError: clientErr,
	}
}

// Init initialises both sub-panels and triggers bucket loading.
func (b *Browse) Init() tea.Cmd {
	cmds := []tea.Cmd{
		b.buckets.Init(),
		b.objects.Init(),
	}
	if b.client != nil {
		cmds = append(cmds, loadBuckets(b.client))
	}
	return tea.Batch(cmds...)
}

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

func loadObjects(client *s3helper.S3, bucket string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		objects, err := client.ListObjects(ctx, bucket, "")
		keys := make([]string, 0, len(objects))
		for _, o := range objects {
			keys = append(keys, o.Key)
		}
		return objectsLoadedMsg{keys: keys, err: err, bucket: bucket}
	}
}

// Resize computes the 2-column layout and forwards adjusted sizes to sub-panels.
func (b *Browse) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	b.lastWindow = ws
	b.helpModel, _ = b.helpModel.Update(ws)

	b.footerH = 1
	b.innerH = ws.Height - b.footerH - 3 // 3 = top border + banner + bottom border

	b.leftW = ws.Width * 25 / 100
	b.rightW = ws.Width - b.leftW

	leftWS := tea.WindowSizeMsg{Width: b.leftW - 2, Height: b.innerH}
	rightWS := tea.WindowSizeMsg{Width: b.rightW - 2, Height: b.innerH}

	return tea.Batch(
		b.buckets.Resize(leftWS),
		b.objects.Resize(rightWS),
	)
}

// View renders the header, two bordered panels, and footer.
func (b *Browse) View() string {
	if b.client == nil {
		msg := "No .env file configured."
		if b.clientError != nil {
			msg = "Error: " + b.clientError.Error()
		}
		return lipgloss.NewStyle().
			Width(b.lastWindow.Width).Height(b.lastWindow.Height).
			AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center).
			Render(msg + "\n\nPress Esc to go back")
	}

	middle := b.renderPanels()
	helpStr := b.helpModel.View(b.keys)
	return lipgloss.JoinVertical(lipgloss.Left, middle, helpStr)
}

// renderPanels builds the two bordered panels joined horizontally.
func (b *Browse) renderPanels() string {
	bucketView := b.wrapPanel(b.buckets.View(), b.leftW, b.focus == 0, "Buckets")
	objectView := b.wrapPanel(b.objects.View(), b.rightW, b.focus == 1, "Objects")
	return lipgloss.JoinHorizontal(lipgloss.Top, bucketView, objectView)
}

// wrapPanel surrounds content with a single bordered container with a banner title.
// Total panel height = b.innerH + 3 (top border + banner + content + bottom border).
func (b *Browse) wrapPanel(content string, w int, active bool, title string) string {
	activeColor := lipgloss.Color("63")
	inactiveColor := lipgloss.Color("240")

	borderColor := inactiveColor
	if active {
		borderColor = activeColor
	}

	innerW := w - 4 // 2 for border + 2 for Padding(0,1)

	bannerStyle := lipgloss.NewStyle().
		Width(innerW).
		Background(borderColor).
		Foreground(lipgloss.Color("255")).
		Bold(true).
		Padding(0, 1)

	banner := bannerStyle.Render(title)

	contentStyle := lipgloss.NewStyle().
		Width(innerW).
		Height(b.innerH)

	paddedContent := contentStyle.Render(content)

	inner := lipgloss.JoinVertical(lipgloss.Left, banner, paddedContent)

	return lipgloss.NewStyle().
		Width(w).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(inner)
}

// Update handles input: Esc returns to S3 menu, arrow/vim keys navigate,
// Left/Right switch panels, Enter loads objects for selected bucket,
// and async loading results are handled outside key handling.
func (b *Browse) Update(msg tea.Msg) tea.Cmd {
	// Handle async loading results first (before key handling).
	if msg, ok := msg.(bucketsLoadedMsg); ok {
		if msg.err == nil && len(msg.names) > 0 {
			b.buckets.SetItems(msg.names)
		}
		return nil
	}

	if msg, ok := msg.(objectsLoadedMsg); ok {
		if msg.err == nil && len(msg.keys) > 0 {
			b.objects.SetItems(msg.keys)
		}
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, b.keys.Esc) {
			return func() tea.Msg { return BackToS3MenuMsg{} }
		}
		if key.Matches(msg, b.keys.Up) {
			if b.focus == 0 {
				b.buckets.MoveUp()
			}
			if b.focus == 1 {
				b.objects.MoveUp()
			}
			return nil
		}
		if key.Matches(msg, b.keys.Down) {
			if b.focus == 0 {
				b.buckets.MoveDown()
			}
			if b.focus == 1 {
				b.objects.MoveDown()
			}
			return nil
		}
		if key.Matches(msg, b.keys.Left) {
			b.focus = 0
			return nil
		}
		if key.Matches(msg, b.keys.Right) {
			b.focus = 1
			return nil
		}
		if key.Matches(msg, b.keys.Enter) {
			if b.focus == 0 {
				selected := b.buckets.Selected()
				if selected != "" && b.client != nil {
					b.objects.SetItems([]string{})
					return loadObjects(b.client, selected)
				}
			}
			return nil
		}
		return nil

	default:
		// Broadcast non-key events to both panels (async data loading, etc.)
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
