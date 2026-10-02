package app

import (
	"context"
	"fmt"
	"uuid"

	"github.com/vekio/overmind/internal/domain"
	"github.com/vekio/overmind/internal/ports"
)

// FindJournalHandler finds the journal for a date without creating it.
type FindJournalHandler struct {
	index ports.NoteIndex
}

func newFindJournalHandler(index ports.NoteIndex) *FindJournalHandler {
	return &FindJournalHandler{index: index}
}

func (handler *FindJournalHandler) Handle(ctx context.Context, date domain.Date) (uuid.UUID, bool, error) {
	id, exists, err := handler.index.JournalID(ctx, date)
	if err != nil {
		return uuid.Nil(), false, fmt.Errorf("find journal for %s: %w", date, err)
	}
	return id, exists, nil
}
