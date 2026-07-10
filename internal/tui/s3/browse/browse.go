// Package browse provides the 2-panel S3 browser coordinator
// (buckets on the left, objects on the right).
package browse

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	s3helper "github.com/IrwantoCia/utility/internal/helper/s3"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/s3/browse/buckets"
	"github.com/IrwantoCia/utility/internal/tui/s3/browse/metadata"
	"github.com/IrwantoCia/utility/internal/tui/s3/browse/objects"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// BackToS3MenuMsg tells the S3 coordinator to return to the S3 sub-menu.
type BackToS3MenuMsg struct{}

type bucketsLoadedMsg struct {
	names []string
	err   error
}

type objectsLoadedMsg struct {
	keys    []string
	objects []s3helper.Object // full objects for metadata
	err     error
	bucket  string
}

type panelFocus int

const (
	focusBuckets panelFocus = iota
	focusObjects
)

// Browse coordinates the 2-panel S3 browser (buckets + objects).
type Browse struct {
	buckets *buckets.Buckets
	objects *objects.Objects

	metadata   *metadata.Metadata
	allObjects []s3helper.Object // full object list for metadata lookup

	focus      panelFocus
	lastWindow tea.WindowSizeMsg
	keys       KeyMap
	helpModel  help.Model

	// Computed layout dimensions.
	leftW, midW, rightW int
	footerH             int
	innerH              int // content height for sub-panels (inside borders)

	client      *s3helper.S3 // S3 client (nil if not configured)
	clientError error        // client init error
}

var _ common.Component = (*Browse)(nil)

// New creates a new Browse coordinator.
func New(client *s3helper.S3, clientErr error) *Browse {
	hm := help.New()
	hm.Styles = BrowseHelpStyles()
	return &Browse{
		buckets:     buckets.New(),
		objects:     objects.New(),
		metadata:    metadata.New(),
		focus:       focusBuckets,
		keys:        DefaultKeyMap,
		helpModel:   hm,
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
		result, err := client.ListObjects(ctx, bucket, "")
		keys := make([]string, 0, len(result))
		for _, o := range result {
			keys = append(keys, o.Key)
		}
		return objectsLoadedMsg{keys: keys, objects: result, err: err, bucket: bucket}
	}
}

// Resize computes the 2-column layout and forwards adjusted sizes to sub-panels.
func (b *Browse) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	b.lastWindow = ws
	b.helpModel, _ = b.helpModel.Update(ws)

	b.footerH = 1
	b.innerH = ws.Height - b.footerH - 3 // 3 = top border + banner + bottom border

	b.leftW  = ws.Width * 25 / 100
	b.midW   = ws.Width * 45 / 100
	b.rightW = ws.Width - b.leftW - b.midW

	leftWS  := tea.WindowSizeMsg{Width: b.leftW - 4, Height: b.innerH}
	midWS   := tea.WindowSizeMsg{Width: b.midW - 4, Height: b.innerH}

	return tea.Batch(
		b.buckets.Resize(leftWS),
		b.objects.Resize(midWS),
	)
}

// View renders the header, two bordered panels, and footer.
func (b *Browse) View() string {
	if b.client == nil {
		msg := "No .env file configured."
		if b.clientError != nil {
			msg = "Error: " + b.clientError.Error()
		}
		errStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Bold(true)
		hintStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
		return lipgloss.NewStyle().
			Width(b.lastWindow.Width).Height(b.lastWindow.Height).
			AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center).
			Render(errStyle.Render("✖  S3 Client Unavailable") +
				"\n\n" + hintStyle.Render(msg) +
				"\n\n" + hintStyle.Render("Press Esc to go back"))
	}

	middle := b.renderPanels()
	helpStr := b.helpModel.View(b.keys)
	return lipgloss.JoinVertical(lipgloss.Left, middle, helpStr)
}

// renderPanels builds the three bordered panels joined horizontally.
func (b *Browse) renderPanels() string {
	bucketTitle := "Buckets"
	objectTitle := "Objects"
	if b.objects.FilterActive() {
		objectTitle = fmt.Sprintf("Objects  🔍  %s  [%d/%d]", b.objects.Filter(), len(b.objects.Items()), b.objects.TotalItems())
	} else if b.objects.Filter() != "" {
		objectTitle = fmt.Sprintf("Objects  ◂ %s  [%d/%d]", b.objects.Filter(), len(b.objects.Items()), b.objects.TotalItems())
	}

	bucketView   := b.wrapPanel(b.buckets.View(), b.leftW, b.focus == focusBuckets, false, bucketTitle)
	objectView   := b.wrapPanel(b.objects.View(), b.midW, b.focus == focusObjects, b.objects.FilterActive(), objectTitle)
	metadataView := b.wrapPanel(b.metadata.View(b.rightW-4), b.rightW, false, false, "Metadata")
	return lipgloss.JoinHorizontal(lipgloss.Top, bucketView, objectView, metadataView)
}

// syncMetadata updates the metadata panel to show the currently selected object.
func (b *Browse) syncMetadata() {
	selected := b.objects.Selected()
	if selected == "" || len(b.allObjects) == 0 {
		b.metadata.SetObject(nil)
		return
	}
	for i := range b.allObjects {
		if b.allObjects[i].Key == selected {
			b.metadata.SetObject(&b.allObjects[i])
			return
		}
	}
	b.metadata.SetObject(nil)
}

