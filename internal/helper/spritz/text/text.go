// Package text provides a plain-text parser for Spritz RSVP reading.
package text

import (
	"os"

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
	tokens := helper.Tokenize(string(raw))
	helper.SaveCache(cachePath, tokens)

	helper.SaveContent(helper.ContentPath(filePath), raw)

	return tokens, nil
}
