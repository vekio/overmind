package parser

import (
	"fmt"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/diagnostic"
)

func (p *Parser) parseSectionHeading(openSections []*ast.Section, metadata ast.BlockMetadata) *ast.Section {
	token := p.current
	level := token.Heading.Level
	previousLevel := 0
	if len(openSections) > 0 {
		previousLevel = openSections[len(openSections)-1].Level
	}
	if level > previousLevel+1 {
		p.appendDiagnostic(
			diagnostic.SeverityWarning,
			fmt.Sprintf("section level jumped from %d to %d", previousLevel, level),
			tokenFragmentSpan(token, 0, level+1),
		)
	}

	section := &ast.Section{
		Source:        sourceWithMetadata(tokenSpan(token), metadata),
		HeadingSource: tokenSpan(token),
		ContentSource: ast.Span{Start: tokenEndPosition(token), End: tokenEndPosition(token)},
		TitleSource:   tokenFragmentSpan(token, token.Heading.TitleByteOffset, len(token.Heading.Title)),
		Level:         level,
		Title:         token.Heading.Title,
		Metadata:      metadata,
	}
	p.advance()
	return section
}

func closeSectionsAtLevel(sections []*ast.Section, level int, end ast.Position) []*ast.Section {
	for len(sections) > 0 && sections[len(sections)-1].Level >= level {
		last := sections[len(sections)-1]
		last.Source.End = end
		last.ContentSource.End = end
		sections = sections[:len(sections)-1]
	}
	return sections
}

func closeAllSections(sections []*ast.Section, end ast.Position) {
	for _, section := range sections {
		section.Source.End = end
		section.ContentSource.End = end
	}
}
