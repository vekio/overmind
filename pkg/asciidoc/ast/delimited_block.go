package ast

// DelimitedBlockKind identifies the structural container used by a delimited
// block. Its semantic style may later be refined by block attributes.
type DelimitedBlockKind uint8

const (
	DelimitedBlockUnknown DelimitedBlockKind = iota
	DelimitedBlockOpen
	DelimitedBlockListing
	DelimitedBlockLiteral
	DelimitedBlockExample
	DelimitedBlockSidebar
	DelimitedBlockQuote
	DelimitedBlockPassthrough
	DelimitedBlockComment
)

// String returns a stable lowercase name for the block kind.
func (k DelimitedBlockKind) String() string {
	switch k {
	case DelimitedBlockOpen:
		return "open"
	case DelimitedBlockListing:
		return "listing"
	case DelimitedBlockLiteral:
		return "literal"
	case DelimitedBlockExample:
		return "example"
	case DelimitedBlockSidebar:
		return "sidebar"
	case DelimitedBlockQuote:
		return "quote"
	case DelimitedBlockPassthrough:
		return "passthrough"
	case DelimitedBlockComment:
		return "comment"
	default:
		return "unknown"
	}
}

// ContentModel describes how a parser should interpret block content.
type ContentModel uint8

const (
	ContentModelUnknown ContentModel = iota
	ContentModelCompound
	ContentModelVerbatim
	ContentModelRaw
)

// String returns a stable lowercase name for the content model.
func (m ContentModel) String() string {
	switch m {
	case ContentModelCompound:
		return "compound"
	case ContentModelVerbatim:
		return "verbatim"
	case ContentModelRaw:
		return "raw"
	default:
		return "unknown"
	}
}

// DelimitedBlock contains the exact source text between a pair of matching
// non-table fences. Compound content is intentionally not parsed yet.
type DelimitedBlock struct {
	Source        Span
	OpeningSource Span
	ContentSource Span
	ClosingSource Span
	Kind          DelimitedBlockKind
	ContentModel  ContentModel
	Marker        string
	Metadata      BlockMetadata
	Content       string
	Closed        bool
}

// SourceSpan returns the block span, including metadata and fences.
func (b *DelimitedBlock) SourceSpan() Span { return b.Source }

func (*DelimitedBlock) blockNode() {}

// BlockMetadata contains lines attached to the following block.
type BlockMetadata struct {
	Source         Span
	Title          *BlockTitle
	Anchor         *Anchor
	AttributeLists []AttributeList
}

// BlockTitle is a title line attached to a block.
type BlockTitle struct {
	Source      Span
	TitleSource Span
	Text        string
}

// Anchor is an ID and optional reference text attached to a block.
type Anchor struct {
	Source        Span
	IDSource      Span
	RefTextSource Span
	ID            string
	RefText       string
}

// AttributeList is one block attribute line and its lexical entries.
type AttributeList struct {
	Source  Span
	Entries []Attribute
}

// Attribute is one positional or named lexical entry. Semantic
// interpretation is deferred to a later parser increment.
type Attribute struct {
	Source Span
	Value  string
}
