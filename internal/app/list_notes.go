package app

import (
	"context"
	"fmt"
	"slices"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain"
	"github.com/vekio/overmind/internal/ports"
)

// ListedNote is the metadata of a note, independent of how clients display it.
type ListedNote struct {
	ID         uuid.UUID
	Kind       domain.NoteKind
	Attributes map[string]string
	Tags       []string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ListNotesResult contains notes ordered by most recent update first.
type ListNotesResult struct {
	Notes []ListedNote
}

// ListNotesHandler retrieves note metadata from the index.
type ListNotesHandler struct {
	lister ports.NoteLister
}

func newListNotesHandler(lister ports.NoteLister) *ListNotesHandler {
	return &ListNotesHandler{lister: lister}
}

func (handler *ListNotesHandler) Handle(ctx context.Context) (ListNotesResult, error) {
	records, err := handler.lister.List(ctx)
	if err != nil {
		return ListNotesResult{}, fmt.Errorf("list notes: %w", err)
	}
	notes := make([]ListedNote, 0, len(records))
	for _, record := range records {
		note := ListedNote{
			ID: record.ID, Kind: record.Kind,
			CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
			Attributes: make(map[string]string, len(record.Attributes)),
			Tags:       slices.Clone(record.Tags),
		}
		for _, attribute := range record.Attributes {
			note.Attributes[attribute.Name] = attribute.Value
		}
		notes = append(notes, note)
	}
	return ListNotesResult{Notes: notes}, nil
}
