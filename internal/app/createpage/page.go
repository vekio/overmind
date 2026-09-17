package createpage

import (
	"fmt"
	"strings"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

// newPageFromCommand validates user input before generating an identifier.
func (handler *CreatePageHandler) newPageFromCommand(command CreatePageCommand) (domain.Page, error) {
	title, err := domain.NewTitle(command.Title)
	if err != nil {
		return domain.Page{}, fmt.Errorf("validate title: %w", err)
	}
	area, err := optionalArea(command.Area)
	if err != nil {
		return domain.Page{}, fmt.Errorf("validate area: %w", err)
	}
	tags, err := tagsFromStrings(command.Tags)
	if err != nil {
		return domain.Page{}, fmt.Errorf("validate tags: %w", err)
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
			return domain.Tags{}, fmt.Errorf("tag %d: %w", index+1, err)
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

func indexEntryForPage(page domain.Page) ports.IndexedDocument {
	return ports.IndexedDocument{
		ID:        page.ID(),
		Kind:      page.Kind(),
		Title:     page.Title().String(),
		Area:      page.Area().String(),
		Tags:      page.Tags().Strings(),
		CreatedAt: page.CreatedAt(),
		UpdatedAt: page.UpdatedAt(),
	}
}
