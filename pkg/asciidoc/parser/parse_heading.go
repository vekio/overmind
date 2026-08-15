package parser

import "git.casta.me/alberto/overmind/pkg/asciidoc/ast"

func (p *Parser) parseDocumentTitle() *ast.DocumentTitle {
	token := p.current
	title := &ast.DocumentTitle{
		Source:      tokenSpan(token),
		TitleSource: tokenFragmentSpan(token, token.Heading.TitleByteOffset, len(token.Heading.Title)),
		Text:        token.Heading.Title,
	}
	p.advance()
	return title
}
