// Package ast defines the public syntax tree produced by the AsciiDoc parser.
package ast

import "fmt"

// Position identifies a point in the original source. Offset is zero-based
// and measured in bytes. Line and Column are one-based, and Column is measured
// in Unicode code points.
type Position struct {
	Offset int
	Line   int
	Column int
}

// String returns the position as line:column@byte-offset.
func (p Position) String() string {
	return fmt.Sprintf("%d:%d@%d", p.Line, p.Column, p.Offset)
}

// Span is a half-open source range: Start is included and End is excluded.
type Span struct {
	Start Position
	End   Position
}

// String returns the span using half-open interval notation.
func (s Span) String() string {
	return fmt.Sprintf("[%s, %s)", s.Start, s.End)
}
