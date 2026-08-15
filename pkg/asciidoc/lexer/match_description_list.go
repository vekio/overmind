package lexer

func matchDescriptionListItem(raw string) (LineToken, bool) {
	for markerStart := 1; markerStart < len(raw); markerStart++ {
		if raw[markerStart] != ':' {
			continue
		}

		markerEnd := markerStart
		for markerEnd < len(raw) && raw[markerEnd] == ':' {
			markerEnd++
		}
		if markerEnd-markerStart < 2 ||
			(markerEnd < len(raw) && !isHorizontalSpace(raw[markerEnd])) {
			continue
		}

		termStart, termEnd := trimHorizontalBounds(raw, 0, markerStart)
		if termStart == termEnd {
			return LineToken{}, false
		}
		descriptionStart := markerEnd
		for descriptionStart < len(raw) && isHorizontalSpace(raw[descriptionStart]) {
			descriptionStart++
		}

		return descriptionListItemLineToken(raw, DescriptionListPayload{
			Term:                  raw[termStart:termEnd],
			TermByteOffset:        termStart,
			Marker:                raw[markerStart:markerEnd],
			MarkerByteOffset:      markerStart,
			Description:           raw[descriptionStart:],
			DescriptionByteOffset: descriptionStart,
		}), true
	}
	return LineToken{}, false
}
