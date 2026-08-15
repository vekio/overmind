package lexer

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
)

// ErrNilReader is returned when a scanner has no input reader.
var ErrNilReader = errors.New("asciidoc lexer: nil reader")

// Scanner incrementally reads and classifies physical AsciiDoc source lines.
// It is not safe for concurrent use.
type Scanner struct {
	reader     *bufio.Reader
	next       Position
	terminal   error
	errPending bool
}

// New creates a scanner positioned at the beginning of reader.
func New(reader io.Reader) *Scanner {
	scanner := &Scanner{next: sourceOrigin}
	if reader == nil {
		scanner.terminal = ErrNilReader
		scanner.errPending = true
		return scanner
	}
	scanner.reader = bufio.NewReader(reader)
	return scanner
}

// Next returns the next classified physical line. At the end of the input it
// returns a stable LineEOF token and a nil error. Read failures are returned
// after any source bytes already read have first been emitted as a token.
func (s *Scanner) Next() (LineToken, error) {
	if s == nil {
		return LineToken{}, ErrNilReader
	}
	if s.errPending {
		s.errPending = false
		err := s.terminal
		s.terminal = io.EOF
		return LineToken{}, err
	}
	if errors.Is(s.terminal, io.EOF) {
		return eofLineToken(s.next), nil
	}

	start := s.next
	var raw bytes.Buffer
	for {
		value, err := s.reader.ReadByte()
		if err != nil {
			if raw.Len() == 0 {
				if errors.Is(err, io.EOF) {
					s.terminal = io.EOF
					return eofLineToken(s.next), nil
				}
				s.terminal = io.EOF
				return LineToken{}, err
			}

			s.terminal = err
			if !errors.Is(err, io.EOF) {
				s.errPending = true
			}
			return s.emit(raw.String(), start, LineEndingNone), nil
		}

		switch value {
		case '\n':
			return s.emit(raw.String(), start, LineEndingLF), nil
		case '\r':
			ending := LineEndingCR
			if next, err := s.reader.Peek(1); err == nil && next[0] == '\n' {
				_, _ = s.reader.ReadByte()
				ending = LineEndingCRLF
			}
			return s.emit(raw.String(), start, ending), nil
		default:
			raw.WriteByte(value)
		}
	}
}

func (s *Scanner) emit(raw string, start Position, ending LineEnding) LineToken {
	end := Position{
		Offset: start.Offset + len(raw),
		Line:   start.Line,
		Column: start.Column + utf8.RuneCountInString(raw),
	}
	token := matchLine(raw)
	token.Ending = ending
	token.Source = Span{Start: start, End: end}

	endingWidth := len(ending.Text())
	if ending == LineEndingNone {
		s.next = end
	} else {
		s.next = Position{
			Offset: end.Offset + endingWidth,
			Line:   end.Line + 1,
			Column: 1,
		}
	}
	return token
}
