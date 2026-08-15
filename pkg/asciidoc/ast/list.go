package ast

// ListKind identifies an ordered or unordered list.
type ListKind uint8

const (
	ListUnknown ListKind = iota
	ListUnordered
	ListOrdered
)

func (k ListKind) String() string {
	switch k {
	case ListUnordered:
		return "unordered"
	case ListOrdered:
		return "ordered"
	default:
		return "unknown"
	}
}

// List contains adjacent items with the same marker kind and level.
type List struct {
	Source   Span
	Kind     ListKind
	Level    int
	Metadata BlockMetadata
	Items    []*ListItem
}

func (l *List) SourceSpan() Span { return l.Source }
func (*List) blockNode()         {}

// ListItem contains its principal text and blocks attached by nesting or a
// continuation marker.
type ListItem struct {
	Source          Span
	MarkerSource    Span
	PrincipalSource Span
	Marker          string
	Principal       string
	Inlines         []Inline
	Blocks          []Block
}

func (i *ListItem) SourceSpan() Span { return i.Source }

// DescriptionList contains term-description pairs.
type DescriptionList struct {
	Source   Span
	Level    int
	Metadata BlockMetadata
	Items    []*DescriptionListItem
}

func (l *DescriptionList) SourceSpan() Span { return l.Source }
func (*DescriptionList) blockNode()         {}

type DescriptionListItem struct {
	Source             Span
	TermSource         Span
	MarkerSource       Span
	DescriptionSource  Span
	Marker             string
	Term               string
	TermInlines        []Inline
	Description        string
	DescriptionInlines []Inline
	Blocks             []Block
}

func (i *DescriptionListItem) SourceSpan() Span { return i.Source }
