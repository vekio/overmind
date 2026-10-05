package journal

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// UpdateCommand replaces the editable fields of an existing journal.
// Empty optional values clear their previous contents.
type UpdateCommand struct {
	ID      string
	Content string
	Tags    []string
}

// UpdateResult contains the journal returned after a successful update operation.
type UpdateResult struct {
	Journal *journals.Journal
}

// UpdateHandler validates replacement values and retains the journal's identity and creation time.
type UpdateHandler struct {
	repository ports.JournalRepository
}

// NewUpdateHandler creates the use-case handler with its dependencies.
func NewUpdateHandler(repository ports.JournalRepository) *UpdateHandler {
	return &UpdateHandler{
		repository: repository,
	}
}

// Handle loads the existing journal, validates its replacement and persists the updated entity.
func (handler *UpdateHandler) Handle(ctx context.Context, command UpdateCommand) (UpdateResult, error) {
	if err := ctx.Err(); err != nil {
		return UpdateResult{}, err
	}

	id, err := uuid.Parse(command.ID)
	if err != nil || id == uuid.Nil() {
		return UpdateResult{}, fmt.Errorf("valid journal ID is required")
	}

	if handler.repository == nil {
		return UpdateResult{}, fmt.Errorf("journal repository is not configured")
	}

	var values []shared.Tag
	for _, raw := range command.Tags {
		tag, err := shared.NewTag(raw)
		if err != nil {
			return UpdateResult{}, err
		}

		values = append(values, tag)
	}

	tags, err := shared.NewTags(values...)
	if err != nil {
		return UpdateResult{}, err
	}

	existing, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("read journal: %w", err)
	}

	if existing == nil || existing.ID() != id {
		return UpdateResult{}, fmt.Errorf("journal identity does not match requested ID")
	}

	metadata, err := existing.Metadata().Updated(time.Now())
	if err != nil {
		return UpdateResult{}, err
	}

	entity, err := journals.NewJournal(id, existing.Date(), command.Content, tags, metadata)
	if err != nil {
		return UpdateResult{}, err
	}

	if err := handler.repository.Update(ctx, entity); err != nil {
		return UpdateResult{}, fmt.Errorf("update journal: %w", err)
	}

	return UpdateResult{Journal: entity}, nil
}
