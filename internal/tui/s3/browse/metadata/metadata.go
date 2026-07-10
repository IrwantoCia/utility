package metadata

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	s3helper "github.com/IrwantoCia/utility/internal/helper/s3"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// Metadata displays detailed information about the selected S3 object.
type Metadata struct {
	obj *s3helper.Object // nil when no object selected
}

// New creates a new Metadata display with no object selected.
func New() *Metadata {
	return &Metadata{}
}

// SetObject sets the object to display. Pass nil to clear.
func (m *Metadata) SetObject(obj *s3helper.Object) {
	m.obj = obj
}

// View renders the metadata panel content.
func (m *Metadata) View(width int) string {
	labelStyle := style.Default.BrowseMetaLabel
	valStyle := style.Default.BrowseMetaValue
	dimStyle := style.Default.BrowseMetaDim

	if m.obj == nil {
		return style.Default.BrowseEmpty.Render("Select an object to view metadata")
	}

	formatSize := func(size int64) string {
		const unit = 1024
		if size < unit {
			return fmt.Sprintf("%d B", size)
		}
		div, exp := int64(unit), 0
		for n := size / unit; n >= unit; n /= unit {
			div *= unit
			exp++
		}
		return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
	}

	row := func(label, value string) string {
		if value == "" {
			value = "-"
		}
		lbl := labelStyle.Render(label)
		val := valStyle.Render(value)
		return lbl + " " + val
	}

	ownerStr := "-"
	if m.obj.Owner.DisplayName != "" {
		ownerStr = m.obj.Owner.DisplayName
	} else if m.obj.Owner.ID != "" {
		ownerStr = m.obj.Owner.ID
	}

	// ── Section separators and titles ──
	sectionTitle := func(title string) string {
		return style.Default.BrowseMetaSection.Render("── " + title + " ──")
	}

	var rows []string

	// ── Object Details ──
	rows = append(rows, sectionTitle("Details"), "")
	rows = append(rows,
		row("Key:", m.obj.Key),
		row("Size:", formatSize(m.obj.Size)),
		row("Type:", m.obj.ContentType),
		row("ETag:", m.obj.ETag),
		row("Class:", m.obj.StorageClass),
		row("Modified:", m.obj.LastModified.Format("2006-01-02 15:04")),
	)

	// ── Owner ──
	rows = append(rows, "", sectionTitle("Owner"), "")
	rows = append(rows, row("Owner:", ownerStr))

	// ── Custom Metadata ──
	rows = append(rows, "", sectionTitle("Metadata"), "")
	if len(m.obj.Metadata) > 0 {
		for k, v := range m.obj.Metadata {
			kv := dimStyle.Render(fmt.Sprintf("  %s: %s", k, v))
			rows = append(rows, kv)
		}
	} else {
		rows = append(rows, dimStyle.Render("  No custom metadata"))
	}

	// Wrap the full content in a subtle background tint
	content := strings.Join(rows, "\n")
	return lipgloss.NewStyle().
		Width(width).
		Render(content)
}
