// Package reader provides a stateful RSVP token iterator driven
// by the TUI tick loop.
package reader

import (
	"github.com/IrwantoCia/utility/internal/helper/spritz"
)

// Reader is a stateful iterator over pre-parsed tokens.
// It tracks position, WPM speed, and calculates frame intervals
// based on per-token PauseFactor.
type Reader struct {
	tokens       []spritz.Token
	idx          int
	wpm          int
	baseInterval int // milliseconds: 60000 / wpm
}

// New creates a Reader with default 300 WPM.
func New() *Reader {
	return &Reader{wpm: 300, baseInterval: 200, idx: -1}
}

// Load pre-loads a token slice and resets position to the beginning.
func (r *Reader) Load(tokens []spritz.Token) {
	r.tokens = tokens
	r.idx = -1
}

// Current returns the current token and its index.
// If no token is active, returns zero Token and -1.
func (r *Reader) Current() (spritz.Token, int) {
	if r.idx < 0 || r.idx >= len(r.tokens) {
		return spritz.Token{}, -1
	}
	return r.tokens[r.idx], r.idx
}

// Next advances to the next token. Returns token, index, and
// a boolean indicating whether more tokens remain.
// On first call after Load, returns tokens[0].
func (r *Reader) Next() (spritz.Token, int, bool) {
	r.idx++
	if r.idx >= len(r.tokens) {
		r.idx = len(r.tokens) - 1
		return spritz.Token{}, r.idx, false
	}
	return r.tokens[r.idx], r.idx, true
}

// Prev rewinds to the previous token.
// Returns token, index, and whether the operation succeeded.
func (r *Reader) Prev() (spritz.Token, int, bool) {
	r.idx--
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

// Progress returns the current position as a ratio (0.0 to 1.0).
func (r *Reader) Progress() float64 {
	if len(r.tokens) == 0 {
		return 0
	}
	return float64(r.idx+1) / float64(len(r.tokens))
}

// Len returns the total number of loaded tokens.
func (r *Reader) Len() int {
	return len(r.tokens)
}

// SetWPM adjusts the reading speed and recalculates base interval.
func (r *Reader) SetWPM(wpm int) {
	r.wpm = wpm
	r.baseInterval = 60000 / wpm
}

// Interval returns the display duration in milliseconds for the
// current token, factoring in its PauseFactor.
func (r *Reader) Interval() int {
	if r.idx < 0 || r.idx >= len(r.tokens) {
		return r.baseInterval
	}
	return r.baseInterval * r.tokens[r.idx].PauseFactor
}
