package format

import "github.com/vekio/overmind/pkg/asciidoc/ast"

func makeJSONSpan(span ast.Span) jsonSpan {
	return jsonSpan{Start: makeJSONPosition(span.Start), End: makeJSONPosition(span.End)}
}

func makeJSONPosition(position ast.Position) jsonPosition {
	return jsonPosition{Offset: position.Offset, Line: position.Line, Column: position.Column}
}
