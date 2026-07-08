package menu

import (
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/IrwantoCia/utility/internal/tui/style"
)

// envVar describes a single S3-related environment variable.
type envVar struct {
	name     string
	value    string // raw from .env, empty if not set
	isSet    bool
	isSecret bool
	example  string
}

// requiredEnvVars defines the S3_* variables that the TUI inspects.
var requiredEnvVars = []envVar{
	{name: "S3_PROVIDER", isSecret: false, example: "b2"},
	{name: "S3_ENDPOINT", isSecret: false, example: "s3.us-west-004.backblazeb2.com"},
	{name: "S3_ACCESS_KEY", isSecret: true, example: "required"},
	{name: "S3_SECRET_KEY", isSecret: true, example: "required"},
	{name: "S3_SECURE", isSecret: false, example: "true"},
}

// EnvInfo holds parsed S3-related environment variables and can render
// them as a bordered info section independent of any particular TUI page.
type EnvInfo struct {
	Vars []envVar
}

// Load reads the given .env file and populates e.Vars.
// If envFile is empty the Vars slice still contains the required
// variable definitions, but every entry has isSet == false.
func (e *EnvInfo) Load(envFile string) {
	// Copy required env vars as a starting list (reset state).
	e.Vars = make([]envVar, len(requiredEnvVars))
	copy(e.Vars, requiredEnvVars)

	if envFile == "" {
		return
	}

	data, err := os.ReadFile(envFile)
	if err != nil {
		return
	}

	envMap := make(map[string]string)
	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envMap[key] = value
	}

	for i, ev := range e.Vars {
		if val, ok := envMap[ev.name]; ok && val != "" {
			e.Vars[i].value = val
			e.Vars[i].isSet = true
		}
	}
}

// View renders a bordered "S3 Environment" box of the given width.
func (e *EnvInfo) View(width int) string {
	innerWidth := width - 4                          // border (2) + padding left/right (2)
	maxValueWidth := innerWidth - 2 - 1 - 2 - 15 - 2 // indent + status + gap + key + gap
	maxValueWidth = max(maxValueWidth, 10)

	var rows []string

	rows = append(rows, style.Default.EnvSectionTitle.Render("S3 Environment"))
	rows = append(rows, "")

	for _, ev := range e.Vars {
		var valStr string
		var valStyle lipgloss.Style

		if ev.isSet {
			if ev.isSecret {
				valStr = "****"
				valStyle = style.Default.EnvValueMasked
			} else {
				valStr = ev.value
				valStyle = style.Default.EnvValue
			}
		} else {
			if ev.example == "required" {
				valStr = "(required)"
			} else {
				valStr = "(ex: " + ev.example + ")"
			}
			valStyle = style.Default.EnvValueExample
		}

		// Truncate value if too long (skip mask which is always 4 chars)
		if valStr != "****" && len(valStr) > maxValueWidth {
			valStr = valStr[:max(0, maxValueWidth-1)] + "…"
		}

		statusStyle := style.Default.EnvStatusMissing
		statusDot := "○"
		if ev.isSet {
			statusStyle = style.Default.EnvStatusSet
			statusDot = "●"
		}

		keyStr := style.Default.EnvKey.Render(ev.name)
		valStr = valStyle.Render(valStr)
		statusStr := statusStyle.Render(statusDot)

		row := lipgloss.JoinHorizontal(lipgloss.Left,
			"  ",
			statusStr,
			"  ",
			keyStr,
			"  ",
			valStr,
		)
		rows = append(rows, row)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return style.Default.EnvSectionBorder.
		Width(width).
		Render(content)
}
