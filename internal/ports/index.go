package ports

import (
	"context"
	"uuid"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/persons"
)

// Index maintains searchable projections; documents remain the source of truth.
type Index interface {
	UpsertPage(context.Context, *pages.Page, string) error
	UpsertJournal(context.Context, *journals.Journal, string) error
	JournalExists(context.Context, calendar.Date) (bool, error)
	UpsertBookmark(context.Context, *bookmarks.Bookmark, string) error
	UpsertInbox(context.Context, *inbox.Inbox, string) error
	UpsertHabit(context.Context, *habits.Habit, string) error
	UpsertPerson(context.Context, *persons.Person, string) error
}

// NoteIndexDeleter removes a note projection and its associated rows.
// Deleting an absent note is successful.
type NoteIndexDeleter interface {
	Delete(context.Context, uuid.UUID) error
}
