package lexer

func matchContinuation(raw string) (LineToken, bool) {
	if raw != "+" {
		return LineToken{}, false
	}
	return continuationLineToken(raw), true
}
