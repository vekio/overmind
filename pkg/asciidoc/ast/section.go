package ast

// Section is a heading and the blocks nested beneath it. Level starts at one;
// level zero is reserved for DocumentTitle. All spans refer to the original
// source and become stale after that source is edited.
type Section struct {
	// Source includes attached metadata, the heading, body, and descendants.
	Source Span
	// HeadingSource covers only the heading line and excludes its line ending.
	HeadingSource Span
	// ContentSource begins after the heading line ending and extends to the
	// next section of the same or a higher level. It includes descendants.
	ContentSource Span
	// TitleSource covers only the title text, excluding the heading marker.
	TitleSource Span
	// Level is the number of heading markers minus the document-title level.
	Level int
	// Title is the heading's unformatted source text.
	Title string
	// Metadata contains attributes, title, and anchor attached to the heading.
	Metadata BlockMetadata
	// Blocks contains direct children; nested sections also occur here.
	Blocks []Block
}

// SourceSpan returns the section span, including all descendant blocks.
func (s *Section) SourceSpan() Span { return s.Source }

func (*Section) blockNode() {}
