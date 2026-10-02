package parser

import (
	"fmt"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
	"github.com/vekio/overmind/pkg/asciidoc/lexer"
)

func (p *Parser) parseDocument() *ast.Document {
	document := &ast.Document{Source: ast.Span{Start: sourceOrigin()}}
	var sections []*ast.Section
	var metadata pendingMetadata
	seenBody := false
	headerStarted := false
	headerClosed := false

	for p.current.Kind != lexer.LineEOF {
		switch p.current.Kind {
		case lexer.LineBlank:
			if !metadata.empty() {
				metadata.discard(p, "block metadata is not attached to a block")
			}
			if headerStarted {
				headerClosed = true
			}
			p.advance()
		case lexer.LineComment:
			if !metadata.empty() {
				metadata.discard(p, "block metadata is not attached to a block")
			}
			p.advance()
		case lexer.LineBlockTitle, lexer.LineAnchor, lexer.LineAttributeList:
			seenBody = true
			metadata.add(p, p.current)
			p.advance()
		case lexer.LineDelimiter:
			seenBody = true
			block := p.parseDelimitedBlock(metadata.take())
			if block != nil {
				appendBlock(document, sections, block)
			}
		case lexer.LineListItem:
			seenBody = true
			appendBlock(document, sections, p.parseList(metadata.take()))
		case lexer.LineDescriptionListItem:
			seenBody = true
			appendBlock(document, sections, p.parseDescriptionList(metadata.take()))
		case lexer.LineAdmonition:
			seenBody = true
			appendBlock(document, sections, p.parseAdmonition(metadata.take()))
		case lexer.LineBlockMacro:
			seenBody = true
			appendBlock(document, sections, p.parseBlockMacro(metadata.take()))
		case lexer.LineAttributeEntry:
			if !metadata.empty() {
				metadata.discard(p, "block metadata is not attached to a supported block")
			}
			header := !seenBody && !headerClosed
			if header {
				headerStarted = true
			}
			appendBlock(document, sections, p.parseAttributeEntry(header))
		case lexer.LineHeading:
			if p.current.Heading.Level == 0 {
				if !metadata.empty() {
					metadata.discard(p, "block metadata is not attached to a document title")
				}
				if document.Title == nil && !seenBody && !headerClosed && len(sections) == 0 {
					headerStarted = true
					document.Title = p.parseDocumentTitle()
					continue
				}
				p.appendDiagnostic(
					diagnostic.SeverityError,
					"document title is only allowed before document body content",
					tokenSpan(p.current),
				)
				p.advance()
				continue
			}

			seenBody = true
			section := p.parseSectionHeading(sections, metadata.take())
			sections = closeSectionsAtLevel(sections, section.Level, section.Source.Start)
			appendBlock(document, sections, section)
			sections = append(sections, section)
		case lexer.LineText:
			if !metadata.empty() {
				metadata.discard(p, "block metadata is not attached to a supported block")
			}
			seenBody = true
			appendBlock(document, sections, p.parseParagraph())
		case lexer.LineThematicBreak:
			if !metadata.empty() {
				metadata.discard(p, "block metadata is not attached to a supported block")
			}
			seenBody = true
			appendBlock(document, sections, p.parseThematicBreak())
		case lexer.LinePageBreak:
			if !metadata.empty() {
				metadata.discard(p, "block metadata is not attached to a supported block")
			}
			seenBody = true
			appendBlock(document, sections, p.parsePageBreak())
		default:
			if !metadata.empty() {
				metadata.discard(p, "block metadata is not attached to a supported block")
			}
			seenBody = true
			p.appendDiagnostic(
				diagnostic.SeverityWarning,
				fmt.Sprintf("parser does not yet support %s lines", p.current.Kind),
				tokenSpan(p.current),
			)
			p.advance()
		}
	}

	if !metadata.empty() {
		metadata.discard(p, "block metadata is not attached to a block")
	}
	end := tokenSpan(p.current).Start
	closeAllSections(sections, end)
	document.Source.End = end
	return document
}

func appendBlock(document *ast.Document, sections []*ast.Section, block ast.Block) {
	if len(sections) == 0 {
		document.Blocks = append(document.Blocks, block)
		return
	}
	parent := sections[len(sections)-1]
	parent.Blocks = append(parent.Blocks, block)
}
