// Package app exposes the application's use-case handlers to input adapters.
package app

import (
	"github.com/vekio/overmind/internal/app/bookmark"
	"github.com/vekio/overmind/internal/app/habit"
	"github.com/vekio/overmind/internal/app/inbox"
	"github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/app/notes"
	"github.com/vekio/overmind/internal/app/page"
	"github.com/vekio/overmind/internal/app/person"
	"github.com/vekio/overmind/internal/app/rawedit"
)

// Application groups the use-case handlers exposed by Overmind.
type Application struct {
	Commands Commands
	Queries  Queries
}

// Queries contains read-only application operations.
type Queries struct {
	GetRawNote  *rawedit.GetHandler
	GetPage     *page.GetHandler
	GetInbox    *inbox.GetHandler
	GetBookmark *bookmark.GetHandler
	GetPerson   *person.GetHandler
	GetHabit    *habit.GetHandler
	ListNotes   *notes.ListHandler
	GetJournal  *journal.GetHandler
	ListGroups  *person.ListGroupsHandler
}

// Commands contains the application's state-changing handlers.
type Commands struct {
	UpdateRawNote  *rawedit.UpdateHandler
	UpdatePage     *page.UpdateHandler
	UpdateInbox    *inbox.UpdateHandler
	UpdateBookmark *bookmark.UpdateHandler
	UpdatePerson   *person.UpdateHandler
	UpdateHabit    *habit.UpdateHandler
	DeleteNote     *notes.DeleteHandler
	Reindex        *notes.ReindexHandler
	CreateHabit    *habit.CreateHandler
	CreatePage     *page.CreateHandler
	CreatePerson   *person.CreateHandler
	CreateJournal  *journal.CreateHandler
	UpdateJournal  *journal.UpdateHandler
	CreateInbox    *inbox.CreateHandler
	CreateBookmark *bookmark.CreateHandler
}

// New creates the application handlers.
func New(dependencies Dependencies) *Application {
	return &Application{
		Queries: Queries{
			GetRawNote:  rawedit.NewGetHandler(dependencies.NoteStore, dependencies.RawNoteCodec),
			GetPage:     page.NewGetHandler(dependencies.Pages),
			GetInbox:    inbox.NewGetHandler(dependencies.Inbox),
			GetBookmark: bookmark.NewGetHandler(dependencies.Bookmarks),
			GetPerson:   person.NewGetHandler(dependencies.Persons),
			GetHabit:    habit.NewGetHandler(dependencies.Habits),
			ListNotes:   notes.NewListHandler(dependencies.Index),
			GetJournal:  journal.NewGetHandler(dependencies.Journals),
			ListGroups:  person.NewListGroupsHandler(dependencies.Index),
		},
		Commands: Commands{
			UpdateRawNote: rawedit.NewUpdateHandler(
				dependencies.NoteStore,
				dependencies.RawNoteCodec,
				dependencies.Index,
			),
			UpdatePage:     page.NewUpdateHandler(dependencies.Pages),
			UpdateInbox:    inbox.NewUpdateHandler(dependencies.Inbox),
			UpdateBookmark: bookmark.NewUpdateHandler(dependencies.Bookmarks),
			UpdatePerson:   person.NewUpdateHandler(dependencies.Persons),
			UpdateHabit:    habit.NewUpdateHandler(dependencies.Habits),
			DeleteNote:     notes.NewDeleteHandler(dependencies.NoteStore, dependencies.Index),
			Reindex:        notes.NewReindexHandler(dependencies.NoteStore, dependencies.Index),
			CreateHabit:    habit.NewCreateHandler(dependencies.Habits, dependencies.IDGenerator),
			CreatePage:     page.NewCreateHandler(dependencies.Pages, dependencies.IDGenerator),
			CreatePerson:   person.NewCreateHandler(dependencies.Persons, dependencies.IDGenerator),
			CreateJournal:  journal.NewCreateHandler(dependencies.Journals, dependencies.IDGenerator),
			UpdateJournal:  journal.NewUpdateHandler(dependencies.Journals),
			CreateInbox:    inbox.NewCreateHandler(dependencies.Inbox, dependencies.IDGenerator),
			CreateBookmark: bookmark.NewCreateHandler(dependencies.Bookmarks, dependencies.IDGenerator),
		},
	}
}
