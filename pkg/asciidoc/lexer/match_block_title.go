package lexer

func matchBlockTitle(raw string) (LineToken, bool) {
	if len(raw) < 2 || raw[0] != '.' || isHorizontalSpace(raw[1]) {
		return LineToken{}, false
	}
	return blockTitleLineToken(raw, BlockTitlePayload{
		Title:           raw[1:],
		TitleByteOffset: 1,
	}), true
}
