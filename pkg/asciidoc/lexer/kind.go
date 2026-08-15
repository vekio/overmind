package lexer

// LineTokenKind identifies the lexical role of one physical source line.
type LineTokenKind uint8

const (
	LineInvalid LineTokenKind = iota
	LineEOF
	LineBlank
	LineHeading
	LineAttributeEntry
	LineAttributeList
	LineAnchor
	LineBlockTitle
	LineDelimiter
	LineListItem
	LineAdmonition
	LineComment
	LineThematicBreak
	LineContinuation
	LinePageBreak
	LineDescriptionListItem
	LineBlockMacro
	LineText
)

// String returns a stable name for the token kind.
func (k LineTokenKind) String() string {
	switch k {
	case LineInvalid:
		return "INVALID"
	case LineEOF:
		return "EOF"
	case LineBlank:
		return "BLANK"
	case LineHeading:
		return "HEADING"
	case LineAttributeEntry:
		return "ATTRIBUTE_ENTRY"
	case LineAttributeList:
		return "ATTRIBUTE_LIST"
	case LineAnchor:
		return "ANCHOR"
	case LineBlockTitle:
		return "BLOCK_TITLE"
	case LineDelimiter:
		return "DELIMITER"
	case LineListItem:
		return "LIST_ITEM"
	case LineAdmonition:
		return "ADMONITION"
	case LineComment:
		return "COMMENT"
	case LineThematicBreak:
		return "THEMATIC_BREAK"
	case LineContinuation:
		return "CONTINUATION"
	case LinePageBreak:
		return "PAGE_BREAK"
	case LineDescriptionListItem:
		return "DESCRIPTION_LIST_ITEM"
	case LineBlockMacro:
		return "BLOCK_MACRO"
	case LineText:
		return "TEXT"
	default:
		return "UNKNOWN"
	}
}

// AdmonitionKind identifies one of AsciiDoc's built-in admonition labels.
type AdmonitionKind uint8

const (
	AdmonitionUnknown AdmonitionKind = iota
	AdmonitionNote
	AdmonitionTip
	AdmonitionImportant
	AdmonitionCaution
	AdmonitionWarning
)

// String returns the AsciiDoc label for the admonition kind.
func (k AdmonitionKind) String() string {
	switch k {
	case AdmonitionUnknown:
		return "UNKNOWN"
	case AdmonitionNote:
		return "NOTE"
	case AdmonitionTip:
		return "TIP"
	case AdmonitionImportant:
		return "IMPORTANT"
	case AdmonitionCaution:
		return "CAUTION"
	case AdmonitionWarning:
		return "WARNING"
	default:
		return "UNKNOWN"
	}
}

// DelimiterKind identifies the construction represented by a delimiter.
type DelimiterKind uint8

const (
	DelimiterUnknown DelimiterKind = iota
	DelimiterOpen
	DelimiterListing
	DelimiterLiteral
	DelimiterExample
	DelimiterSidebar
	DelimiterQuote
	DelimiterPassthrough
	DelimiterComment
	DelimiterTable
)

// String returns a stable name for the delimiter kind.
func (k DelimiterKind) String() string {
	switch k {
	case DelimiterUnknown:
		return "UNKNOWN"
	case DelimiterOpen:
		return "OPEN"
	case DelimiterListing:
		return "LISTING"
	case DelimiterLiteral:
		return "LITERAL"
	case DelimiterExample:
		return "EXAMPLE"
	case DelimiterSidebar:
		return "SIDEBAR"
	case DelimiterQuote:
		return "QUOTE"
	case DelimiterPassthrough:
		return "PASSTHROUGH"
	case DelimiterComment:
		return "COMMENT"
	case DelimiterTable:
		return "TABLE"
	default:
		return "UNKNOWN"
	}
}

// AttributeOperation identifies whether a document attribute is set or unset.
type AttributeOperation uint8

const (
	AttributeOperationUnknown AttributeOperation = iota
	AttributeSet
	AttributeUnset
)

// String returns a stable name for the attribute operation.
func (o AttributeOperation) String() string {
	switch o {
	case AttributeOperationUnknown:
		return "UNKNOWN"
	case AttributeSet:
		return "SET"
	case AttributeUnset:
		return "UNSET"
	default:
		return "UNKNOWN"
	}
}

// ListKind identifies whether a list marker is ordered or unordered.
type ListKind uint8

const (
	ListUnknown ListKind = iota
	ListUnordered
	ListOrdered
)

// String returns a stable name for the list kind.
func (k ListKind) String() string {
	switch k {
	case ListUnknown:
		return "UNKNOWN"
	case ListUnordered:
		return "UNORDERED"
	case ListOrdered:
		return "ORDERED"
	default:
		return "UNKNOWN"
	}
}
