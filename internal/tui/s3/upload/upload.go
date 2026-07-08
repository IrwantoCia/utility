// Package upload provides a placeholder upload page for the S3 workflow.
package upload

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
)

// BackToS3MenuMsg tells the S3 coordinator to return to the S3 sub-menu.
type BackToS3MenuMsg struct{}

var esc = key.NewBinding(key.WithKeys("esc"))

// Upload is the placeholder upload page.
type Upload struct {
	lastWindow tea.WindowSizeMsg
}

var _ common.Component = (*Upload)(nil)

// New creates a new Upload placeholder page.
func New() *Upload {
	return &Upload{}
}

// Init is a no-op for the placeholder.
func (u *Upload) Init() tea.Cmd { return nil }

// Resize stores the window dimensions.
func (u *Upload) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	u.lastWindow = ws
	return nil
}

// View renders a centred placeholder message.
func (u *Upload) View() string {
	content := "Upload page\n\n(placeholder)"
	w := u.lastWindow.Width
	h := u.lastWindow.Height
	if w <= 0 || h <= 0 {
		return content
	}

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(content)
}

// Update handles Esc to return to the S3 sub-menu.
func (u *Upload) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	if key.Matches(keyMsg, esc) {
		return func() tea.Msg {
			return BackToS3MenuMsg{}
		}
	}

	return nil
}
