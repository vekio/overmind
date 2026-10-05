package person

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

// UpdateCommand replaces the editable fields of an existing person.
// Empty optional values clear their previous contents.
type UpdateCommand struct {
	Content string
	ID      string
	Name    string
	Groups  []string
	Tags    []string
}

// UpdateResult contains the person returned after a successful update operation.
type UpdateResult struct {
	Person *persons.Person
}

// UpdateHandler validates replacement values and retains the person's identity and creation time.
type UpdateHandler struct {
	repository ports.PersonRepository
}

// NewUpdateHandler creates the use-case handler with its dependencies.
func NewUpdateHandler(repository ports.PersonRepository) *UpdateHandler {
	return &UpdateHandler{
		repository: repository,
	}
}

// Handle loads the existing person, validates its replacement and persists the updated entity.
func (handler *UpdateHandler) Handle(ctx context.Context, command UpdateCommand) (UpdateResult, error) {
	if err := ctx.Err(); err != nil {
		return UpdateResult{}, err
	}
	if handler.repository == nil {
		return UpdateResult{}, fmt.Errorf("update person dependencies are not configured")
	}
	name, err := shared.NewTitle(command.Name)
	if err != nil {
		return UpdateResult{}, err
	}
	groupValues := make([]persons.Group, 0, len(command.Groups))
	for _, value := range command.Groups {
		group, err := persons.NewGroup(value)
		if err != nil {
			return UpdateResult{}, err
		}
		groupValues = append(groupValues, group)
	}
	groups, err := persons.NewGroups(groupValues...)
	if err != nil {
		return UpdateResult{}, err
	}
	tagValues := make([]shared.Tag, 0, len(command.Tags))
	for _, value := range command.Tags {
		tag, err := shared.NewTag(value)
		if err != nil {
			return UpdateResult{}, err
		}
		tagValues = append(tagValues, tag)
	}
	tags, err := shared.NewTags(tagValues...)
	if err != nil {
		return UpdateResult{}, err
	}
	id, err := uuid.Parse(command.ID)
	if err != nil || id == uuid.Nil() {
		return UpdateResult{}, fmt.Errorf("valid person ID is required")
	}
	existing, err := handler.repository.ByID(ctx, id)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("read person: %w", err)
	}
	if existing == nil || existing.ID() != id {
		return UpdateResult{}, fmt.Errorf("person identity does not match requested ID")
	}
	now := time.Now()
	metadata, err := existing.Metadata().Updated(now)
	if err != nil {
		return UpdateResult{}, err
	}
	person, err := persons.NewPerson(id, name, groups, tags, metadata)
	if err != nil {
		return UpdateResult{}, err
	}
	if err := person.Rewrite(command.Content, now); err != nil {
		return UpdateResult{}, err
	}
	if err := handler.repository.Update(ctx, person); err != nil {
		return UpdateResult{}, fmt.Errorf("update person: %w", err)
	}
	return UpdateResult{Person: person}, nil
}
