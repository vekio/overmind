package lexer

func matchThematicBreak(raw string) (LineToken, bool) {
	if raw != "'''" {
		return LineToken{}, false
	}
	return thematicBreakLineToken(raw), true
}
