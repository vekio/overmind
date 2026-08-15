package lexer

import "fmt"

// LineToken is one classified physical source line. Raw and Source exclude the
// line ending. Kind is the discriminator; only the payload associated with
// that kind is meaningful.
type LineToken struct {
	Kind   LineTokenKind
	Raw    string
	Ending LineEnding
	Source Span

	Heading         HeadingPayload
	AttributeEntry  AttributeEntryPayload
	AttributeList   AttributeListPayload
	Anchor          AnchorPayload
	BlockTitle      BlockTitlePayload
	List            ListPayload
	DescriptionList DescriptionListPayload
	Admonition      AdmonitionPayload
	Comment         CommentPayload
	Delimiter       DelimiterPayload
	BlockMacro      BlockMacroPayload
}

// String returns the token kind, raw source, and source span.
func (t LineToken) String() string {
	return fmt.Sprintf("%s(%q) %s", t.Kind, t.Raw, t.Source)
}

// SourceText returns the exact source bytes consumed for this token, including
// its line ending. Concatenating SourceText for all non-EOF tokens reconstructs
// the original input.
func (t LineToken) SourceText() string {
	return t.Raw + t.Ending.Text()
}

func blankLineToken(raw string) LineToken {
	return LineToken{Kind: LineBlank, Raw: raw}
}

func textLineToken(raw string) LineToken {
	return LineToken{Kind: LineText, Raw: raw}
}

func eofLineToken(at Position) LineToken {
	return LineToken{Kind: LineEOF, Source: Span{Start: at, End: at}}
}

func thematicBreakLineToken(raw string) LineToken {
	return LineToken{Kind: LineThematicBreak, Raw: raw}
}

func continuationLineToken(raw string) LineToken {
	return LineToken{Kind: LineContinuation, Raw: raw}
}

func pageBreakLineToken(raw string) LineToken {
	return LineToken{Kind: LinePageBreak, Raw: raw}
}

func matchLine(raw string) LineToken {
	if raw == "" {
		return blankLineToken(raw)
	}

	switch raw[0] {
	case ' ', '\t':
		if token, ok := matchBlankLine(raw); ok {
			return token
		}
	case '=':
		if token, ok := matchDelimiter(raw); ok {
			return token
		}
		if token, ok := matchHeading(raw); ok {
			return token
		}
	case '[':
		if token, ok := matchAnchor(raw); ok {
			return token
		}
		if token, ok := matchAttributeList(raw); ok {
			return token
		}
	case '-':
		if token, ok := matchDelimiter(raw); ok {
			return token
		}
	case '*', '.':
		if token, ok := matchDelimiter(raw); ok {
			return token
		}
		if token, ok := matchListItem(raw); ok {
			return token
		}
		if token, ok := matchBlockTitle(raw); ok {
			return token
		}
	case '_', '|', ',', '!':
		if token, ok := matchDelimiter(raw); ok {
			return token
		}
	case '+':
		if token, ok := matchDelimiter(raw); ok {
			return token
		}
		if token, ok := matchContinuation(raw); ok {
			return token
		}
	case '/':
		if token, ok := matchDelimiter(raw); ok {
			return token
		}
		if token, ok := matchLineComment(raw); ok {
			return token
		}
	case ':':
		if token, ok := matchDelimiter(raw); ok {
			return token
		}
		if token, ok := matchAttributeEntry(raw); ok {
			return token
		}
	case '<':
		if token, ok := matchPageBreak(raw); ok {
			return token
		}
	case '\'':
		if token, ok := matchThematicBreak(raw); ok {
			return token
		}
	case 'N', 'T', 'I', 'C', 'W':
		if token, ok := matchAdmonitionParagraph(raw); ok {
			return token
		}
	}
	if token, ok := matchBlockMacro(raw); ok {
		return token
	}
	if token, ok := matchDescriptionListItem(raw); ok {
		return token
	}
	return textLineToken(raw)
}
