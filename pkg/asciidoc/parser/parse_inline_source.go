package parser

import (
	"strings"
	"unicode/utf8"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

func (p *inlineParser) span(start, end int) ast.Span {
	return ast.Span{Start: p.positions[start], End: p.positions[end]}
}

func normalizedRuneSize(source string, at, end int, target *strings.Builder) int {
	if source[at] == '\r' {
		target.WriteByte('\n')
		if at+1 < end && source[at+1] == '\n' {
			return 2
		}
		return 1
	}
	_, size := utf8.DecodeRuneInString(source[at:end])
	target.WriteString(source[at : at+size])
	return size
}

func inlinePositions(source string, origin ast.Position) []ast.Position {
	positions := make([]ast.Position, len(source)+1)
	position := origin
	for cursor := 0; cursor < len(source); {
		positions[cursor] = position
		if source[cursor] == '\r' {
			position.Offset++
			cursor++
			positions[cursor] = position
			if cursor < len(source) && source[cursor] == '\n' {
				position.Offset++
				cursor++
				position.Line++
				position.Column = 1
				continue
			}
			position.Line++
			position.Column = 1
			continue
		}
		if source[cursor] == '\n' {
			position.Offset++
			position.Line++
			position.Column = 1
			cursor++
			continue
		}
		_, size := utf8.DecodeRuneInString(source[cursor:])
		for offset := 1; offset <= size; offset++ {
			positions[cursor+offset] = ast.Position{
				Offset: position.Offset + offset,
				Line:   position.Line,
				Column: position.Column + 1,
			}
		}
		position = positions[cursor+size]
		cursor += size
	}
	positions[len(source)] = position
	return positions
}
