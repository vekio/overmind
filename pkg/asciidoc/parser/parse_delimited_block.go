package parser

import (
	"fmt"
	"strings"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
	"github.com/vekio/overmind/pkg/asciidoc/lexer"
)

func (p *Parser) parseDelimitedBlock(metadata ast.BlockMetadata) ast.Block {
	if p.current.Delimiter.Kind == lexer.DelimiterTable {
		return p.parseTable(metadata)
	}
	opening := p.current
	openingSource := tokenSpan(opening)
	contentStart := tokenEndPosition(opening)
	block := &ast.DelimitedBlock{
		Source:        ast.Span{Start: openingSource.Start},
		OpeningSource: openingSource,
		ContentSource: ast.Span{Start: contentStart, End: contentStart},
		Kind:          delimitedBlockKind(opening.Delimiter.Kind),
		ContentModel:  delimitedBlockContentModel(opening.Delimiter.Kind),
		Marker:        opening.Delimiter.Marker,
		Metadata:      metadata,
	}
	if !metadataEmpty(metadata) {
		block.Source.Start = metadata.Source.Start
	}

	p.advance()
	var content strings.Builder
	for p.current.Kind != lexer.LineEOF {
		if p.current.Kind == lexer.LineDelimiter && p.current.Delimiter.Marker == block.Marker {
			block.ClosingSource = tokenSpan(p.current)
			block.ContentSource.End = block.ClosingSource.Start
			block.Source.End = block.ClosingSource.End
			block.Content = content.String()
			block.Closed = true
			p.advance()
			if block.Kind == ast.DelimitedBlockComment {
				return nil
			}
			return block
		}
		content.WriteString(p.current.SourceText())
		p.advance()
	}

	eof := tokenSpan(p.current).Start
	block.ContentSource.End = eof
	block.Source.End = eof
	block.Content = content.String()
	p.appendDiagnostic(
		diagnostic.SeverityError,
		fmt.Sprintf("unclosed %s block; expected %q", block.Kind, block.Marker),
		openingSource,
	)
	if block.Kind == ast.DelimitedBlockComment {
		return nil
	}
	return block
}

func metadataEmpty(metadata ast.BlockMetadata) bool {
	return metadata.Title == nil && metadata.Anchor == nil && len(metadata.AttributeLists) == 0
}

func delimitedBlockKind(kind lexer.DelimiterKind) ast.DelimitedBlockKind {
	switch kind {
	case lexer.DelimiterOpen:
		return ast.DelimitedBlockOpen
	case lexer.DelimiterListing:
		return ast.DelimitedBlockListing
	case lexer.DelimiterLiteral:
		return ast.DelimitedBlockLiteral
	case lexer.DelimiterExample:
		return ast.DelimitedBlockExample
	case lexer.DelimiterSidebar:
		return ast.DelimitedBlockSidebar
	case lexer.DelimiterQuote:
		return ast.DelimitedBlockQuote
	case lexer.DelimiterPassthrough:
		return ast.DelimitedBlockPassthrough
	case lexer.DelimiterComment:
		return ast.DelimitedBlockComment
	default:
		return ast.DelimitedBlockUnknown
	}
}

func delimitedBlockContentModel(kind lexer.DelimiterKind) ast.ContentModel {
	switch kind {
	case lexer.DelimiterOpen, lexer.DelimiterExample, lexer.DelimiterSidebar, lexer.DelimiterQuote:
		return ast.ContentModelCompound
	case lexer.DelimiterListing, lexer.DelimiterLiteral:
		return ast.ContentModelVerbatim
	case lexer.DelimiterPassthrough:
		return ast.ContentModelRaw
	default:
		return ast.ContentModelUnknown
	}
}
