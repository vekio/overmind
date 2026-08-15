package ast

// Inline is a node that may occur inside textual block content.
type Inline interface {
	Node
	inlineNode()
}

// Text is an unformatted inline fragment. Value has physical line endings
// normalized to LF and escaped formatting marks have their escape removed.
type Text struct {
	Source Span
	Value  string
}

// SourceSpan returns the source occupied by the text fragment.
func (t *Text) SourceSpan() Span { return t.Source }

func (*Text) inlineNode() {}

// Strong is text enclosed by a constrained pair of asterisks.
type Strong struct {
	Source        Span
	ContentSource Span
	Children      []Inline
}

// SourceSpan returns the source including the formatting marks.
func (s *Strong) SourceSpan() Span { return s.Source }

func (*Strong) inlineNode() {}

// Emphasis is text enclosed by a constrained pair of underscores.
type Emphasis struct {
	Source        Span
	ContentSource Span
	Children      []Inline
}

// SourceSpan returns the source including the formatting marks.
func (e *Emphasis) SourceSpan() Span { return e.Source }

func (*Emphasis) inlineNode() {}

// Monospace is text enclosed by a constrained pair of backticks.
type Monospace struct {
	Source        Span
	ContentSource Span
	Children      []Inline
}

// SourceSpan returns the source including the formatting marks.
func (m *Monospace) SourceSpan() Span { return m.Source }

func (*Monospace) inlineNode() {}

// Link is an explicit URL or link macro. Children contain the optional label.
type Link struct {
	Source       Span
	TargetSource Span
	LabelSource  Span
	Target       string
	Children     []Inline
}

func (l *Link) SourceSpan() Span { return l.Source }
func (*Link) inlineNode()        {}

// CrossReference refers to an anchor in this or another document.
type CrossReference struct {
	Source       Span
	TargetSource Span
	LabelSource  Span
	Target       string
	Children     []Inline
}

func (r *CrossReference) SourceSpan() Span { return r.Source }
func (*CrossReference) inlineNode()        {}
