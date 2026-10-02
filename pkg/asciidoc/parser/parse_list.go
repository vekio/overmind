package parser

import (
	"fmt"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
	"github.com/vekio/overmind/pkg/asciidoc/lexer"
)

func (p *Parser) parseList(metadata ast.BlockMetadata) *ast.List {
	kind := listKind(p.current.List.Kind)
	level := p.current.List.Level
	list := &ast.List{Kind: kind, Level: level, Metadata: metadata}
	list.Source.Start = tokenSpan(p.current).Start
	if !metadataEmpty(metadata) {
		list.Source.Start = metadata.Source.Start
	}

	for p.current.Kind == lexer.LineListItem && listKind(p.current.List.Kind) == kind && p.current.List.Level == level {
		item := p.parseListItem()
		p.parseListItemBlocks(item, kind, level)
		list.Items = append(list.Items, item)
		list.Source.End = item.Source.End

		for p.current.Kind == lexer.LineBlank || p.current.Kind == lexer.LineComment {
			p.advance()
		}
	}
	return list
}

func (p *Parser) parseListItem() *ast.ListItem {
	token := p.current
	markerSource := tokenFragmentSpan(token, token.List.MarkerByteOffset, len(token.List.Marker))
	principalSource := tokenFragmentSpan(token, token.List.PrincipalByteOffset, len(token.List.Principal))
	p.advance()
	principal, inlines, completePrincipalSource := p.parseTextContinuation(token.List.Principal, principalSource, token.Ending)
	return &ast.ListItem{
		Source:          ast.Span{Start: tokenSpan(token).Start, End: completePrincipalSource.End},
		MarkerSource:    markerSource,
		PrincipalSource: completePrincipalSource,
		Marker:          token.List.Marker,
		Principal:       principal,
		Inlines:         inlines,
	}
}

func (p *Parser) parseListItemBlocks(item *ast.ListItem, kind ast.ListKind, level int) {
	for {
		if p.current.Kind == lexer.LineListItem && listKind(p.current.List.Kind) == kind && p.current.List.Level > level {
			if p.current.List.Level > level+1 {
				p.appendDiagnostic(diagnostic.SeverityWarning, fmt.Sprintf("list level jumped from %d to %d", level, p.current.List.Level), tokenSpan(p.current))
			}
			block := p.parseList(ast.BlockMetadata{})
			item.Blocks = append(item.Blocks, block)
			item.Source.End = block.Source.End
			continue
		}
		if p.current.Kind != lexer.LineContinuation {
			return
		}
		continuationSource := tokenSpan(p.current)
		p.advance()
		block := p.parseAttachedBlock()
		if block == nil {
			p.appendDiagnostic(diagnostic.SeverityError, "list continuation is not followed by a supported block", continuationSource)
			return
		}
		item.Blocks = append(item.Blocks, block)
		item.Source.End = block.SourceSpan().End
	}
}

func listKind(kind lexer.ListKind) ast.ListKind {
	if kind == lexer.ListUnordered {
		return ast.ListUnordered
	}
	if kind == lexer.ListOrdered {
		return ast.ListOrdered
	}
	return ast.ListUnknown
}
