package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// ListNotesQuery filters the index; zero Limit defaults to 100 notes.
type ListNotesQuery struct {
	Type   string
	Tag    string
	Limit  int
	Offset int
}

type ListNotesResult struct{ Notes []ports.NoteSummary }

type ListNotesHandler struct{ finder ports.NoteFinder }

func newListNotesHandler(finder ports.NoteFinder) *ListNotesHandler {
	return &ListNotesHandler{finder: finder}
}

func (handler *ListNotesHandler) Handle(ctx context.Context, query ListNotesQuery) (ListNotesResult, error) {
	if err := ctx.Err(); err != nil {
		return ListNotesResult{}, err
	}
	kind := strings.ToLower(strings.TrimSpace(query.Type))
	switch kind {
	case "", "habit", "person", "bookmark", "inbox", "page", "journal":
	default:
		return ListNotesResult{}, fmt.Errorf("unsupported note type %q: use habit, person, bookmark, inbox, page or journal", query.Type)
	}
	tag := strings.TrimSpace(query.Tag)
	if tag != "" {
		value, err := shared.NewTag(tag)
		if err != nil {
			return ListNotesResult{}, err
		}
		tag = value.String()
	}
	if query.Limit < 0 || query.Limit > 1000 || query.Offset < 0 {
		return ListNotesResult{}, fmt.Errorf("limit must be between 0 and 1000 and offset must be nonnegative")
	}
	limit := query.Limit
	if limit == 0 {
		limit = 100
	}
	if handler.finder == nil {
		return ListNotesResult{}, fmt.Errorf("note finder is not configured")
	}
	notes, err := handler.finder.FindNotes(ctx, ports.NoteFilter{Type: kind, Tag: tag, Limit: limit, Offset: query.Offset})
	if err != nil {
		return ListNotesResult{}, fmt.Errorf("find notes: %w", err)
	}
	return ListNotesResult{Notes: notes}, nil
}
