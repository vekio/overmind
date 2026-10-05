package codecs

import (
	"fmt"
	"strings"
	"uuid"

	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// PageCodec encodes managed headers and decodes validated page entities.
type PageCodec struct{}

var _ ports.PageCodec = PageCodec{}

// Encode renders managed attributes and the entity body; unrelated source formatting is not retained.
func (PageCodec) Encode(entity *pages.Page) ([]byte, error) {
	if entity == nil || entity.ID() == uuid.Nil() {
		return nil, fmt.Errorf("initialized page is required")
	}
	return render("page", entity)
}

// Decode validates the page header and domain values while keeping the body verbatim.
// It does not validate the body's AsciiDoc syntax.
func (PageCodec) Decode(source []byte) (*pages.Page, error) {
	note, err := decodeHeader(source, ports.NoteKindPage.String())
	if err != nil {
		return nil, err
	}
	if note.header.Title == nil {
		return nil, fmt.Errorf("page title is required")
	}
	title, err := shared.NewTitle(note.header.Title.Text)
	if err != nil {
		return nil, err
	}
	var area pages.Area
	if text, ok := note.header.Attributes.Lookup("overmind-area"); ok && strings.TrimSpace(text) != "" {
		area, err = pages.NewArea(text)
		if err != nil {
			return nil, err
		}
	}
	entity, err := pages.NewPage(note.id, title, area, note.tags, note.metadata)
	if err != nil {
		return nil, err
	}
	if err := entity.Rewrite(note.body, note.metadata.UpdatedAt()); err != nil {
		return nil, err
	}
	return entity, nil
}
