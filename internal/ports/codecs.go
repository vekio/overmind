package ports

import (
	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/persons"
)

// HabitCodec maps habit entities to managed AsciiDoc documents.
// Decode validates the header and domain fields; body bytes remain opaque.
type HabitCodec interface {
	// Encode writes the managed header followed by the entity's content.
	Encode(*habits.Habit) ([]byte, error)

	// Decode requires the expected note kind and retains the original body bytes.
	Decode([]byte) (*habits.Habit, error)
}

// PersonCodec maps person entities to managed AsciiDoc documents.
// Decode validates the header and domain fields; body bytes remain opaque.
type PersonCodec interface {
	// Encode writes the managed header followed by the entity's content.
	Encode(*persons.Person) ([]byte, error)

	// Decode requires the expected note kind and retains the original body bytes.
	Decode([]byte) (*persons.Person, error)
}

// BookmarkCodec maps bookmark entities to managed AsciiDoc documents.
// Decode validates the header and domain fields; body bytes remain opaque.
type BookmarkCodec interface {
	// Encode writes the managed header followed by the entity's content.
	Encode(*bookmarks.Bookmark) ([]byte, error)

	// Decode requires the expected note kind and retains the original body bytes.
	Decode([]byte) (*bookmarks.Bookmark, error)
}

// InboxCodec maps inbox entities to managed AsciiDoc documents.
// Decode validates the header and domain fields; body bytes remain opaque.
type InboxCodec interface {
	// Encode writes the managed header followed by the entity's content.
	Encode(*inbox.Inbox) ([]byte, error)

	// Decode requires the expected note kind and retains the original body bytes.
	Decode([]byte) (*inbox.Inbox, error)
}

// PageCodec maps page entities to managed AsciiDoc documents.
// Decode validates the header and domain fields; body bytes remain opaque.
type PageCodec interface {
	// Encode writes the managed header followed by the entity's content.
	Encode(*pages.Page) ([]byte, error)

	// Decode requires the expected note kind and retains the original body bytes.
	Decode([]byte) (*pages.Page, error)
}

// JournalCodec maps journal entities to managed AsciiDoc documents.
// Decode validates the header and domain fields; body bytes remain opaque.
type JournalCodec interface {
	// Encode writes the managed header followed by the entity's content.
	Encode(*journals.Journal) ([]byte, error)

	// Decode requires the expected note kind and retains the original body bytes.
	Decode([]byte) (*journals.Journal, error)
}
