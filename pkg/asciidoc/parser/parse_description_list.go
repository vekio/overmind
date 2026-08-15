package parser

import (
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/diagnostic"
	"git.casta.me/alberto/overmind/pkg/asciidoc/lexer"
)

func (p *Parser) parseDescriptionList(metadata ast.BlockMetadata) *ast.DescriptionList {
	level := len(p.current.DescriptionList.Marker) - 1
	list := &ast.DescriptionList{Level: level, Metadata: metadata}
	list.Source.Start = tokenSpan(p.current).Start
	if !metadataEmpty(metadata) {
		list.Source.Start = metadata.Source.Start
	}
	for p.current.Kind == lexer.LineDescriptionListItem && len(p.current.DescriptionList.Marker)-1 == level {
		item := p.parseDescriptionListItem()
		p.parseDescriptionListItemBlocks(item, level)
		list.Items = append(list.Items, item)
		list.Source.End = item.Source.End
		for p.current.Kind == lexer.LineBlank || p.current.Kind == lexer.LineComment {
			p.advance()
		}
	}
	return list
}

func (p *Parser) parseDescriptionListItemBlocks(item *ast.DescriptionListItem, level int) {
	for {
		if p.current.Kind == lexer.LineDescriptionListItem && len(p.current.DescriptionList.Marker)-1 > level {
			block := p.parseDescriptionList(ast.BlockMetadata{})
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

func (p *Parser) parseDescriptionListItem() *ast.DescriptionListItem {
	token := p.current
	termSource := tokenFragmentSpan(token, token.DescriptionList.TermByteOffset, len(token.DescriptionList.Term))
	markerSource := tokenFragmentSpan(token, token.DescriptionList.MarkerByteOffset, len(token.DescriptionList.Marker))
	descriptionSource := tokenFragmentSpan(token, token.DescriptionList.DescriptionByteOffset, len(token.DescriptionList.Description))
	p.advance()
	description, descriptionInlines, completeDescriptionSource := p.parseTextContinuation(token.DescriptionList.Description, descriptionSource, token.Ending)
	return &ast.DescriptionListItem{
		Source:             ast.Span{Start: tokenSpan(token).Start, End: completeDescriptionSource.End},
		TermSource:         termSource,
		MarkerSource:       markerSource,
		DescriptionSource:  completeDescriptionSource,
		Marker:             token.DescriptionList.Marker,
		Term:               token.DescriptionList.Term,
		TermInlines:        parseInlines(token.DescriptionList.Term, termSource.Start),
		Description:        description,
		DescriptionInlines: descriptionInlines,
	}
}
