package lexer

const (
	minimumHeadingMarkerWidth = 1
	maximumHeadingMarkerWidth = 6
)

func matchHeading(raw string) (LineToken, bool) {
	markerWidth := 0
	for markerWidth < len(raw) && raw[markerWidth] == '=' {
		markerWidth++
	}

	if markerWidth < minimumHeadingMarkerWidth || markerWidth > maximumHeadingMarkerWidth {
		return LineToken{}, false
	}
	if markerWidth == len(raw) || !isHorizontalSpace(raw[markerWidth]) {
		return LineToken{}, false
	}

	titleStart := markerWidth
	for titleStart < len(raw) && isHorizontalSpace(raw[titleStart]) {
		titleStart++
	}
	titleEnd := len(raw)
	for titleEnd > titleStart && isHorizontalSpace(raw[titleEnd-1]) {
		titleEnd--
	}
	if titleStart == titleEnd {
		return LineToken{}, false
	}

	return headingLineToken(raw, HeadingPayload{
		Level:           markerWidth - 1,
		Title:           raw[titleStart:titleEnd],
		TitleByteOffset: titleStart,
	}), true
}
