package ast

type AttributeOperation uint8

const (
	AttributeOperationUnknown AttributeOperation = iota
	AttributeSet
	AttributeUnset
)

func (o AttributeOperation) String() string {
	switch o {
	case AttributeSet:
		return "set"
	case AttributeUnset:
		return "unset"
	default:
		return "unknown"
	}
}

// AttributeEntry sets or unsets a document attribute. Attribute references
// are intentionally not expanded during syntax parsing.
type AttributeEntry struct {
	Source      Span
	NameSource  Span
	ValueSource Span
	Operation   AttributeOperation
	Name        string
	Value       string
	Header      bool
}

func (a *AttributeEntry) SourceSpan() Span { return a.Source }
func (*AttributeEntry) blockNode()         {}
