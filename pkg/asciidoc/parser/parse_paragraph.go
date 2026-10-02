package parser

import (
	"strings"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/lexer"
)

func (p *Parser) parseParagraph() *ast.Paragraph {
	start := tokenSpan(p.current).Start
	end := tokenSpan(p.current).End
	var tokens []lexer.LineToken
	for p.current.Kind == lexer.LineText {
		tokens = append(tokens, p.current)
		end = tokenSpan(p.current).End
		p.advance()
	}

	var normalized strings.Builder
	var original strings.Builder
	for index, token := range tokens {
		if index > 0 {
			normalized.WriteByte('\n')
		}
		normalized.WriteString(token.Raw)
		original.WriteString(token.Raw)
		if index+1 < len(tokens) {
			original.WriteString(token.Ending.Text())
		}
	}
	return &ast.Paragraph{
		Source:  ast.Span{Start: start, End: end},
		Text:    normalized.String(),
		Inlines: parseInlines(original.String(), start),
	}
}
