package lexer

func matchDelimiter(raw string) (LineToken, bool) {
	var kind DelimiterKind
	if raw == "--" {
		kind = DelimiterOpen
	} else if len(raw) >= 4 && repeatedByte(raw) {
		switch raw[0] {
		case '-':
			kind = DelimiterListing
		case '.':
			kind = DelimiterLiteral
		case '=':
			kind = DelimiterExample
		case '*':
			kind = DelimiterSidebar
		case '_':
			kind = DelimiterQuote
		case '+':
			kind = DelimiterPassthrough
		case '/':
			kind = DelimiterComment
		}
	} else if len(raw) >= 4 && isTableDelimiterPrefix(raw[0]) && allBytes(raw[1:], '=') {
		kind = DelimiterTable
	}
	if kind == DelimiterUnknown {
		return LineToken{}, false
	}

	return delimiterLineToken(raw, DelimiterPayload{
		Kind:             kind,
		Marker:           raw,
		MarkerByteOffset: 0,
	}), true
}

func repeatedByte(raw string) bool {
	return allBytes(raw[1:], raw[0])
}

func allBytes(raw string, want byte) bool {
	for index := range len(raw) {
		if raw[index] != want {
			return false
		}
	}
	return true
}

func isTableDelimiterPrefix(value byte) bool {
	return value == '|' || value == ',' || value == ':' || value == '!'
}
