package createpage

import (
	"fmt"
	"path"
	"strings"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

// newPageFromCommand validates user input before generating an identifier.
func (handler *CreatePageHandler) newPageFromCommand(command CreatePageCommand) (domain.Page, error) {
	title, err := domain.NewTitle(command.Title)
	if err != nil {
		return domain.Page{}, err
	}
	area, err := optionalArea(command.Area)
	if err != nil {
		return domain.Page{}, err
	}
	tags, err := tagsFromStrings(command.Tags)
	if err != nil {
		return domain.Page{}, err
	}

	generatedID, err := handler.idGenerator.Generate()
	if err != nil {
		return domain.Page{}, fmt.Errorf("generate document id: %w", err)
	}
	documentID, err := domain.NewDocumentID(generatedID)
	if err != nil {
		return domain.Page{}, fmt.Errorf("invalid generated document id: %w", err)
	}

	return domain.NewPage(documentID, title, tags, area, handler.clock.Now()), nil
}

func tagsFromStrings(values []string) (domain.Tags, error) {
	tags := make([]domain.Tag, len(values))
	for index, value := range values {
		tag, err := domain.NewTag(value)
		if err != nil {
			return domain.Tags{}, err
		}
		tags[index] = tag
	}
	return domain.NewTags(tags...)
}

func optionalArea(value string) (domain.Area, error) {
	if strings.TrimSpace(value) == "" {
		return domain.Area{}, nil
	}
	return domain.NewArea(value)
}

// pageDocumentKey returns the store-relative logical key. Storage adapters are
// responsible for translating it to their concrete path or object key.
func pageDocumentKey(page domain.Page) string {
	return path.Join(page.Kind().String(), page.Area().String(), page.Title().Slug()+".adoc")
}

// indexEntryForPage maps common metadata to columns and keeps only
// page-specific values in Attributes.
func indexEntryForPage(page domain.Page, documentKey string) ports.IndexedDocument {
	return ports.IndexedDocument{
		ID:        page.ID(),
		Path:      documentKey,
		Kind:      page.Kind(),
		Title:     page.Title().String(),
		Tags:      page.Tags().Strings(),
		CreatedAt: page.CreatedAt(),
		Attributes: map[string]string{
			"area": page.Area().String(),
		},
	}
}
