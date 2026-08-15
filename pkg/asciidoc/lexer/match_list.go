package lexer

func matchListItem(raw string) (LineToken, bool) {
	if len(raw) < 3 || (raw[0] != '*' && raw[0] != '.') {
		return LineToken{}, false
	}

	markerEnd := 1
	for markerEnd < len(raw) && raw[markerEnd] == raw[0] {
		markerEnd++
	}
	if markerEnd == len(raw) || !isHorizontalSpace(raw[markerEnd]) {
		return LineToken{}, false
	}

	principalStart := markerEnd
	for principalStart < len(raw) && isHorizontalSpace(raw[principalStart]) {
		principalStart++
	}
	if principalStart == len(raw) {
		return LineToken{}, false
	}

	listKind := ListUnordered
	if raw[0] == '.' {
		listKind = ListOrdered
	}

	return listItemLineToken(raw, ListPayload{
		Kind:                listKind,
		Level:               markerEnd,
		Marker:              raw[:markerEnd],
		MarkerByteOffset:    0,
		Principal:           raw[principalStart:],
		PrincipalByteOffset: principalStart,
	}), true
}
