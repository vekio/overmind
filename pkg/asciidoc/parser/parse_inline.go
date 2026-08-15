package parser

import (
	"strings"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

// inlineParser parses constrained formatting directly from original source
// bytes. positions maps every byte boundary back to the input document.
type inlineParser struct {
	source    string
	positions []ast.Position
}

func parseInlines(source string, origin ast.Position) []ast.Inline {
	parser := inlineParser{
		source:    source,
		positions: inlinePositions(source, origin),
	}
	return parser.parseRange(0, len(source))
}

func (p *inlineParser) parseRange(start, end int) []ast.Inline {
	result := make([]ast.Inline, 0)
	textStart := start
	var text strings.Builder

	flushText := func(sourceEnd int) {
		if text.Len() == 0 {
			textStart = sourceEnd
			return
		}
		result = append(result, &ast.Text{
			Source: p.span(textStart, sourceEnd),
			Value:  text.String(),
		})
		text.Reset()
		textStart = sourceEnd
	}

	for cursor := start; cursor < end; {
		if p.source[cursor] == '\\' && cursor+1 < end && isFormattingMark(p.source[cursor+1]) {
			width := 1
			if cursor+2 < end && p.source[cursor+2] == p.source[cursor+1] {
				width = 2
			}
			text.WriteString(p.source[cursor+1 : cursor+1+width])
			cursor += 1 + width
			continue
		}
		if p.source[cursor] == '\\' && cursor+1 < end {
			if macro, ok := p.matchInlineMacro(cursor+1, end); ok {
				text.WriteString(p.source[cursor+1 : macro.end])
				cursor = macro.end
				continue
			}
		}
		if macro, ok := p.matchInlineMacro(cursor, end); ok {
			flushText(cursor)
			children := make([]ast.Inline, 0)
			if macro.labelStart < macro.labelEnd {
				children = p.parseRange(macro.labelStart, macro.labelEnd)
			}
			result = append(result, macro.inline(p, children))
			cursor = macro.end
			textStart = cursor
			continue
		}

		marker := p.source[cursor]
		if width := p.formattingWidth(cursor, start, end); width > 0 {
			if closing := p.findClosing(marker, width, cursor+width, end); closing >= 0 {
				flushText(cursor)
				children := p.parseRange(cursor+width, closing)
				result = append(result, formattedInline(marker, p.span(cursor, closing+width), p.span(cursor+width, closing), children))
				cursor = closing + width
				textStart = cursor
				continue
			}
		}

		size := normalizedRuneSize(p.source, cursor, end, &text)
		cursor += size
	}
	flushText(end)
	return result
}
