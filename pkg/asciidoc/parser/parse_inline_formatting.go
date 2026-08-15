package parser

import (
	"unicode"
	"unicode/utf8"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

func (p *inlineParser) formattingWidth(at, rangeStart, rangeEnd int) int {
	if !isFormattingMark(p.source[at]) {
		return 0
	}
	if at+1 < rangeEnd && p.source[at+1] == p.source[at] {
		if (at == rangeStart || p.source[at-1] != p.source[at]) && (at+2 == rangeEnd || p.source[at+2] != p.source[at]) {
			return 2
		}
		return 0
	}
	if p.isConstrainedOpening(at, rangeStart, rangeEnd) {
		return 1
	}
	return 0
}

func (p *inlineParser) findClosing(marker byte, width, start, end int) int {
	for cursor := start; cursor < end; {
		if p.source[cursor] == '\\' && cursor+1 < end {
			cursor += 1 + escapedFormattingWidth(p.source, cursor+1, end)
			continue
		}
		if p.source[cursor] == marker {
			if width == 2 && cursor+1 < end && p.source[cursor+1] == marker && (cursor+2 == end || p.source[cursor+2] != marker) {
				return cursor
			}
			if width == 1 && p.isConstrainedClosing(cursor, start, end) {
				return cursor
			}
		}
		_, size := utf8.DecodeRuneInString(p.source[cursor:end])
		cursor += size
	}
	return -1
}

func escapedFormattingWidth(source string, at, end int) int {
	if at >= end || !isFormattingMark(source[at]) {
		return 1
	}
	if at+1 < end && source[at+1] == source[at] {
		return 2
	}
	return 1
}

func (p *inlineParser) isConstrainedOpening(at, rangeStart, rangeEnd int) bool {
	if adjacentSameMarker(p.source, at, rangeStart, rangeEnd) || at+1 >= rangeEnd {
		return false
	}
	next, _ := utf8.DecodeRuneInString(p.source[at+1 : rangeEnd])
	if unicode.IsSpace(next) {
		return false
	}
	if at == rangeStart {
		return true
	}
	previous, _ := utf8.DecodeLastRuneInString(p.source[rangeStart:at])
	return constrainedOpeningBoundary(previous)
}

func (p *inlineParser) isConstrainedClosing(at, rangeStart, rangeEnd int) bool {
	if adjacentSameMarker(p.source, at, rangeStart, rangeEnd) || at == rangeStart {
		return false
	}
	previous, _ := utf8.DecodeLastRuneInString(p.source[rangeStart:at])
	if unicode.IsSpace(previous) {
		return false
	}
	if at+1 == rangeEnd {
		return true
	}
	next, _ := utf8.DecodeRuneInString(p.source[at+1 : rangeEnd])
	return constrainedClosingBoundary(next)
}

func constrainedOpeningBoundary(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != ':' && r != ';' && r != '}'
}

func constrainedClosingBoundary(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
}

func adjacentSameMarker(source string, at, start, end int) bool {
	marker := source[at]
	return at > start && source[at-1] == marker || at+1 < end && source[at+1] == marker
}

func isFormattingMark(value byte) bool {
	return value == '*' || value == '_' || value == '`'
}

func formattedInline(marker byte, source, contentSource ast.Span, children []ast.Inline) ast.Inline {
	switch marker {
	case '*':
		return &ast.Strong{Source: source, ContentSource: contentSource, Children: children}
	case '_':
		return &ast.Emphasis{Source: source, ContentSource: contentSource, Children: children}
	default:
		return &ast.Monospace{Source: source, ContentSource: contentSource, Children: children}
	}
}
