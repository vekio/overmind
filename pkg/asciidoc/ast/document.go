package ast

// Document is the root of a parsed AsciiDoc syntax tree.
type Document struct {
	Source Span
	Title  *DocumentTitle
	Blocks []Block
}

// SourceSpan returns the source occupied by the complete document.
func (d *Document) SourceSpan() Span { return d.Source }

// DocumentTitle is a level-zero heading used as the document title.
type DocumentTitle struct {
	Source      Span
	TitleSource Span
	Text        string
}

// SourceSpan returns the complete document-title line span.
func (t *DocumentTitle) SourceSpan() Span { return t.Source }
