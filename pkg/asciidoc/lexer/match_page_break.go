package lexer

func matchPageBreak(raw string) (LineToken, bool) {
	if raw != "<<<" {
		return LineToken{}, false
	}
	return pageBreakLineToken(raw), true
}
