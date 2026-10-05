package codecs

import (
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/ports"
)

// BookmarkCodec encodes managed headers and decodes validated bookmark entities.
type BookmarkCodec struct{}

var _ ports.BookmarkCodec = BookmarkCodec{}

// Encode renders managed attributes and the entity body; unrelated source formatting is not retained.
func (BookmarkCodec) Encode(entity *bookmarks.Bookmark) ([]byte, error) {
	if entity == nil || entity.ID() == uuid.Nil() {
		return nil, fmt.Errorf("initialized bookmark is required")
	}
	return render("bookmark", entity)
}

// Decode validates the bookmark header and domain values while keeping the body verbatim.
// It does not validate the body's AsciiDoc syntax.
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
	entity, err := bookmarks.NewBookmark(note.id, url, note.tags, note.metadata)
	if err != nil {
		return nil, err
	}
	if err := entity.Rewrite(note.body, note.metadata.UpdatedAt()); err != nil {
		return nil, err
	}
	return entity, nil
}
