package parser

import (
	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/lexer"
)

func (p *Parser) parseAdmonition(metadata ast.BlockMetadata) *ast.Admonition {
	token := p.current
	labelSource := tokenFragmentSpan(token, token.Admonition.LabelByteOffset, len(token.Admonition.Label))
	contentSource := tokenFragmentSpan(token, token.Admonition.ContentByteOffset, len(token.Admonition.Content))
	p.advance()
	text, inlines, completeContentSource := p.parseTextContinuation(token.Admonition.Content, contentSource, token.Ending)
	return &ast.Admonition{
		Source:        sourceWithMetadata(ast.Span{Start: tokenSpan(token).Start, End: completeContentSource.End}, metadata),
		LabelSource:   labelSource,
		ContentSource: completeContentSource,
		Kind:          admonitionKind(token.Admonition.Kind),
		Label:         token.Admonition.Label,
		Text:          text,
		Inlines:       inlines,
		Metadata:      metadata,
	}
}

func admonitionKind(kind lexer.AdmonitionKind) ast.AdmonitionKind {
	switch kind {
	case lexer.AdmonitionNote:
		return ast.AdmonitionNote
	case lexer.AdmonitionTip:
		return ast.AdmonitionTip
	case lexer.AdmonitionImportant:
		return ast.AdmonitionImportant
	case lexer.AdmonitionCaution:
		return ast.AdmonitionCaution
	case lexer.AdmonitionWarning:
		return ast.AdmonitionWarning
	default:
		return ast.AdmonitionUnknown
	}
}
