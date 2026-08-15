package parser

import (
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/lexer"
)

func (p *Parser) parseAttachedBlock() ast.Block {
	var metadata pendingMetadata
	for p.current.Kind == lexer.LineBlockTitle || p.current.Kind == lexer.LineAnchor || p.current.Kind == lexer.LineAttributeList {
		metadata.add(p, p.current)
		p.advance()
	}
	switch p.current.Kind {
	case lexer.LineText:
		if !metadata.empty() {
			metadata.discard(p, "block metadata is not attached to a supported block")
		}
		return p.parseParagraph()
	case lexer.LineDelimiter:
		return p.parseDelimitedBlock(metadata.take())
	case lexer.LineListItem:
		return p.parseList(metadata.take())
	case lexer.LineDescriptionListItem:
		return p.parseDescriptionList(metadata.take())
	case lexer.LineAdmonition:
		return p.parseAdmonition(metadata.take())
	case lexer.LineBlockMacro:
		return p.parseBlockMacro(metadata.take())
	case lexer.LineThematicBreak:
		return p.parseThematicBreak()
	case lexer.LinePageBreak:
		return p.parsePageBreak()
	default:
		if !metadata.empty() {
			metadata.discard(p, "block metadata is not attached to a supported block")
		}
		return nil
	}
}
