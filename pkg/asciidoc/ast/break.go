package ast

// ThematicBreak represents a line containing three apostrophes.
type ThematicBreak struct {
	Source Span
}

// SourceSpan returns the thematic break's source span.
func (b *ThematicBreak) SourceSpan() Span { return b.Source }

func (*ThematicBreak) blockNode() {}

// PageBreak represents a line containing three less-than signs.
type PageBreak struct {
	Source Span
}

// SourceSpan returns the page break's source span.
func (b *PageBreak) SourceSpan() Span { return b.Source }

func (*PageBreak) blockNode() {}
