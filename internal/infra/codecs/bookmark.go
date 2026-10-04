package codecs

import (
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/ports"
)

type BookmarkCodec struct{}

var _ ports.BookmarkEncoder = BookmarkCodec{}

func (BookmarkCodec) Encode(entity *bookmarks.Bookmark) ([]byte, error) {
	if entity == nil || entity.ID() == uuid.Nil() {
		return nil, fmt.Errorf("initialized bookmark is required")
	}
	return render("bookmark", entity)
}

func (BookmarkCodec) Decode(source []byte) (*bookmarks.Bookmark, error) {
	note, err := decodeHeader(source, ports.NoteKindBookmark.String())
	if err != nil {
		return nil, err
	}
	text, err := required(note.header.Attributes, "overmind-url")
	if err != nil {
		return nil, err
	}
	url, err := bookmarks.NewURL(text)
	if err != nil {
		return nil, err
	}
	return bookmarks.NewBookmark(note.id, url, note.tags, note.metadata)
}
