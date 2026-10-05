package inbox

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// UpdateCommand replaces the editable fields of an existing inbox.
// Empty optional values clear their previous contents.
type UpdateCommand struct {
	ID      string
	Content string
	Tags    []string
}

// UpdateResult contains the inbox returned after a successful update operation.
type UpdateResult struct {
	Inbox *inbox.Inbox
}

// UpdateHandler validates replacement values and retains the inbox's identity and creation time.
type UpdateHandler struct {
	repository ports.InboxRepository
}

// NewUpdateHandler creates the use-case handler with its dependencies.
func NewUpdateHandler(repository ports.InboxRepository) *UpdateHandler {
	return &UpdateHandler{
		repository: repository,
	}
}

// Handle loads the existing inbox, validates its replacement and persists the updated entity.
func (handler *UpdateHandler) Handle(ctx context.Context, command UpdateCommand) (UpdateResult, error) {
	if err := ctx.Err(); err != nil {
		return UpdateResult{}, err
	}
	if handler.repository == nil {
		return UpdateResult{}, fmt.Errorf("update inbox dependencies are not configured")
	}
	values := make([]shared.Tag, 0, len(command.Tags))
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
	id, err := uuid.Parse(command.ID)
	if err != nil || id == uuid.Nil() {
		return UpdateResult{}, fmt.Errorf("valid inbox ID is required")
	}
	existing, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("read inbox: %w", err)
	}
	if existing == nil || existing.ID() != id {
		return UpdateResult{}, fmt.Errorf("inbox identity does not match requested ID")
	}
	now := time.Now()
	metadata, err := existing.Metadata().Updated(now)
	if err != nil {
		return UpdateResult{}, err
	}
	entity, err := inbox.NewInbox(id, command.Content, tags, metadata)
	if err != nil {
		return UpdateResult{}, err
	}
	if err := handler.repository.Update(ctx, entity); err != nil {
		return UpdateResult{}, fmt.Errorf("update inbox: %w", err)
	}
	return UpdateResult{Inbox: entity}, nil
}
