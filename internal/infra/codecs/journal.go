package codecs

import (
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/ports"
)

type JournalCodec struct{}

var _ ports.JournalEncoder = JournalCodec{}

func (JournalCodec) Encode(entity *journals.Journal) ([]byte, error) {
	if entity == nil || entity.ID() == uuid.Nil() {
		return nil, fmt.Errorf("initialized journal is required")
	}
	return render("journal", entity)
}

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
	return journals.NewJournal(note.id, date, note.tags, note.metadata)
}
