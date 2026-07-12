// Package helper provides shared tokenization and caching utilities
// for Spritz parsers.
package helper

import (
	"encoding/json"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/IrwantoCia/utility/internal/helper/spritz"
)

// Tokenize splits raw plain text into tokens with ORP indices and pause factors.
// Paragraph breaks (\n\n) set PauseFactor=3, sentence-ending punctuation
// (.!?;:) sets PauseFactor=2.
func Tokenize(raw string) []spritz.Token {
	paragraphs := strings.Split(raw, "\n\n")
	tokens := make([]spritz.Token, 0, len(paragraphs)*8)

	for pi, para := range paragraphs {
		words := strings.Fields(para)
		for wi, word := range words {
			if word == "" {
				continue
			}
			orp := calcORP(word)
			pause := 1

			if endsWithSentencePunct(word) {
				pause = 2
			}
			if pi > 0 && wi == 0 {
				pause = 3
			}

			tokens = append(tokens, spritz.Token{
				Word:        word,
				ORPIndex:    orp,
				PauseFactor: pause,
			})
		}
	}
	return tokens
}

const sentencePunctuation = ".!?;:"

func endsWithSentencePunct(word string) bool {
	if len(word) == 0 {
		return false
	}
	last := word[len(word)-1]
	return strings.ContainsRune(sentencePunctuation, rune(last))
}

// calcORP returns the 0-based Optimal Recognition Point index for word.
// Formula: ceil(len(word) * 0.3), clamped to [0, len(word)-1].
func calcORP(word string) int {
	n := len(word)
	if n == 0 {
		return 0
	}
	orp := int(math.Ceil(float64(n) * 0.3))
	if orp >= n {
		orp = n - 1
	}
	return orp
}

// LoadCache attempts to load tokens from cachePath. Returns (tokens, true)
// when the cache is valid (exists and not older than sourcePath).
func LoadCache(cachePath, sourcePath string) ([]spritz.Token, bool) {
	cacheStat, err := os.Stat(cachePath)
	if err != nil {
		return nil, false
	}
	sourceStat, err := os.Stat(sourcePath)
	if err != nil {
		return nil, false
	}
	if cacheStat.ModTime().Before(sourceStat.ModTime()) {
		return nil, false
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}

	var tokens []spritz.Token
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, false
	}
	return tokens, true
}

// CachePath returns the cache file path for a source file.
//
// Layout: .spritz/<relative-dir>/<folder>/<filename>.spritz.json
// Falls back to .spritz/<folder>/<filename>.spritz.json when relative path
// cannot be resolved.
func CachePath(filePath string) string {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		name := filepath.Base(filePath)
		folder := nameNoExt(name)
		return filepath.Join(".spritz", folder, name+".spritz.json")
	}
	cwd, err := os.Getwd()
	if err != nil {
		name := filepath.Base(abs)
		folder := nameNoExt(name)
		return filepath.Join(".spritz", folder, name+".spritz.json")
	}
	rel, err := filepath.Rel(cwd, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		name := filepath.Base(abs)
		folder := nameNoExt(name)
		return filepath.Join(".spritz", folder, name+".spritz.json")
	}

	dir := filepath.Dir(rel)
	name := filepath.Base(rel)
	folder := nameNoExt(name)

	if dir == "." {
		return filepath.Join(".spritz", folder, name+".spritz.json")
	}
	return filepath.Join(".spritz", dir, folder, name+".spritz.json")
}

// ContentPath returns the path where plain text content is cached.
//
// Layout: .spritz/<relative-dir>/<folder>/<filename>
func ContentPath(filePath string) string {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		name := filepath.Base(filePath)
		folder := nameNoExt(name)
		return filepath.Join(".spritz", folder, name)
	}
	cwd, err := os.Getwd()
	if err != nil {
		name := filepath.Base(abs)
		folder := nameNoExt(name)
		return filepath.Join(".spritz", folder, name)
	}
	rel, err := filepath.Rel(cwd, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		name := filepath.Base(abs)
		folder := nameNoExt(name)
		return filepath.Join(".spritz", folder, name)
	}

	dir := filepath.Dir(rel)
	name := filepath.Base(rel)
	folder := nameNoExt(name)

	if dir == "." {
		return filepath.Join(".spritz", folder, name)
	}
	return filepath.Join(".spritz", dir, folder, name)
}

