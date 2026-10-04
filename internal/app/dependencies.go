package app

import "github.com/vekio/overmind/internal/ports"

// Dependencies contains the ports required by the application.
type Dependencies struct {
	NoteStore      ports.NoteStore
	IndexDeleter   ports.NoteIndexDeleter
	NoteScanner    ports.NoteScanner
	NoteProjector  ports.NoteProjector
	IndexRebuilder ports.IndexRebuilder
	Pages          ports.PageRepository
	Journals       ports.JournalRepository
	Habits         ports.HabitRepository
	Persons        ports.PersonRepository
	Bookmarks      ports.BookmarkRepository
	Inbox          ports.InboxRepository
	IDGenerator    ports.IDGenerator
	NoteFinder     ports.NoteFinder
}
