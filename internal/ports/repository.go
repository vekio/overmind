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

// ErrHabitNotFound identifies a missing habit source document.
var ErrHabitNotFound = errors.New("habit not found")

// HabitRepository reads source documents and persists habit entities with
// their index projections. An index error can occur after the document is saved.
type HabitRepository interface {
	// Save persists the entity's document and refreshes its searchable projection.
	Save(context.Context, *habits.Habit) error

	// ByID decodes the source document and verifies its kind and requested UUID.
	ByID(context.Context, uuid.UUID) (*habits.Habit, error)

	// Update requires an existing entity and preserves its creation timestamp.
	Update(context.Context, *habits.Habit) error
}

// ErrPersonNotFound identifies a missing person source document.
var ErrPersonNotFound = errors.New("person not found")

// PersonRepository reads source documents and persists person entities with
// their index projections. An index error can occur after the document is saved.
type PersonRepository interface {
	// Save persists the entity's document and refreshes its searchable projection.
	Save(context.Context, *persons.Person) error

	// ByID decodes the source document and verifies its kind and requested UUID.
	ByID(context.Context, uuid.UUID) (*persons.Person, error)

	// Update requires an existing entity and preserves its creation timestamp.
	Update(context.Context, *persons.Person) error
}

// BookmarkRepository reads source documents and persists bookmark entities with
// their index projections. An index error can occur after the document is saved.
type BookmarkRepository interface {
	// Save persists the entity's document and refreshes its searchable projection.
	Save(context.Context, *bookmarks.Bookmark) error

	// ByID decodes the source document and verifies its kind and requested UUID.
	ByID(context.Context, uuid.UUID) (*bookmarks.Bookmark, error)

	// Update requires an existing entity and preserves its creation timestamp.
	Update(context.Context, *bookmarks.Bookmark) error
}

// InboxRepository reads source documents and persists inbox entities with
// their index projections. An index error can occur after the document is saved.
type InboxRepository interface {
	// Save persists the entity's document and refreshes its searchable projection.
	Save(context.Context, *inbox.Inbox) error

	// ByID decodes the source document and verifies its kind and requested UUID.
	ByID(context.Context, uuid.UUID) (*inbox.Inbox, error)

	// Update requires an existing entity and preserves its creation timestamp.
	Update(context.Context, *inbox.Inbox) error
}

// PageRepository reads source documents and persists page entities with
// their index projections. An index error can occur after the document is saved.
type PageRepository interface {
	// Save persists the entity's document and refreshes its searchable projection.
	Save(context.Context, *pages.Page) error

	// ByID decodes the source document and verifies its kind and requested UUID.
	ByID(context.Context, uuid.UUID) (*pages.Page, error)

	// Update requires an existing entity and preserves its creation timestamp.
	Update(context.Context, *pages.Page) error
}

// ErrJournalAlreadyExists identifies a journal date assigned to another UUID.
var ErrJournalAlreadyExists = errors.New("journal already exists")

// JournalRepository reads source documents and persists journal entities with
// their index projections. An index error can occur after the document is saved.
type JournalRepository interface {
	// Save persists the entity's document and refreshes its searchable projection.
	Save(context.Context, *journals.Journal) error

	// ByID decodes the source document and verifies its kind and requested UUID.
	ByID(context.Context, uuid.UUID) (*journals.Journal, error)

	// Update requires an existing entity and preserves its creation timestamp.
	Update(context.Context, *journals.Journal) error
}
