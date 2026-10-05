package codecs

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/ports"
	"github.com/vekio/overmind/pkg/asciidoc"
	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
	"github.com/vekio/overmind/pkg/asciidoc/edit"
)

// RawNoteCodec uses the same domain codecs as the typed repositories.
type RawNoteCodec struct{}

var _ ports.RawNoteCodec = RawNoteCodec{}

// Decode validates managed metadata, the supported AsciiDoc syntax and typed domain fields.
// An empty expected kind accepts any supported note type; diagnostics include line and column.
func (codec RawNoteCodec) Decode(kind ports.NoteKind, source []byte) (ports.RawNoteHeader, error) {
	header, err := codec.Inspect(source)
	if err != nil {
		return ports.RawNoteHeader{}, err
	}
	if kind != "" && header.Kind != kind {
		return ports.RawNoteHeader{}, fmt.Errorf("note type cannot change: expected %s, found %s", kind, header.Kind)
	}
	processed := asciidoc.Process(source)
	if processed.HasErrors() {
		var messages []string
		for _, item := range processed.Diagnostics {
			if item.Severity == diagnostic.SeverityError {
				messages = append(messages, fmt.Sprintf("line %d, column %d: %s", item.Source.Start.Line, item.Source.Start.Column, item.Message))
			}
		}
		return ports.RawNoteHeader{}, fmt.Errorf("invalid AsciiDoc document:\n%s", strings.Join(messages, "\n"))
	}
	switch header.Kind {
	case ports.NoteKindPage:
		_, err = (PageCodec{}).Decode(source)
	case ports.NoteKindInbox:
		_, err = (InboxCodec{}).Decode(source)
	case ports.NoteKindBookmark:
		_, err = (BookmarkCodec{}).Decode(source)
	case ports.NoteKindPerson:
		_, err = (PersonCodec{}).Decode(source)
	case ports.NoteKindHabit:
		_, err = (HabitCodec{}).Decode(source)
	case ports.NoteKindJournal:
		_, err = (JournalCodec{}).Decode(source)
	}
	if err != nil {
		return ports.RawNoteHeader{}, err
	}
	return header, nil
}

// Inspect deliberately parses only the header so invalid bodies can be repaired.
func (RawNoteCodec) Inspect(source []byte) (ports.RawNoteHeader, error) {
	note, err := decodeHeader(source, "")
	if err != nil {
		return ports.RawNoteHeader{}, err
	}
	draft := ports.RawNoteHeader{ID: note.id, Kind: ports.NoteKind(note.kind), Metadata: note.metadata}
	switch draft.Kind {
	case ports.NoteKindPage, ports.NoteKindInbox, ports.NoteKindBookmark, ports.NoteKindPerson, ports.NoteKindHabit:
	case ports.NoteKindJournal:
		value, err := required(note.header.Attributes, "overmind-date")
		if err != nil {
			return ports.RawNoteHeader{}, err
		}
		date, err := calendar.NewDate(value)
		if err != nil {
			return ports.RawNoteHeader{}, err
		}
		draft.Date = date.String()
	default:
		return ports.RawNoteHeader{}, fmt.Errorf("unsupported note type %q", draft.Kind)
	}
	return draft, nil
}

// Stamp replaces the value span of the effective header declaration. Everything
// outside that span, including line endings and body attributes, is unchanged.
func (RawNoteCodec) Stamp(source []byte, at time.Time) ([]byte, error) {
	header, _ := splitDocument(source)
	parsed := asciidoc.Parse(header)
	if parsed.HasErrors() {
		return nil, fmt.Errorf("invalid note header: %v", parsed.Diagnostics)
	}
	var effective *ast.AttributeEntry
	for _, block := range parsed.Document.Blocks {
		attribute, ok := block.(*ast.AttributeEntry)
		if ok && attribute.Header && attribute.Name == "overmind-updated-at" {
			effective = attribute
		}
	}
	if effective == nil || effective.Operation != ast.AttributeSet {
		return nil, fmt.Errorf("overmind-updated-at is required")
	}
	span := effective.ValueSource
	value := source[span.Start.Offset:span.End.Offset]
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("overmind-updated-at value is required")
	}
	span.Start.Offset += bytes.Index(value, trimmed)
	span.End.Offset = span.Start.Offset + len(trimmed)
	return edit.Apply(source, []edit.TextEdit{{Source: span, Replacement: []byte(at.UTC().Format(time.RFC3339Nano))}})
}
