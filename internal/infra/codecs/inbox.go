package codecs

import (
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/ports"
)

type InboxCodec struct{}

var _ ports.InboxEncoder = InboxCodec{}

func (InboxCodec) Encode(entity *inbox.Inbox) ([]byte, error) {
	if entity == nil || entity.ID() == uuid.Nil() {
		return nil, fmt.Errorf("initialized inbox is required")
	}
	return render("inbox", entity)
}

func (InboxCodec) Decode(source []byte) (*inbox.Inbox, error) {
	note, err := decodeHeader(source, ports.NoteKindInbox.String())
	if err != nil {
		return nil, err
	}
	return inbox.NewInbox(note.id, note.body, note.tags, note.metadata)
}
