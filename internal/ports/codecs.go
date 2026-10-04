package ports

import (
	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/persons"
)

// HabitCodec translates between a habit and its persisted document.
type HabitCodec interface {
	Encode(*habits.Habit) ([]byte, error)
	Decode([]byte) (*habits.Habit, error)
}

type PersonCodec interface {
	Encode(*persons.Person) ([]byte, error)
	Decode([]byte) (*persons.Person, error)
}

type BookmarkEncoder interface {
	Encode(*bookmarks.Bookmark) ([]byte, error)
}
type InboxEncoder interface {
	Encode(*inbox.Inbox) ([]byte, error)
}

type PageEncoder interface {
	Encode(*pages.Page) ([]byte, error)
}
type JournalEncoder interface {
	Encode(*journals.Journal) ([]byte, error)
}
