package lexer

// HeadingPayload contains the values extracted from a heading line.
type HeadingPayload struct {
	Level           int
	Title           string
	TitleByteOffset int
}

func headingLineToken(raw string, payload HeadingPayload) LineToken {
	return LineToken{Kind: LineHeading, Raw: raw, Heading: payload}
}

// AttributeEntryPayload contains a document attribute assignment or unset.
type AttributeEntryPayload struct {
	Operation       AttributeOperation
	Name            string
	NameByteOffset  int
	Value           string
	ValueByteOffset int
}

func attributeEntryLineToken(raw string, payload AttributeEntryPayload) LineToken {
	return LineToken{Kind: LineAttributeEntry, Raw: raw, AttributeEntry: payload}
}

// AttributeListEntry is one positional or named entry. Value remains lexical;
// assigning it semantic meaning is the parser's responsibility.
type AttributeListEntry struct {
	Value           string
	ValueByteOffset int
}

// AttributeListPayload contains a block attribute list's comma-separated
// entries.
type AttributeListPayload struct {
	Entries []AttributeListEntry
}

func attributeListLineToken(raw string, payload AttributeListPayload) LineToken {
	return LineToken{Kind: LineAttributeList, Raw: raw, AttributeList: payload}
}

// AnchorPayload contains the identifier and optional reference text from a
// block anchor line.
type AnchorPayload struct {
	ID                string
	IDByteOffset      int
	RefText           string
	RefTextByteOffset int
}

func anchorLineToken(raw string, payload AnchorPayload) LineToken {
	return LineToken{Kind: LineAnchor, Raw: raw, Anchor: payload}
}

// BlockTitlePayload contains the title following the leading dot.
type BlockTitlePayload struct {
	Title           string
	TitleByteOffset int
}

func blockTitleLineToken(raw string, payload BlockTitlePayload) LineToken {
	return LineToken{Kind: LineBlockTitle, Raw: raw, BlockTitle: payload}
}

// ListPayload contains the values extracted from a list item. Level is the
// number of repeated marker characters.
type ListPayload struct {
	Kind                ListKind
	Level               int
	Marker              string
	MarkerByteOffset    int
	Principal           string
	PrincipalByteOffset int
}

// DescriptionListPayload contains a term, its repeated colon delimiter, and
// the optional description on the same physical line.
type DescriptionListPayload struct {
	Term                  string
	TermByteOffset        int
	Marker                string
	MarkerByteOffset      int
	Description           string
	DescriptionByteOffset int
}

func descriptionListItemLineToken(raw string, payload DescriptionListPayload) LineToken {
	return LineToken{Kind: LineDescriptionListItem, Raw: raw, DescriptionList: payload}
}

func listItemLineToken(raw string, payload ListPayload) LineToken {
	return LineToken{Kind: LineListItem, Raw: raw, List: payload}
}

// AdmonitionPayload contains the values extracted from an admonition paragraph
// prefix. Content is the first line of the paragraph after the label.
type AdmonitionPayload struct {
	Kind              AdmonitionKind
	Label             string
	LabelByteOffset   int
	Content           string
	ContentByteOffset int
}

func admonitionLineToken(raw string, payload AdmonitionPayload) LineToken {
	return LineToken{Kind: LineAdmonition, Raw: raw, Admonition: payload}
}

// CommentPayload contains the source text following a line-comment prefix.
type CommentPayload struct {
	Content           string
	ContentByteOffset int
}

func commentLineToken(raw string, payload CommentPayload) LineToken {
	return LineToken{Kind: LineComment, Raw: raw, Comment: payload}
}

// DelimiterPayload describes a delimited block fence.
type DelimiterPayload struct {
	Kind             DelimiterKind
	Marker           string
	MarkerByteOffset int
}

func delimiterLineToken(raw string, payload DelimiterPayload) LineToken {
	return LineToken{Kind: LineDelimiter, Raw: raw, Delimiter: payload}
}

// BlockMacroPayload contains the lexical parts of a block macro. Attributes
// excludes the surrounding square brackets and remains uninterpreted.
type BlockMacroPayload struct {
	Name                 string
	NameByteOffset       int
	Target               string
	TargetByteOffset     int
	Attributes           string
	AttributesByteOffset int
}

func blockMacroLineToken(raw string, payload BlockMacroPayload) LineToken {
	return LineToken{Kind: LineBlockMacro, Raw: raw, BlockMacro: payload}
}
