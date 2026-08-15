package ast

// TableFormat identifies the data syntax used inside a table delimiter.
type TableFormat uint8

const (
	TableFormatUnknown TableFormat = iota
	TableFormatPSV
	TableFormatCSV
	TableFormatDSV
	TableFormatTSV
)

func (f TableFormat) String() string {
	switch f {
	case TableFormatPSV:
		return "psv"
	case TableFormatCSV:
		return "csv"
	case TableFormatDSV:
		return "dsv"
	case TableFormatTSV:
		return "tsv"
	default:
		return "unknown"
	}
}

// Table is a parsed delimited table. Content preserves the exact bytes between
// the fences, while Rows provides the supported structural interpretation.
type Table struct {
	// Source includes metadata and both fences when the table is closed.
	Source Span
	// OpeningSource covers the opening delimiter without its line ending.
	OpeningSource Span
	// ContentSource is the exact replaceable range between the fences.
	ContentSource Span
	// ClosingSource is zero for an unclosed table.
	ClosingSource Span
	Marker        string
	Format        TableFormat
	Separator     string
	Metadata      BlockMetadata
	Columns       []TableColumn
	Rows          []*TableRow
	// Content preserves the exact original bytes between the fences.
	Content string
	Header  bool
	Closed  bool
}

// SourceSpan returns the complete table span.
func (t *Table) SourceSpan() Span { return t.Source }
func (*Table) blockNode()         {}

// TableColumn is one lexical column specifier from the cols attribute.
type TableColumn struct {
	Spec string
}

// TableRow is one logical row grouped according to the effective column count.
type TableRow struct {
	Source Span
	Cells  []*TableCell
	Header bool
}

// SourceSpan returns the row span from its first through its last cell.
func (r *TableRow) SourceSpan() Span { return r.Source }

// TableCell contains normalized text, parsed inline nodes, and source ranges
// for both the cell marker and its content.
type TableCell struct {
	Source        Span
	MarkerSource  Span
	ContentSource Span
	Text          string
	Inlines       []Inline
}

// SourceSpan returns the complete cell span including its marker.
func (c *TableCell) SourceSpan() Span { return c.Source }
