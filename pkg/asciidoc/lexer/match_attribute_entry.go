package lexer

import (
	"strings"
)

func matchAttributeEntry(raw string) (LineToken, bool) {
	if len(raw) < 3 || raw[0] != ':' {
		return LineToken{}, false
	}

	closingOffset := strings.IndexByte(raw[1:], ':')
	if closingOffset < 0 {
		return LineToken{}, false
	}
	closing := closingOffset + 1
	name := raw[1:closing]
	nameStart := 1
	operation := AttributeSet

	prefixUnset := strings.HasPrefix(name, "!")
	suffixUnset := strings.HasSuffix(name, "!")
	if prefixUnset && suffixUnset {
		return LineToken{}, false
	}
	if prefixUnset {
		name = name[1:]
		nameStart++
		operation = AttributeUnset
	} else if suffixUnset {
		name = name[:len(name)-1]
		operation = AttributeUnset
	}
	if !isAttributeName(name) {
		return LineToken{}, false
	}

	valueStart := closing + 1
	if operation == AttributeUnset {
		if strings.Trim(raw[valueStart:], " \t") != "" {
			return LineToken{}, false
		}
		valueStart = len(raw)
	} else if valueStart < len(raw) {
		if !isHorizontalSpace(raw[valueStart]) {
			return LineToken{}, false
		}
		for valueStart < len(raw) && isHorizontalSpace(raw[valueStart]) {
			valueStart++
		}
	}

	return attributeEntryLineToken(raw, AttributeEntryPayload{
		Operation:       operation,
		Name:            name,
		NameByteOffset:  nameStart,
		Value:           raw[valueStart:],
		ValueByteOffset: valueStart,
	}), true
}

func isAttributeName(name string) bool {
	if name == "" || !isAttributeWordByte(name[0]) {
		return false
	}
	for index := 1; index < len(name); index++ {
		if !isAttributeWordByte(name[index]) && name[index] != '-' {
			return false
		}
	}
	return true
}

func isAttributeWordByte(value byte) bool {
	return value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9' ||
		value == '_'
}
