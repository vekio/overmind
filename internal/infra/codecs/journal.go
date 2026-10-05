package codecs

import (
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/ports"
)

// JournalCodec encodes managed headers and decodes validated journal entities.
type JournalCodec struct{}

var _ ports.JournalCodec = JournalCodec{}

// Encode renders managed attributes and the entity body; unrelated source formatting is not retained.
func (JournalCodec) Encode(entity *journals.Journal) ([]byte, error) {
	if entity == nil || entity.ID() == uuid.Nil() {
		return nil, fmt.Errorf("initialized journal is required")
	}
	return render("journal", entity)
}

// Decode validates the journal header and domain values while keeping the body verbatim.
// It does not validate the body's AsciiDoc syntax.
func (JournalCodec) Decode(source []byte) (*journals.Journal, error) {
	note, err := decodeHeader(source, ports.NoteKindJournal.String())
	if err != nil {
		return nil, err
	}
	text, err := required(note.header.Attributes, "overmind-date")
	if err != nil {
		return nil, err
	}
	date, err := calendar.NewDate(text)
	if err != nil {
		return nil, err
	}
	return journals.NewJournal(note.id, date, note.body, note.tags, note.metadata)
}
