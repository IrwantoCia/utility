// Package markdown provides a markdown parser for Spritz RSVP reading.
package markdown

import (
	"os"
	"regexp"

	"github.com/IrwantoCia/utility/internal/helper/spritz"
	"github.com/IrwantoCia/utility/internal/helper/spritz/helper"
)

var _ spritz.Parser = (*Parser)(nil)

type Parser struct{}

func New() *Parser { return &Parser{} }

func (p *Parser) Parse(filePath string) ([]spritz.Token, error) {
	cachePath := helper.CachePath(filePath)
	if tokens, ok := helper.LoadCache(cachePath, filePath); ok {
		return tokens, nil
	}
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	plain := stripMarkdown(string(raw))
	tokens := helper.Tokenize(plain)
	helper.SaveCache(cachePath, tokens)

	helper.SaveContent(helper.ContentPath(filePath), []byte(plain))

	return tokens, nil
}

const (
	reFencedCodeBlock  = "(?s)```.*?```"
	reImage            = `!\[[^\]]*\]\([^)]*\)`
	reLink             = `\[([^\]]+)\]\([^)]+\)`
	reBold             = `\*\*([^*]+)\*\*`
	reItalic           = `\*([^*]+)\*`
	reStrikethrough    = `~~([^~]+)~~`
	reInlineCode       = "`([^`]+)`"
	reHeading          = `(?m)^#{1,6}\s+`
	reHorizontalRule   = `(?m)^[-*_]{3,}\s*$`
	reBlockquote       = `(?m)^>\s?`
	reUnorderedList    = `(?m)^[-*+]\s+`
	reOrderedList      = `(?m)^\d+\.\s+`
	reCollapseNewlines = `\n{3,}`
)

var stripSteps = []struct {
	re   *regexp.Regexp
	repl string
}{
	{regexp.MustCompile(reFencedCodeBlock), ""},
	{regexp.MustCompile(reImage), ""},
	{regexp.MustCompile(reLink), "$1"},
	{regexp.MustCompile(reBold), "$1"},
	{regexp.MustCompile(reItalic), "$1"},
	{regexp.MustCompile(reStrikethrough), "$1"},
	{regexp.MustCompile(reInlineCode), "$1"},
	{regexp.MustCompile(reHeading), ""},
	{regexp.MustCompile(reHorizontalRule), ""},
	{regexp.MustCompile(reBlockquote), ""},
	{regexp.MustCompile(reUnorderedList), ""},
	{regexp.MustCompile(reOrderedList), ""},
	{regexp.MustCompile(reCollapseNewlines), "\n\n"},
}

func stripMarkdown(raw string) string {
	s := raw
	for _, step := range stripSteps {
		s = step.re.ReplaceAllString(s, step.repl)
	}
	return s
}
