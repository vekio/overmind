package ast

type AdmonitionKind uint8

const (
	AdmonitionUnknown AdmonitionKind = iota
	AdmonitionNote
	AdmonitionTip
	AdmonitionImportant
	AdmonitionCaution
	AdmonitionWarning
)

func (k AdmonitionKind) String() string {
	switch k {
	case AdmonitionNote:
		return "note"
	case AdmonitionTip:
		return "tip"
	case AdmonitionImportant:
		return "important"
	case AdmonitionCaution:
		return "caution"
	case AdmonitionWarning:
		return "warning"
	default:
		return "unknown"
	}
}

type Admonition struct {
	Source        Span
	LabelSource   Span
	ContentSource Span
	Kind          AdmonitionKind
	Label         string
	Text          string
	Inlines       []Inline
	Metadata      BlockMetadata
}

func (a *Admonition) SourceSpan() Span { return a.Source }
func (*Admonition) blockNode()         {}
