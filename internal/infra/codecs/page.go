package codecs

import (
	"fmt"
	"strings"
	"uuid"

	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

type PageCodec struct{}

var _ ports.PageEncoder = PageCodec{}

func (PageCodec) Encode(entity *pages.Page) ([]byte, error) {
	if entity == nil || entity.ID() == uuid.Nil() {
		return nil, fmt.Errorf("initialized page is required")
	}
	return render("page", entity)
}

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
	return pages.NewPage(note.id, title, area, note.tags, note.metadata)
}
