package parser

import (
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/lexer"
)

func (p *Parser) parseAttributeEntry(header bool) *ast.AttributeEntry {
	token := p.current
	entry := &ast.AttributeEntry{
		Source:     tokenSpan(token),
		NameSource: tokenFragmentSpan(token, token.AttributeEntry.NameByteOffset, len(token.AttributeEntry.Name)),
		Operation:  attributeOperation(token.AttributeEntry.Operation),
		Name:       token.AttributeEntry.Name,
		Value:      token.AttributeEntry.Value,
		Header:     header,
	}
	if token.AttributeEntry.Value != "" {
		entry.ValueSource = tokenFragmentSpan(token, token.AttributeEntry.ValueByteOffset, len(token.AttributeEntry.Value))
	}
	p.advance()
	return entry
}

func attributeOperation(operation lexer.AttributeOperation) ast.AttributeOperation {
	if operation == lexer.AttributeSet {
		return ast.AttributeSet
	}
	if operation == lexer.AttributeUnset {
		return ast.AttributeUnset
	}
	return ast.AttributeOperationUnknown
}
