package ast

// Paragraph contains one or more contiguous plain-text source lines. Physical
// lines are normalized to LF in Text; Source still points to the original.
type Paragraph struct {
	Source  Span
	Text    string
	Inlines []Inline
}

// SourceSpan returns the paragraph's source span.
func (p *Paragraph) SourceSpan() Span { return p.Source }

func (*Paragraph) blockNode() {}
