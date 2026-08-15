package lexer

// matchAttributeList recognizes any complete block attribute line. It splits
// entries on commas outside double quotes, but deliberately does not assign
// semantic meaning to positional entries.
func matchAttributeList(raw string) (LineToken, bool) {
	if len(raw) < 2 || raw[0] != '[' || raw[len(raw)-1] != ']' {
		return LineToken{}, false
	}

	return attributeListLineToken(raw, AttributeListPayload{
		Entries: splitAttributeListEntries(raw, 1, len(raw)-1),
	}), true
}

func splitAttributeListEntries(raw string, contentStart, contentEnd int) []AttributeListEntry {
	if contentStart == contentEnd {
		return nil
	}

	var entries []AttributeListEntry
	start := contentStart
	quoted := false
	escaped := false
	for index := contentStart; index < contentEnd; index++ {
		switch {
		case escaped:
			escaped = false
		case raw[index] == '\\':
			escaped = true
		case raw[index] == '"':
			quoted = !quoted
		case raw[index] == ',' && !quoted:
			entries = append(entries, attributeListEntry(raw, start, index))
			start = index + 1
		}
	}
	return append(entries, attributeListEntry(raw, start, contentEnd))
}

func attributeListEntry(raw string, start, end int) AttributeListEntry {
	start, end = trimHorizontalBounds(raw, start, end)
	return AttributeListEntry{Value: raw[start:end], ValueByteOffset: start}
}