// wrapPanel surrounds content with a single bordered container with a title banner.
// When active, the panel uses bright cyan accent; when inactive, muted gray.
// When filterActive is true, a bright magenta border is used regardless of focus.
func (b *Browse) wrapPanel(content string, w int, active bool, filterActive bool, title string) string {
	innerW := w - 4 // 2 for border + 2 for Padding(0,1)

	bannerStyle := lipgloss.NewStyle()
	borderStyle := lipgloss.NewStyle()

	switch {
	case filterActive:
		borderStyle = style.Default.BrowseFilterBorder
		bannerStyle = lipgloss.NewStyle().
			Width(innerW).
			Background(lipgloss.Color("206")).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1)
	case active:
		borderStyle = style.Default.BrowseBorderActive
		bannerStyle = lipgloss.NewStyle().
			Width(innerW).
			Background(lipgloss.Color("45")).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1)
	default:
		borderStyle = style.Default.BrowseBorderInactive
		bannerStyle = lipgloss.NewStyle().
			Width(innerW).
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("250")).
			Padding(0, 1)
	}

	// Prefix icon for focused panel title
	prefix := ""
	if active {
		prefix = "◈ "
	}
	banner := bannerStyle.Render(prefix + title)

	contentStyle := lipgloss.NewStyle().
		Width(innerW).
		Height(b.innerH)

	paddedContent := contentStyle.Render(content)

	inner := lipgloss.JoinVertical(lipgloss.Left, banner, paddedContent)

	return lipgloss.NewStyle().
		Width(w).
		Inherit(borderStyle).
		Padding(0, 1).
		Render(inner)
}

// Update handles input: Esc returns to S3 menu, arrow/vim keys navigate,
// Left/Right switch panels, Enter loads objects for selected bucket,
// and async loading results are handled outside key handling.
func (b *Browse) Update(msg tea.Msg) tea.Cmd {
	// Handle async loading results first (before key handling).
	if msg, ok := msg.(bucketsLoadedMsg); ok {
		if msg.err != nil {
			b.buckets.SetStatus("Error: " + msg.err.Error())
			return nil
		}
		if len(msg.names) == 0 {
			b.buckets.SetStatus("No buckets found")
			return nil
		}
		b.buckets.SetItems(msg.names)
		return nil
	}

	if msg, ok := msg.(objectsLoadedMsg); ok {
		b.allObjects = msg.objects
		if msg.err == nil && len(msg.keys) > 0 {
			b.objects.SetItems(msg.keys)
		}
		b.syncMetadata()
		return nil
	}

	// ── Filter mode for objects panel ──────────────────────────────
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if b.objects.FilterActive() {
			if key.Matches(keyMsg, b.keys.Esc) {
				b.objects.ClearFilter()
				return nil
			}
			if key.Matches(keyMsg, b.keys.Backspace) {
				b.objects.DeleteFilter()
				return nil
			}
			if keyMsg.Text != "" {
				for _, r := range keyMsg.Text {
					b.objects.AppendFilter(r)
				}
				return nil
			}
			// Any other key: exit filter edit mode, fall through to normal handling
			b.objects.ExitFilterMode()
		} else {
			if key.Matches(keyMsg, b.keys.Filter) && b.focus == focusObjects && b.objects.TotalItems() > 0 {
				b.objects.EnterFilter()
				return nil
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, b.keys.Esc) {
			return func() tea.Msg { return BackToS3MenuMsg{} }
		}
		if key.Matches(msg, b.keys.Up) {
			if b.focus == focusBuckets {
				b.buckets.MoveUp()
			}
			if b.focus == focusObjects {
				b.objects.MoveUp()
				b.syncMetadata()
			}
			return nil
		}
		if key.Matches(msg, b.keys.Down) {
			if b.focus == focusBuckets {
				b.buckets.MoveDown()
			}
			if b.focus == focusObjects {
				b.objects.MoveDown()
				b.syncMetadata()
			}
			return nil
		}
		if key.Matches(msg, b.keys.PgUp) {
			if b.focus == focusBuckets {
				b.buckets.PageUp()
			}
			if b.focus == focusObjects {
				b.objects.PageUp()
				b.syncMetadata()
			}
			return nil
		}
		if key.Matches(msg, b.keys.PgDown) {
			if b.focus == focusBuckets {
				b.buckets.PageDown()
			}
			if b.focus == focusObjects {
				b.objects.PageDown()
				b.syncMetadata()
			}
			return nil
		}
		if key.Matches(msg, b.keys.Left) {
			b.focus = focusBuckets
			return nil
		}
		if key.Matches(msg, b.keys.Right) {
			b.focus = focusObjects
			return nil
		}
		if key.Matches(msg, b.keys.Enter) {
			if b.focus == focusBuckets {
				selected := b.buckets.Selected()
				if selected != "" && b.client != nil {
					b.objects.SetItems([]string{})
					b.focus = focusObjects
					b.allObjects = nil
					b.metadata.SetObject(nil)
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
