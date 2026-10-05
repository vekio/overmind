package inbox

import (
	"context"
	"fmt"
	"time"

	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// CreateCommand supplies raw values for a new inbox; optional collections may be empty.
type CreateCommand struct {
	Content string
	Tags    []string
}

// CreateResult contains the inbox returned after a successful create operation.
type CreateResult struct {
	Inbox *inbox.Inbox
}

// CreateHandler validates input, assigns identity and lifecycle metadata, and saves a inbox.
type CreateHandler struct {
	repository ports.InboxRepository
	ids        ports.IDGenerator
}

// NewCreateHandler creates the use-case handler with its dependencies.
func NewCreateHandler(repository ports.InboxRepository, ids ports.IDGenerator) *CreateHandler {
	return &CreateHandler{
		repository: repository,
		ids:        ids,
	}
}

// Handle validates raw input before persisting a new inbox and its index projection.
func (handler *CreateHandler) Handle(ctx context.Context, command CreateCommand) (CreateResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreateResult{}, fmt.Errorf("create inbox dependencies are not configured")
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
	entity, err := inbox.NewInbox(handler.ids.Generate(), command.Content, tags, metadata)
	if err != nil {
		return CreateResult{}, err
	}
	if err := handler.repository.Save(ctx, entity); err != nil {
		return CreateResult{}, fmt.Errorf("save inbox: %w", err)
	}
	return CreateResult{Inbox: entity}, nil
}
