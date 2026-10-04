package app

import (
	"context"
	"fmt"
	"time"

	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

type CreateInboxCommand struct {
	Content string
	Tags    []string
}
type CreateInboxResult struct{ Inbox *inbox.Inbox }
type CreateInboxHandler struct {
	repository ports.InboxRepository
	ids        ports.IDGenerator
}

func newCreateInboxHandler(repository ports.InboxRepository, ids ports.IDGenerator) *CreateInboxHandler {
	return &CreateInboxHandler{repository: repository, ids: ids}
}
func (handler *CreateInboxHandler) Handle(ctx context.Context, command CreateInboxCommand) (CreateInboxResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateInboxResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreateInboxResult{}, fmt.Errorf("create inbox dependencies are not configured")
	}
	values := make([]shared.Tag, 0, len(command.Tags))
	for _, raw := range command.Tags {
		tag, err := shared.NewTag(raw)
		if err != nil {
			return CreateInboxResult{}, err
		}
		values = append(values, tag)
	}
	tags, err := shared.NewTags(values...)
	if err != nil {
		return CreateInboxResult{}, err
	}
	now := time.Now()
	metadata, err := shared.NewEntityMetadata(now, now)
	if err != nil {
		return CreateInboxResult{}, err
	}
	entity, err := inbox.NewInbox(handler.ids.Generate(), command.Content, tags, metadata)
	if err != nil {
		return CreateInboxResult{}, err
	}
	if err := handler.repository.Save(ctx, entity); err != nil {
		return CreateInboxResult{}, fmt.Errorf("save inbox: %w", err)
	}
	return CreateInboxResult{Inbox: entity}, nil
}
