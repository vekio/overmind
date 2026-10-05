package notes

import (
	"context"
	"fmt"
	"strings"

	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// ListQuery filters the index; zero Limit defaults to 100 notes.
type ListQuery struct {
	Type   string
	Tag    string
	Limit  int
	Offset int
}

// ListResult contains one page of indexed note summaries.
type ListResult struct {
	Notes []ports.NoteSummary
}

// ListHandler validates filters and queries derived note summaries.
type ListHandler struct {
	index ports.Index
}

// NewListHandler creates the use-case handler with its dependencies.
func NewListHandler(index ports.Index) *ListHandler {
	return &ListHandler{
		index: index,
	}
}

// Handle normalizes type and tag filters and applies bounded pagination to the index query.
func (handler *ListHandler) Handle(ctx context.Context, query ListQuery) (ListResult, error) {
	if err := ctx.Err(); err != nil {
		return ListResult{}, err
	}
	kind := strings.ToLower(strings.TrimSpace(query.Type))
	switch kind {
	case "", "habit", "person", "bookmark", "inbox", "page", "journal":
	default:
		return ListResult{}, fmt.Errorf("unsupported note type %q: use habit, person, bookmark, inbox, page or journal", query.Type)
	}
	tag := strings.TrimSpace(query.Tag)
	if tag != "" {
		value, err := shared.NewTag(tag)
		if err != nil {
			return ListResult{}, err
		}
		tag = value.String()
	}
	if query.Limit < 0 || query.Limit > 1000 || query.Offset < 0 {
		return ListResult{}, fmt.Errorf("limit must be between 0 and 1000 and offset must be nonnegative")
	}
	limit := query.Limit
	if limit == 0 {
		limit = 100
	}
	if handler.index == nil {
		return ListResult{}, fmt.Errorf("note index is not configured")
	}
	notes, err := handler.index.FindNotes(ctx, ports.NoteFilter{Type: kind, Tag: tag, Limit: limit, Offset: query.Offset})
	if err != nil {
		return ListResult{}, fmt.Errorf("find notes: %w", err)
	}
	return ListResult{Notes: notes}, nil
}
