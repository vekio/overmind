package app

import (
	"context"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

// CreatePersonCommand contains the input for creating a person.
type CreatePersonCommand struct {
	Name   domain.Title
	Groups domain.Groups
	Tags   domain.Tags
}

// CreatePersonResult contains the created person and its path.
type CreatePersonResult struct {
	Person domain.Person
	Path   string
}

// CreatePersonHandler creates person notes.
type CreatePersonHandler struct {
	saver *noteSaver
	ids   ports.IDGenerator
}

func newCreatePersonHandler(saver *noteSaver, ids ports.IDGenerator) *CreatePersonHandler {
	return &CreatePersonHandler{saver: saver, ids: ids}
}

// Handle creates, persists and indexes a person.
func (handler *CreatePersonHandler) Handle(ctx context.Context, command CreatePersonCommand) (CreatePersonResult, error) {
	person, err := domain.NewPerson(handler.ids.Generate(), command.Name, command.Groups, command.Tags, time.Now())
	if err != nil {
		return CreatePersonResult{}, err
	}
	path, err := handler.saver.save(ctx, person)
	if err != nil {
		return CreatePersonResult{}, err
	}

	return CreatePersonResult{Person: person, Path: path}, nil
}
