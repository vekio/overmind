package parser

import "git.casta.me/alberto/overmind/pkg/asciidoc/ast"

func (p *Parser) parseBlockMacro(metadata ast.BlockMetadata) *ast.BlockMacro {
	token := p.current
	macro := &ast.BlockMacro{
		Source:           sourceWithMetadata(tokenSpan(token), metadata),
		NameSource:       tokenFragmentSpan(token, token.BlockMacro.NameByteOffset, len(token.BlockMacro.Name)),
		TargetSource:     tokenFragmentSpan(token, token.BlockMacro.TargetByteOffset, len(token.BlockMacro.Target)),
		AttributesSource: tokenFragmentSpan(token, token.BlockMacro.AttributesByteOffset, len(token.BlockMacro.Attributes)),
		Name:             token.BlockMacro.Name,
		Target:           token.BlockMacro.Target,
		Attributes:       token.BlockMacro.Attributes,
		Metadata:         metadata,
	}
	p.advance()
	return macro
}
