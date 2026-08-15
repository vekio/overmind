package lexer

import "strings"

func matchAdmonitionParagraph(raw string) (LineToken, bool) {
	labels := [...]struct {
		label string
		kind  AdmonitionKind
	}{
		{label: "NOTE", kind: AdmonitionNote},
		{label: "TIP", kind: AdmonitionTip},
		{label: "IMPORTANT", kind: AdmonitionImportant},
		{label: "CAUTION", kind: AdmonitionCaution},
		{label: "WARNING", kind: AdmonitionWarning},
	}

	for _, candidate := range labels {
		prefix := candidate.label + ": "
		if !strings.HasPrefix(raw, prefix) || len(raw) == len(prefix) {
			continue
		}
		if isHorizontalSpace(raw[len(prefix)]) {
			return LineToken{}, false
		}
		return admonitionLineToken(raw, AdmonitionPayload{
			Kind:              candidate.kind,
			Label:             candidate.label,
			LabelByteOffset:   0,
			Content:           raw[len(prefix):],
			ContentByteOffset: len(prefix),
		}), true
	}
	return LineToken{}, false
}
