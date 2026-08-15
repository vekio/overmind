package lexer

import "strings"

func matchBlockMacro(raw string) (LineToken, bool) {
	separator := strings.Index(raw, "::")
	if separator <= 0 || !isMacroName(raw[:separator]) {
		return LineToken{}, false
	}

	targetStart := separator + 2
	if targetStart < len(raw) && isHorizontalSpace(raw[targetStart]) {
		return LineToken{}, false
	}
	attributesOpen := strings.LastIndexByte(raw[targetStart:], '[')
	if attributesOpen < 0 || raw[len(raw)-1] != ']' {
		return LineToken{}, false
	}
	attributesOpen += targetStart

	return blockMacroLineToken(raw, BlockMacroPayload{
		Name:                 raw[:separator],
		NameByteOffset:       0,
		Target:               raw[targetStart:attributesOpen],
		TargetByteOffset:     targetStart,
		Attributes:           raw[attributesOpen+1 : len(raw)-1],
		AttributesByteOffset: attributesOpen + 1,
	}), true
}

func isMacroName(name string) bool {
	if name == "" || !isASCIILetter(name[0]) {
		return false
	}
	for index := 1; index < len(name); index++ {
		value := name[index]
		if !isASCIILetter(value) && (value < '0' || value > '9') && value != '_' && value != '-' {
			return false
		}
	}
	return true
}

func isASCIILetter(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}
