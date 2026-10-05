package app

import "github.com/vekio/overmind/internal/ports"

// Dependencies contains the ports required by the application.
type Dependencies struct {
	Index        ports.Index
	RawNoteCodec ports.RawNoteCodec
	NoteStore    ports.NoteStore
	Pages        ports.PageRepository
	Journals     ports.JournalRepository
	Habits       ports.HabitRepository
	Persons      ports.PersonRepository
	Bookmarks    ports.BookmarkRepository
	Inbox        ports.InboxRepository
	IDGenerator  ports.IDGenerator
}
