// Package filepicker provides the TUI component for file selection.
package filepicker

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"time"
	"unsafe"

	fp "charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/IrwantoCia/utility/internal/tui/common"
)

type clearErrorMsg struct{}

type FilePicker struct {
	Model        fp.Model
	SelectedFile string
	err          error
	keys         KeyMap
	helpModel    help.Model
}

var _ common.Component = (*FilePicker)(nil)

// Close implements common.Component.
func (f *FilePicker) Close() tea.Cmd { return nil }

func New(allowedTypes ...string) *FilePicker {
	m := fp.New()
	m.CurrentDirectory, _ = os.Getwd()
	m.ShowHidden = true
	m.ShowPermissions = false
	m.KeyMap.Back = key.NewBinding(
		key.WithKeys("h", "backspace", "left"),
		key.WithHelp("h", "back"),
	)

	if len(allowedTypes) > 0 {
		m.AllowedTypes = allowedTypes
	}

	return &FilePicker{
		Model:     m,
		keys:      DefaultKeyMap,
		helpModel: help.New(),
	}
}

func (f *FilePicker) Init() tea.Cmd {
	return f.Model.Init()
}

func (f *FilePicker) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	f.helpModel, _ = f.helpModel.Update(ws)
	var cmd tea.Cmd
	f.Model, cmd = f.Model.Update(ws)
	return cmd
}

func (f *FilePicker) canSelect(filename string) bool {
	if len(f.Model.AllowedTypes) == 0 {
		return true
	}
	for _, ext := range f.Model.AllowedTypes {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

func (f *FilePicker) filterFiles() {
	rv := reflect.ValueOf(&f.Model).Elem()
	filesField := rv.FieldByName("files")
	if !filesField.IsValid() {
		return
	}
	if len(f.Model.AllowedTypes) == 0 {
		return
	}

	// Read unexported field via unsafe pointer (bypasses Go 1.26 Interface() restriction)
	entries := *(*[]os.DirEntry)(unsafe.Pointer(filesField.UnsafeAddr()))

	filtered := make([]os.DirEntry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || f.canSelect(e.Name()) {
			filtered = append(filtered, e)
		}
	}

	// Write back via unsafe pointer
	*(*[]os.DirEntry)(unsafe.Pointer(filesField.UnsafeAddr())) = filtered

	// Clamp selected field to prevent out-of-bounds
	selField := rv.FieldByName("selected")
	if selField.IsValid() {
		sel := *(*int)(unsafe.Pointer(selField.UnsafeAddr()))
		if sel >= len(filtered) {
			*(*int)(unsafe.Pointer(selField.UnsafeAddr())) = 0
		}
	}
}

func (f *FilePicker) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case clearErrorMsg:
		f.err = nil
	}

	var cmd tea.Cmd
	f.Model, cmd = f.Model.Update(msg)

	f.filterFiles()

	if didSelect, path := f.Model.DidSelectFile(msg); didSelect {
		f.SelectedFile = path
	}

	if didSelect, path := f.Model.DidSelectDisabledFile(msg); didSelect {
		f.err = errors.New(path + " is not valid.")
		f.SelectedFile = ""
		return tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return cmd
}

func (f *FilePicker) View() string {
	var s strings.Builder
	s.WriteString("Please select:\n\n")
	if f.err != nil {
		s.WriteString(f.Model.Styles.DisabledFile.Render(f.err.Error()))
		s.WriteString("\n\n")
	}
	s.WriteString(f.Model.View())
	s.WriteString("\n\n")
	s.WriteString(f.helpModel.View(f.keys))
	return s.String()
}

func clearErrorAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}
