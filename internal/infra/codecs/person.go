package codecs

import (
	"fmt"
	"strings"
	"uuid"

	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// Encode renders managed attributes and the entity body; unrelated source formatting is not retained.
func (PersonCodec) Encode(person *persons.Person) ([]byte, error) {
	if person == nil || person.ID() == uuid.Nil() {
		return nil, fmt.Errorf("initialized person is required")
	}
	return render("person", person)
}

// Decode validates the person header and domain values while keeping the body verbatim.
// It does not validate the body's AsciiDoc syntax.
func (PersonCodec) Decode(source []byte) (*persons.Person, error) {
	note, err := decodeHeader(source, ports.NoteKindPerson.String())
	if err != nil {
		return nil, err
	}
	header := note.header
	if header.Title == nil {
		return nil, fmt.Errorf("person title is required")
	}
	title, err := shared.NewTitle(header.Title.Text)
	if err != nil {
		return nil, err
	}
	var groupValues []persons.Group
	if text, ok := header.Attributes.Lookup("overmind-groups"); ok && strings.TrimSpace(text) != "" {
		for _, part := range strings.Split(text, ",") {
			group, err := persons.NewGroup(part)
			if err != nil {
				return nil, err
			}
			groupValues = append(groupValues, group)
		}
	}
	groups, err := persons.NewGroups(groupValues...)
	if err != nil {
		return nil, err
	}
	entity, err := persons.NewPerson(note.id, title, groups, note.tags, note.metadata)
	if err != nil {
		return nil, err
	}
	if err := entity.Rewrite(note.body, note.metadata.UpdatedAt()); err != nil {
		return nil, err
	}
	return entity, nil
}

// PersonCodec encodes managed headers and decodes validated person entities.
type PersonCodec struct{}

var _ ports.PersonCodec = PersonCodec{}
