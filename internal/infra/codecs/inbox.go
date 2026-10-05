package codecs

import (
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/ports"
)

// InboxCodec encodes managed headers and decodes validated inbox entities.
type InboxCodec struct{}

var _ ports.InboxCodec = InboxCodec{}

// Encode renders managed attributes and the entity body; unrelated source formatting is not retained.
func (InboxCodec) Encode(entity *inbox.Inbox) ([]byte, error) {
	if entity == nil || entity.ID() == uuid.Nil() {
		return nil, fmt.Errorf("initialized inbox is required")
	}
	return render("inbox", entity)
}

// Decode validates the inbox header and domain values while keeping the body verbatim.
// It does not validate the body's AsciiDoc syntax.
func (InboxCodec) Decode(source []byte) (*inbox.Inbox, error) {
	note, err := decodeHeader(source, ports.NoteKindInbox.String())
	if err != nil {
		return nil, err
	}
	return inbox.NewInbox(note.id, note.body, note.tags, note.metadata)
}
