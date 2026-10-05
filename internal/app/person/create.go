package person

import (
	"context"
	"fmt"
	"time"

	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// CreateCommand supplies raw values for a new person; optional collections may be empty.
type CreateCommand struct {
	Content string
	Name    string
	Groups  []string
	Tags    []string
}

// CreateResult contains the person returned after a successful create operation.
type CreateResult struct {
	Person *persons.Person
}

// CreateHandler validates input, assigns identity and lifecycle metadata, and saves a person.
type CreateHandler struct {
	repository ports.PersonRepository
	ids        ports.IDGenerator
}

// NewCreateHandler creates the use-case handler with its dependencies.
func NewCreateHandler(repository ports.PersonRepository, ids ports.IDGenerator) *CreateHandler {
	return &CreateHandler{
		repository: repository,
		ids:        ids,
	}
}

// Handle validates raw input before persisting a new person and its index projection.
func (handler *CreateHandler) Handle(ctx context.Context, command CreateCommand) (CreateResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateResult{}, err
	}
	if handler.repository == nil || handler.ids == nil {
		return CreateResult{}, fmt.Errorf("create person dependencies are not configured")
	}
	name, err := shared.NewTitle(command.Name)
	if err != nil {
		return CreateResult{}, err
	}
	groupValues := make([]persons.Group, 0, len(command.Groups))
	for _, value := range command.Groups {
		group, err := persons.NewGroup(value)
		if err != nil {
			return CreateResult{}, err
		}
		groupValues = append(groupValues, group)
	}
	groups, err := persons.NewGroups(groupValues...)
	if err != nil {
		return CreateResult{}, err
	}
	tagValues := make([]shared.Tag, 0, len(command.Tags))
	for _, value := range command.Tags {
		tag, err := shared.NewTag(value)
		if err != nil {
			return CreateResult{}, err
		}
		tagValues = append(tagValues, tag)
	}
	tags, err := shared.NewTags(tagValues...)
	if err != nil {
		return CreateResult{}, err
	}
	now := time.Now()
	metadata, err := shared.NewEntityMetadata(now, now)
	if err != nil {
		return CreateResult{}, err
	}
	person, err := persons.NewPerson(handler.ids.Generate(), name, groups, tags, metadata)
	if err != nil {
		return CreateResult{}, err
	}
	if err := person.Rewrite(command.Content, now); err != nil {
		return CreateResult{}, err
	}
	if err := handler.repository.Save(ctx, person); err != nil {
		return CreateResult{}, fmt.Errorf("save person: %w", err)
	}
	return CreateResult{Person: person}, nil
}
