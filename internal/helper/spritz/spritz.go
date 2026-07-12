// Package spritz provides text parsing and RSVP reading primitives.
package spritz

// Token represents a single word unit ready for RSVP display.
// It includes ORP alignment info and pause timing hints.
type Token struct {
	Word        string `json:"word"`   // word including trailing punctuation ("Hello,")
	ORPIndex    int    `json:"orp"`    // character index of the Optimal Recognition Point (highlighted red during display)
	PauseFactor int    `json:"pause"`  // 1=normal, 2=sentence end, 3=paragraph break
}

// Chunk represents a group of 2-4 adjacent tokens displayed as a single RSVP frame.
// The ORP anchor points to the middle word of the chunk.
type Chunk struct {
	Tokens      []Token `json:"tokens"`
	ORPWordIdx  int     `json:"orp_word"` // index of ORP word within Tokens (typically len/2)
	ORPCharIdx  int     `json:"orp_char"` // char index within the ORP word (from Token.ORPIndex)
	PauseFactor int     `json:"pause"`    // max PauseFactor across all tokens in chunk
}

// Parser reads a text/markdown file, strips markup, tokenizes,
// and caches the result as {filename}.spritz.json alongside the source.
type Parser interface {
	Parse(filePath string) ([]Token, error)
}

