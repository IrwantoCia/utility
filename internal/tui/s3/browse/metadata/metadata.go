package metadata

import (
	"fmt"
	"strings"

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
	labelStyle := style.Default.HelpKey   // reuse: gray bold for labels
	valStyle := style.Default.CardDesc    // reuse: normal text
	dimStyle := style.Default.CardDesc    // reuse: dim text for dash

	if m.obj == nil {
		return dimStyle.Render("Select an object")
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
		lbl := labelStyle.Width(10).Render(label)
		val := valStyle.Render(value)
		return lbl + val
	}

	ownerStr := "-"
	if m.obj.Owner.DisplayName != "" {
		ownerStr = m.obj.Owner.DisplayName
	} else if m.obj.Owner.ID != "" {
		ownerStr = m.obj.Owner.ID
	}

	rows := []string{
		row("Key:", m.obj.Key),
		row("Size:", formatSize(m.obj.Size)),
		row("Type:", m.obj.ContentType),
		row("ETag:", m.obj.ETag),
		row("Class:", m.obj.StorageClass),
		row("Owner:", ownerStr),
		row("Modified:", m.obj.LastModified.Format("2006-01-02 15:04")),
	}

	// Custom metadata section
	if len(m.obj.Metadata) > 0 {
		rows = append(rows, "")
		metaLabel := style.Default.HelpKey.Render("Metadata:")
		rows = append(rows, metaLabel)
		indent := style.Default.CardDesc
		for k, v := range m.obj.Metadata {
			kv := indent.Render(fmt.Sprintf("  %s: %s", k, v))
			rows = append(rows, kv)
		}
	} else {
		rows = append(rows, "", row("Metadata:", "-"))
	}

	return strings.Join(rows, "\n")
}
