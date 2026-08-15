package lexer

import (
	"strings"
	"unicode"
)

func matchAnchor(raw string) (LineToken, bool) {
	if len(raw) < 5 || !strings.HasPrefix(raw, "[[") || !strings.HasSuffix(raw, "]]") {
		return LineToken{}, false
	}

	contentStart := 2
	contentEnd := len(raw) - 2
	comma := strings.IndexByte(raw[contentStart:contentEnd], ',')
	idEnd := contentEnd
	if comma >= 0 {
		comma += contentStart
		idEnd = comma
	}
	idStart, idEnd := trimHorizontalBounds(raw, contentStart, idEnd)
	id := raw[idStart:idEnd]
	refText := ""
	refTextStart := 0
	if comma >= 0 {
		refTextStart, contentEnd = trimHorizontalBounds(raw, comma+1, contentEnd)
		refText = raw[refTextStart:contentEnd]
	}
	if id == "" ||
		strings.ContainsAny(id, "[]\r\n") ||
		strings.IndexFunc(id, unicode.IsSpace) >= 0 ||
		(comma >= 0 && refText == "") {
		return LineToken{}, false
	}

	return anchorLineToken(raw, AnchorPayload{
		ID:                id,
		IDByteOffset:      idStart,
		RefText:           refText,
		RefTextByteOffset: refTextStart,
	}), true
}