// nameNoExt returns the filename without its extension.
func nameNoExt(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return name
	}
	return strings.TrimSuffix(name, ext)
}

// SaveCache marshals tokens as indented JSON and writes them to cachePath.
// Best-effort — errors are silently discarded.
func SaveCache(cachePath string, tokens []spritz.Token) {
	data, _ := json.MarshalIndent(tokens, "", "  ")
	_ = os.MkdirAll(filepath.Dir(cachePath), 0755)
	_ = os.WriteFile(cachePath, data, 0644)
}

// SaveContent writes content bytes to the given path, creating
// parent directories as needed. Best-effort — errors are silently
// discarded.
func SaveContent(path string, content []byte) {
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	_ = os.WriteFile(path, content, 0644)
}

// ListCached walks the .spritz/ cache directory and returns the
// reconstructed source file paths for all cached entries. Returns
// (empty slice, nil) when .spritz/ does not exist.
func ListCached() ([]string, error) {
	var sources []string
	err := filepath.Walk(".spritz", func(walkPath string, info fs.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".spritz.json") {
			return nil
		}

		// Walk path: .spritz/<rel-dir>/<name-no-ext>/<name>.spritz.json
		// Reconstruct source: <rel-dir>/<name>
		rel := strings.TrimPrefix(walkPath, ".spritz/")
		rel = strings.TrimSuffix(rel, ".spritz.json")

		dir := filepath.Dir(rel)
		file := filepath.Base(rel)

		parentDir := filepath.Dir(dir)
		if parentDir == "." {
			sources = append(sources, file)
		} else {
			sources = append(sources, filepath.Join(parentDir, file))
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return sources, nil
		}
		return nil, err
	}
	return sources, nil
}

// Chunkify groups tokens into phrase-aware chunks (2 to maxWords words each).
// Breaks occur on: sentence end (.!?), clause punctuation (,:;), or conjunctions (and/but/or/nor/yet/so)
// when the chunk already has 2+ words. Last chunk may be any size.
func Chunkify(tokens []spritz.Token, maxWords int) []spritz.Chunk {
	if maxWords < 2 {
		maxWords = 2
	}
	var chunks []spritz.Chunk
	var current []spritz.Token

	for i, tok := range tokens {
		current = append(current, tok)
		isLast := i == len(tokens)-1

		shouldBreak := false

		if len(current) >= maxWords {
			shouldBreak = true
		} else if len(current) >= 2 {
			word := strings.TrimRight(tok.Word, ".,;:!?\"'")
			lastRune := rune(tok.Word[len(tok.Word)-1])

			if lastRune == '.' || lastRune == '!' || lastRune == '?' {
				shouldBreak = true
			}
			if lastRune == ',' || lastRune == ';' || lastRune == ':' {
				shouldBreak = true
			}
			lower := strings.ToLower(word)
			switch lower {
			case "and", "but", "or", "nor", "yet", "so",
				"that", "which", "who", "whom", "whose",
				"if", "when", "where", "because", "since", "although", "while":
				shouldBreak = true
			}
		}

		if shouldBreak || isLast {
			if len(current) > 0 {
				chunks = append(chunks, makeChunk(current))
				current = nil
			}
		}
	}

	return chunks
}

func makeChunk(tokens []spritz.Token) spritz.Chunk {
	orpIdx := len(tokens) / 2
	maxPause := 1
	for _, t := range tokens {
		if t.PauseFactor > maxPause {
			maxPause = t.PauseFactor
		}
	}
	return spritz.Chunk{
		Tokens:      tokens,
		ORPWordIdx:  orpIdx,
		ORPCharIdx:  tokens[orpIdx].ORPIndex,
		PauseFactor: maxPause,
	}
}
