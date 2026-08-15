package parser

import (
	"strings"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/lexer"
)

func (p *Parser) parseTextContinuation(initial string, source ast.Span, ending lexer.LineEnding) (string, []ast.Inline, ast.Span) {
	var normalized strings.Builder
	var original strings.Builder
	normalized.WriteString(initial)
	original.WriteString(initial)
	end := source.End
	previousEnding := ending
	for p.current.Kind == lexer.LineText {
		normalized.WriteByte('\n')
		normalized.WriteString(p.current.Raw)
		original.WriteString(previousEnding.Text())
		original.WriteString(p.current.Raw)
		end = tokenSpan(p.current).End
		previousEnding = p.current.Ending
		p.advance()
	}
	completeSource := ast.Span{Start: source.Start, End: end}
	return normalized.String(), parseInlines(original.String(), source.Start), completeSource
}

func sourceWithMetadata(source ast.Span, metadata ast.BlockMetadata) ast.Span {
	if !metadataEmpty(metadata) {
		source.Start = metadata.Source.Start
	}
	return source
}
