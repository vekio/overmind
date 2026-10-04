package app

import (
	"context"
	"fmt"
	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
	"time"
)

type CreatePersonCommand struct {
	Name   string
	Groups []string
	Tags   []string
}
type CreatePersonResult struct{ Person *persons.Person }
type CreatePersonHandler struct {
	repository ports.PersonRepository
	ids        ports.IDGenerator
}

func newCreatePersonHandler(repository ports.PersonRepository, ids ports.IDGenerator) *CreatePersonHandler {
	return &CreatePersonHandler{repository: repository, ids: ids}
}
func (handler *CreatePersonHandler) Handle(ctx context.Context, command CreatePersonCommand) (CreatePersonResult, error) {
	if err := ctx.Err(); err != nil {
		return CreatePersonResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreatePersonResult{}, fmt.Errorf("create person dependencies are not configured")
	}
	name, err := shared.NewTitle(command.Name)
	if err != nil {
		return CreatePersonResult{}, err
	}
	groupValues := make([]persons.Group, 0, len(command.Groups))
	for _, value := range command.Groups {
		group, err := persons.NewGroup(value)
		if err != nil {
			return CreatePersonResult{}, err
		}
		groupValues = append(groupValues, group)
	}
	groups, err := persons.NewGroups(groupValues...)
	if err != nil {
		return CreatePersonResult{}, err
	}
	tagValues := make([]shared.Tag, 0, len(command.Tags))
	for _, value := range command.Tags {
		tag, err := shared.NewTag(value)
		if err != nil {
			return CreatePersonResult{}, err
		}
		tagValues = append(tagValues, tag)
	}
	tags, err := shared.NewTags(tagValues...)
	if err != nil {
		return CreatePersonResult{}, err
	}
	now := time.Now()
	metadata, err := shared.NewEntityMetadata(now, now)
	if err != nil {
		return CreatePersonResult{}, err
	}
	person, err := persons.NewPerson(handler.ids.Generate(), name, groups, tags, metadata)
	if err != nil {
		return CreatePersonResult{}, err
	}
	if err := handler.repository.Save(ctx, person); err != nil {
		return CreatePersonResult{}, fmt.Errorf("save person: %w", err)
	}
	return CreatePersonResult{Person: person}, nil
}
