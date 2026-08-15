package ast

// BlockMacro preserves a block-macro-shaped line. Preprocessor directives
// such as include are represented but are not evaluated by the parser.
type BlockMacro struct {
	Source           Span
	NameSource       Span
	TargetSource     Span
	AttributesSource Span
	Name             string
	Target           string
	Attributes       string
	Metadata         BlockMetadata
}

func (m *BlockMacro) SourceSpan() Span { return m.Source }
func (*BlockMacro) blockNode()         {}
