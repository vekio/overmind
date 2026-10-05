package ports

import (
	"context"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/persons"
)

// Index stores searchable projections derived from managed note documents.
// The vault is authoritative: updating this index never rewrites a document.
type Index interface {
	// FindNotes applies type/tag filters and pagination, ordered by update time
	// descending and UUID ascending. Labels are normalized to a single line.
	FindNotes(context.Context, NoteFilter) ([]NoteSummary, error)

	// FindGroups returns normalized group names in alphabetical order, including
	// only groups currently assigned to at least one person.
	FindGroups(context.Context) ([]string, error)

	// Delete removes a note's projection and its tag/group associations.
	// Deleting an absent UUID succeeds; the source document is left untouched.
	Delete(context.Context, uuid.UUID) error

	// Project decodes a document's managed header, verifies its UUID and updates
	// the appropriate typed projection. RawNoteCodec validates body syntax before
	// raw saves; index projection itself does not validate the entire body.
	Project(context.Context, Note) error

	// Rebuild replaces all projections in one transaction. Callback errors or
	// cancellation roll back the replacement, preserving the previous index.
	// The callback must use the supplied index and must not retain it afterward.
	Rebuild(context.Context, func(Index) error) error

	// UpsertPage atomically replaces a page's metadata, title, area and tags.
	// The path identifies its already-persisted source document.
	UpsertPage(context.Context, *pages.Page, string) error

	// UpsertJournal atomically replaces a journal's metadata, date and tags.
	// A date already assigned to another UUID returns ErrJournalAlreadyExists.
	UpsertJournal(context.Context, *journals.Journal, string) error

	// JournalExists reports whether a valid date is already assigned to a journal.
	JournalExists(context.Context, calendar.Date) (bool, error)

	// UpsertBookmark atomically replaces a bookmark's metadata, URL and tags.
	UpsertBookmark(context.Context, *bookmarks.Bookmark, string) error

	// UpsertInbox atomically replaces an inbox note's metadata, content and tags.
	UpsertInbox(context.Context, *inbox.Inbox, string) error

	// UpsertHabit atomically replaces a habit's metadata, title, goal and tags.
	UpsertHabit(context.Context, *habits.Habit, string) error

	// UpsertPerson atomically replaces a person's metadata, name, groups and tags.
	// Unused group definitions may remain stored but are excluded by FindGroups.
	UpsertPerson(context.Context, *persons.Person, string) error
}

// NoteFilter selects indexed notes. Empty type and tag match all notes.
type NoteFilter struct {
	Type   string
	Tag    string
	Limit  int
	Offset int
}

// NoteSummary is a read projection, independent of the note's domain entity.
type NoteSummary struct {
	ID        uuid.UUID
	Type      string
	Label     string
	Tags      []string
	UpdatedAt time.Time
}
