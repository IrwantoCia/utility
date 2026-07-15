// Package whisper provides model discovery for whisper.cpp .bin files.
package whisper

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Model represents a discovered whisper.cpp model file.
type Model struct {
	Name string // e.g. "small", "medium", "large-v3"
	Path string // full path to the .bin file
	Size int64  // file size in bytes
}

// DefaultModelDir is the hardcoded path for whisper model storage.
const DefaultModelDir = "~/.config/utility/whisper/models"

// ScanModels scans baseDir for ggml-*.bin files and returns parsed models.
// baseDir supports ~ expansion. Returns empty slice if directory doesn't exist.
func ScanModels(baseDir string) ([]Model, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(baseDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home dir: %w", err)
		}
		baseDir = filepath.Join(home, baseDir[1:])
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading dir %s: %w", baseDir, err)
	}

	var models []Model
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "ggml-") || !strings.HasSuffix(name, ".bin") {
			continue
		}

		// Parse model name: ggml-<name>.bin -> <name>
		modelName := strings.TrimPrefix(name, "ggml-")
		modelName = strings.TrimSuffix(modelName, ".bin")

		info, err := entry.Info()
		if err != nil {
			// Skip files we can't stat
			continue
		}

		models = append(models, Model{
			Name: modelName,
			Path: filepath.Join(baseDir, name),
			Size: info.Size(),
		})
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	return models, nil
}
