// Package buckets provides the left panel of the S3 browser, showing a
// navigable list of S3 bucket names.
package buckets

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

var (
	navUp   = key.NewBinding(key.WithKeys("k", "up"))
	navDown = key.NewBinding(key.WithKeys("j", "down"))
)

// Buckets renders a navigable list of S3 buckets.
type Buckets struct {
	items      []string
	cursor     int
	lastWindow tea.WindowSizeMsg
}

var _ common.Component = (*Buckets)(nil)

// New creates a Buckets panel with hardcoded demo data.
func New() *Buckets {
	return &Buckets{
		items: []string{"prod", "staging", "dev"},
	}
}

// Init is a no-op for the static bucket list.
func (b *Buckets) Init() tea.Cmd { return nil }

// Resize stores the window dimensions forwarded by the S3 coordinator.
func (b *Buckets) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	b.lastWindow = ws
	return nil
}

// View renders the bucket list with the cursor row highlighted.
func (b *Buckets) View() string {
	var sb strings.Builder

	for i, item := range b.items {
		cursor := "  "
		if i == b.cursor {
			cursor = "▸ "
		}

		line := cursor + item
		if i == b.cursor {
			line = style.Default.Highlighted.Render(line)
		}
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	// Pad to fill the panel height.
	targetH := b.lastWindow.Height
	if targetH > 0 {
		currentH := lipgloss.Height(sb.String())
		for j := currentH; j < targetH; j++ {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// Update handles keyboard navigation: k/up and j/down.
func (b *Buckets) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(keyMsg, navUp):
		if b.cursor > 0 {
			b.cursor--
		}
	case key.Matches(keyMsg, navDown):
		if b.cursor < len(b.items)-1 {
			b.cursor++
		}
	}

	return nil
}
