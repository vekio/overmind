package parser

import (
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/diagnostic"
	"git.casta.me/alberto/overmind/pkg/asciidoc/lexer"
)

type pendingMetadata struct {
	metadata ast.BlockMetadata
	lines    []ast.Span
}

func (m *pendingMetadata) empty() bool {
	return len(m.lines) == 0
}

func (m *pendingMetadata) add(p *Parser, token lexer.LineToken) {
	source := tokenSpan(token)
	if m.empty() {
		m.metadata.Source.Start = source.Start
	}
	m.metadata.Source.End = source.End
	m.lines = append(m.lines, source)

	switch token.Kind {
	case lexer.LineBlockTitle:
		if m.metadata.Title != nil {
			p.appendDiagnostic(diagnostic.SeverityWarning, "duplicate block title; the closest title is used", source)
		}
		m.metadata.Title = &ast.BlockTitle{
			Source:      source,
			TitleSource: tokenFragmentSpan(token, token.BlockTitle.TitleByteOffset, len(token.BlockTitle.Title)),
			Text:        token.BlockTitle.Title,
		}
	case lexer.LineAnchor:
		if m.metadata.Anchor != nil {
			p.appendDiagnostic(diagnostic.SeverityWarning, "duplicate block anchor; the closest anchor is used", source)
		}
		anchor := &ast.Anchor{
			Source:   source,
			IDSource: tokenFragmentSpan(token, token.Anchor.IDByteOffset, len(token.Anchor.ID)),
			ID:       token.Anchor.ID,
			RefText:  token.Anchor.RefText,
		}
		if token.Anchor.RefText != "" {
			anchor.RefTextSource = tokenFragmentSpan(token, token.Anchor.RefTextByteOffset, len(token.Anchor.RefText))
		}
		m.metadata.Anchor = anchor
	case lexer.LineAttributeList:
		attributeList := ast.AttributeList{Source: source}
		attributeList.Entries = make([]ast.Attribute, 0, len(token.AttributeList.Entries))
		for _, entry := range token.AttributeList.Entries {
			attributeList.Entries = append(attributeList.Entries, ast.Attribute{
				Source: tokenFragmentSpan(token, entry.ValueByteOffset, len(entry.Value)),
				Value:  entry.Value,
			})
		}
		m.metadata.AttributeLists = append(m.metadata.AttributeLists, attributeList)
	}
}

func (m *pendingMetadata) take() ast.BlockMetadata {
	metadata := m.metadata
	*m = pendingMetadata{}
	return metadata
}

func (m *pendingMetadata) discard(p *Parser, reason string) {
	for _, source := range m.lines {
		p.appendDiagnostic(diagnostic.SeverityWarning, reason, source)
	}
	*m = pendingMetadata{}
}
