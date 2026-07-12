// Package reader provides a stateful RSVP token iterator driven
// by the TUI tick loop.
package reader

import (
	"strings"

	"github.com/IrwantoCia/utility/internal/helper/spritz"
	"github.com/IrwantoCia/utility/internal/helper/spritz/helper"
)

// Reader is a stateful iterator over pre-parsed tokens.
// It tracks position, WPM speed, and calculates frame intervals
// based on per-token PauseFactor.
type Reader struct {
	tokens       []spritz.Token
	chunks       []spritz.Chunk
	chunkMode    bool
	idx          int
	wpm          int
	baseInterval int // milliseconds: 60000 / wpm
}

// New creates a Reader with default 300 WPM.
func New() *Reader {
	return &Reader{wpm: 300, baseInterval: 200, idx: -1}
}

// Load pre-loads a token slice, chunkifies it, and resets position to the beginning.
func (r *Reader) Load(tokens []spritz.Token) {
	r.tokens = tokens
	r.chunks = helper.Chunkify(tokens, 4)
	r.idx = -1
}

// Current returns the current token and its index.
// If no token is active, returns zero Token and -1.
func (r *Reader) Current() (spritz.Token, int) {
	tok, idx, _ := r.currentToken()
	return tok, idx
}

// Next advances to the next token (or chunk, in chunk mode).
// Returns token, index, and a boolean indicating whether more items remain.
// On first call after Load, returns the first item.
func (r *Reader) Next() (spritz.Token, int, bool) {
	r.idx++
	if r.chunkMode && len(r.chunks) > 0 {
		if r.idx >= len(r.chunks) {
			r.idx = len(r.chunks) - 1
			return spritz.Token{}, r.idx, false
		}
		return r.chunkToToken(r.chunks[r.idx]), r.idx, true
	}
	if r.idx >= len(r.tokens) {
		r.idx = len(r.tokens) - 1
		return spritz.Token{}, r.idx, false
	}
	return r.tokens[r.idx], r.idx, true
}

// Prev rewinds to the previous token (or chunk, in chunk mode).
// Returns token, index, and whether the operation succeeded.
func (r *Reader) Prev() (spritz.Token, int, bool) {
	r.idx--
	if r.chunkMode && len(r.chunks) > 0 {
		if r.idx < 0 {
			r.idx = 0
			return spritz.Token{}, r.idx, false
		}
		return r.chunkToToken(r.chunks[r.idx]), r.idx, true
	}
	if r.idx < 0 {
		r.idx = 0
		return spritz.Token{}, r.idx, false
	}
	return r.tokens[r.idx], r.idx, true
}

// Reset jumps back to position -1 (before the first token).
func (r *Reader) Reset() {
	r.idx = -1
}

// ChunkMode returns whether chunked reading is active.
func (r *Reader) ChunkMode() bool {
	return r.chunkMode && len(r.chunks) > 0
}

// ToggleChunkMode switches between single-word and chunked reading.
func (r *Reader) ToggleChunkMode() {
	r.chunkMode = !r.chunkMode
	if r.chunkMode {
		r.chunks = helper.Chunkify(r.tokens, 4)
	}
	r.idx = -1
}

// currentToken returns the token for the current position,
// handling both single-word and chunked modes.
func (r *Reader) currentToken() (spritz.Token, int, bool) {
	if r.chunkMode && len(r.chunks) > 0 {
		if r.idx < 0 || r.idx >= len(r.chunks) {
			return spritz.Token{}, r.idx, false
		}
		return r.chunkToToken(r.chunks[r.idx]), r.idx, true
	}
	if r.idx < 0 || r.idx >= len(r.tokens) {
		return spritz.Token{}, r.idx, false
	}
	return r.tokens[r.idx], r.idx, true
}

// chunkToToken converts a Chunk into a synthetic Token for RSVP display.
// Word is the joined chunk words, ORP is the middle word's ORP char offset.
func (r *Reader) chunkToToken(c spritz.Chunk) spritz.Token {
	var parts []string
	for _, t := range c.Tokens {
		parts = append(parts, t.Word)
	}
	joined := strings.Join(parts, " ")

	offset := 0
	for i := 0; i < c.ORPWordIdx; i++ {
		offset += len(c.Tokens[i].Word) + 1 // +1 for space separator
	}
	orpIndex := offset + c.ORPCharIdx

	return spritz.Token{
		Word:        joined,
		ORPIndex:    orpIndex,
		PauseFactor: c.PauseFactor,
	}
}

// Progress returns the current position as a ratio (0.0 to 1.0).
func (r *Reader) Progress() float64 {
	total := r.Len()
	if total == 0 {
		return 0
	}
	return float64(r.idx+1) / float64(total)
}

// Len returns the total number of loaded items (tokens or chunks).
func (r *Reader) Len() int {
	if r.chunkMode && len(r.chunks) > 0 {
		return len(r.chunks)
	}
	return len(r.tokens)
}

// SetWPM adjusts the reading speed and recalculates base interval.
func (r *Reader) SetWPM(wpm int) {
	r.wpm = wpm
	r.baseInterval = 60000 / wpm
}

// Interval returns the display duration in milliseconds for the
// current token (or chunk), factoring in its PauseFactor.
func (r *Reader) Interval() int {
	if r.chunkMode && len(r.chunks) > 0 {
		if r.idx < 0 || r.idx >= len(r.chunks) {
			return r.baseInterval
		}
		return r.baseInterval * r.chunks[r.idx].PauseFactor
	}
	if r.idx < 0 || r.idx >= len(r.tokens) {
		return r.baseInterval
	}
	return r.baseInterval * r.tokens[r.idx].PauseFactor
}

// Token returns the token (or chunk-as-token) at absolute index idx
// without moving the cursor. Returns false if idx is out of bounds.
func (r *Reader) Token(idx int) (spritz.Token, bool) {
	if r.chunkMode && len(r.chunks) > 0 {
		if idx < 0 || idx >= len(r.chunks) {
			return spritz.Token{}, false
		}
		return r.chunkToToken(r.chunks[idx]), true
	}
	if idx < 0 || idx >= len(r.tokens) {
		return spritz.Token{}, false
	}
	return r.tokens[idx], true
}
