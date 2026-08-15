package lexer

func matchLineComment(raw string) (LineToken, bool) {
	if len(raw) < 2 || raw[0] != '/' || raw[1] != '/' || (len(raw) > 2 && raw[2] == '/') {
		return LineToken{}, false
	}

	contentStart := 2
	if contentStart < len(raw) && raw[contentStart] == ' ' {
		contentStart++
	}
	return commentLineToken(raw, CommentPayload{
		Content:           raw[contentStart:],
		ContentByteOffset: contentStart,
	}), true
}
