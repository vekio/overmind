package ports

import (
	"context"
	"errors"
	"uuid"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/persons"
)

var ErrHabitNotFound = errors.New("habit not found")

// HabitRepository stores habit definitions.
type HabitRepository interface {
	Save(context.Context, *habits.Habit) error
	ByID(context.Context, uuid.UUID) (*habits.Habit, error)
}

var ErrPersonNotFound = errors.New("person not found")

type PersonRepository interface {
	Save(context.Context, *persons.Person) error
	ByID(context.Context, uuid.UUID) (*persons.Person, error)
}

type BookmarkRepository interface {
	Save(context.Context, *bookmarks.Bookmark) error
}

type InboxRepository interface {
	Save(context.Context, *inbox.Inbox) error
}

var ErrJournalAlreadyExists = errors.New("journal already exists")

type PageRepository interface {
	Save(context.Context, *pages.Page) error
}
type JournalRepository interface {
	Save(context.Context, *journals.Journal) error
}
