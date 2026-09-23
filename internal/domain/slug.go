package domain

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Slugify applies the current slugging policy.
// It lowercases, strips diacritics and collapses separators into hyphens.
func Slugify(value string) string {
	var b strings.Builder
	lastWasHyphen := false

	for _, r := range norm.NFD.String(strings.ToLower(value)) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastWasHyphen = false
			continue
		}

		if !lastWasHyphen {
			b.WriteByte('-')
			lastWasHyphen = true
		}
	}

	return strings.Trim(b.String(), "-")
}
