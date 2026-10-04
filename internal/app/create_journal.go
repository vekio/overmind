package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

var ErrJournalAlreadyExists = ports.ErrJournalAlreadyExists

// CreateJournalCommand accepts an optional YYYY-MM-DD date; empty means today.
type CreateJournalCommand struct {
	Date string
	Tags []string
}
type CreateJournalResult struct{ Journal *journals.Journal }
type CreateJournalHandler struct {
	repository ports.JournalRepository
	ids        ports.IDGenerator
}

func newCreateJournalHandler(repository ports.JournalRepository, ids ports.IDGenerator) *CreateJournalHandler {
	return &CreateJournalHandler{repository: repository, ids: ids}
}
func (handler *CreateJournalHandler) Handle(ctx context.Context, command CreateJournalCommand) (CreateJournalResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateJournalResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreateJournalResult{}, fmt.Errorf("create journal dependencies are not configured")
	}
	dateText := strings.TrimSpace(command.Date)
	if dateText == "" {
		dateText = time.Now().Format(time.DateOnly)
	}
	date, err := calendar.NewDate(dateText)
	if err != nil {
		return CreateJournalResult{}, err
	}

	values := make([]shared.Tag, 0, len(command.Tags))
	for _, raw := range command.Tags {
		tag, err := shared.NewTag(raw)
		if err != nil {
			return CreateJournalResult{}, err
		}
		values = append(values, tag)
	}
	tags, err := shared.NewTags(values...)
	if err != nil {
		return CreateJournalResult{}, err
	}
	now := time.Now()
	metadata, err := shared.NewEntityMetadata(now, now)
	if err != nil {
		return CreateJournalResult{}, err
	}
	entity, err := journals.NewJournal(handler.ids.Generate(), date, tags, metadata)
	if err != nil {
		return CreateJournalResult{}, err
	}
	if err := handler.repository.Save(ctx, entity); err != nil {
		return CreateJournalResult{}, fmt.Errorf("save journal: %w", err)
	}
	return CreateJournalResult{Journal: entity}, nil
}
