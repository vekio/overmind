package journal

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

// ErrAlreadyExists identifies a journal date already present in the index.
var ErrAlreadyExists = ports.ErrJournalAlreadyExists

// CreateCommand accepts an optional YYYY-MM-DD date; empty means today.
type CreateCommand struct {
	Date    string
	Content string
	Tags    []string
}

// CreateResult contains the journal returned after a successful create operation.
type CreateResult struct {
	Journal *journals.Journal
}

// CreateHandler validates input, assigns identity and lifecycle metadata, and saves a journal.
type CreateHandler struct {
	repository ports.JournalRepository
	ids        ports.IDGenerator
}

// NewCreateHandler creates the use-case handler with its dependencies.
func NewCreateHandler(repository ports.JournalRepository, ids ports.IDGenerator) *CreateHandler {
	return &CreateHandler{
		repository: repository,
		ids:        ids,
	}
}

// Handle validates raw input before persisting a new journal and its index projection.
func (handler *CreateHandler) Handle(ctx context.Context, command CreateCommand) (CreateResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreateResult{}, fmt.Errorf("create journal dependencies are not configured")
	}
	dateText := strings.TrimSpace(command.Date)
	if dateText == "" {
		dateText = time.Now().Format(time.DateOnly)
	}
	date, err := calendar.NewDate(dateText)
	if err != nil {
		return CreateResult{}, err
	}

	values := make([]shared.Tag, 0, len(command.Tags))
	for _, raw := range command.Tags {
		tag, err := shared.NewTag(raw)
		if err != nil {
			return CreateResult{}, err
		}
		values = append(values, tag)
	}
	tags, err := shared.NewTags(values...)
	if err != nil {
		return CreateResult{}, err
	}
	now := time.Now()
	metadata, err := shared.NewEntityMetadata(now, now)
	if err != nil {
		return CreateResult{}, err
	}
	entity, err := journals.NewJournal(handler.ids.Generate(), date, command.Content, tags, metadata)
	if err != nil {
		return CreateResult{}, err
	}
	if err := handler.repository.Save(ctx, entity); err != nil {
		return CreateResult{}, fmt.Errorf("save journal: %w", err)
	}
	return CreateResult{Journal: entity}, nil
}
