package lexer

import "strings"

func matchBlankLine(raw string) (LineToken, bool) {
	if strings.Trim(raw, " \t") != "" {
		return LineToken{}, false
	}
	return blankLineToken(raw), true
}
