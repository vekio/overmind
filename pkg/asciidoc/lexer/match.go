package lexer

func isHorizontalSpace(value byte) bool {
	return value == ' ' || value == '\t'
}

func trimHorizontalBounds(raw string, start, end int) (int, int) {
	for start < end && isHorizontalSpace(raw[start]) {
		start++
	}
	for end > start && isHorizontalSpace(raw[end-1]) {
		end--
	}
	return start, end
}
