package renderer

import (
	"fmt"

	"github.com/vekio/overmind/internal/domain"
)

const (
	pageTemplate     = "page"
	journalTemplate  = "journal"
	inboxTemplate    = "inbox"
	bookmarkTemplate = "bookmark"
	personTemplate   = "person"
)

func templateName(note domain.Note) (string, error) {
	switch note.(type) {
	case domain.Page:
		return pageTemplate, nil
	case domain.Person:
		return personTemplate, nil
	case domain.Journal:
		return journalTemplate, nil
	case domain.Inbox:
		return inboxTemplate, nil
	case domain.Bookmark:
		return bookmarkTemplate, nil
	default:
		return "", fmt.Errorf("render unsupported note type %T", note)
	}
}
