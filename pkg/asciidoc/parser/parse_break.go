package parser

import "github.com/vekio/overmind/pkg/asciidoc/ast"

func (p *Parser) parseThematicBreak() *ast.ThematicBreak {
	result := &ast.ThematicBreak{Source: tokenSpan(p.current)}
	p.advance()
	return result
}

func (p *Parser) parsePageBreak() *ast.PageBreak {
	result := &ast.PageBreak{Source: tokenSpan(p.current)}
	p.advance()
	return result
}
