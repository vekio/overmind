package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

// ErrJournalAlreadyExists indicates that a journal exists for a date.
var ErrJournalAlreadyExists = errors.New("journal already exists")

// CreateJournalCommand contains the input for creating a journal.
type CreateJournalCommand struct {
	Tags domain.Tags
}

// CreateJournalResult contains the created journal and its path.
type CreateJournalResult struct {
	Journal domain.Journal
	Path    string
}

// CreateJournalHandler creates journal notes.
type CreateJournalHandler struct {
	saver *noteSaver
	index ports.NoteIndex
	ids   ports.IDGenerator
}

func newCreateJournalHandler(
	saver *noteSaver,
	index ports.NoteIndex,
	ids ports.IDGenerator,
) *CreateJournalHandler {
	return &CreateJournalHandler{saver: saver, index: index, ids: ids}
}

// Handle creates, persists and indexes today's journal note.
func (handler *CreateJournalHandler) Handle(ctx context.Context, command CreateJournalCommand) (CreateJournalResult, error) {
	now := time.Now()
	date := domain.DateFromTime(now)
	exists, err := handler.index.JournalExists(ctx, date)
	if err != nil {
		return CreateJournalResult{}, fmt.Errorf("check journal existence: %w", err)
	}
	if exists {
		return CreateJournalResult{}, ErrJournalAlreadyExists
	}

	journal, err := domain.NewJournal(handler.ids.Generate(), date, command.Tags, now)
	if err != nil {
		return CreateJournalResult{}, err
	}
	path, err := handler.saver.save(ctx, journal)
	if err != nil {
		return CreateJournalResult{}, err
	}

	return CreateJournalResult{Journal: journal, Path: path}, nil
}
