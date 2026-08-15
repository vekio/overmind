package lexer

// LineEnding identifies the exact byte sequence that terminated a source
// line. LineEndingNone is used for a final line without a terminator and EOF.
type LineEnding uint8

const (
	LineEndingNone LineEnding = iota
	LineEndingLF
	LineEndingCRLF
	LineEndingCR
)

// Text returns the original line-ending bytes as a string.
func (e LineEnding) Text() string {
	switch e {
	case LineEndingLF:
		return "\n"
	case LineEndingCRLF:
		return "\r\n"
	case LineEndingCR:
		return "\r"
	default:
		return ""
	}
}

// String returns a stable name for the line ending.
func (e LineEnding) String() string {
	switch e {
	case LineEndingNone:
		return "NONE"
	case LineEndingLF:
		return "LF"
	case LineEndingCRLF:
		return "CRLF"
	case LineEndingCR:
		return "CR"
	default:
		return "UNKNOWN"
	}
}
