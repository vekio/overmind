package codecs

import (
	"fmt"
	"strings"
	"uuid"

	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

func (PersonCodec) Encode(person *persons.Person) ([]byte, error) {
	if person == nil || person.ID() == uuid.Nil() {
		return nil, fmt.Errorf("initialized person is required")
	}
	return render("person", person)
}

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
	return persons.NewPerson(note.id, title, groups, note.tags, note.metadata)
}

type PersonCodec struct{}

var _ ports.PersonCodec = PersonCodec{}
