// Package app exposes the application's use-case handlers to input adapters.
package app

// Application groups the use-case handlers exposed by Overmind.
type Application struct {
	Commands Commands
	Queries  Queries
}

// Queries contains read-only application operations.
type Queries struct{ ListNotes *ListNotesHandler }

// Commands contains the application's state-changing handlers.
type Commands struct {
	DeleteNote     *DeleteNoteHandler
	Reindex        *ReindexHandler
	CreateHabit    *CreateHabitHandler
	CreatePage     *CreatePageHandler
	CreatePerson   *CreatePersonHandler
	CreateJournal  *CreateJournalHandler
	CreateInbox    *CreateInboxHandler
	CreateBookmark *CreateBookmarkHandler
}

// New creates the application handlers.
func New(dependencies Dependencies) *Application {
	return &Application{
		Queries: Queries{ListNotes: newListNotesHandler(dependencies.NoteFinder)},
		Commands: Commands{
			DeleteNote:     newDeleteNoteHandler(dependencies.NoteStore, dependencies.IndexDeleter),
			Reindex:        newReindexHandler(dependencies.NoteScanner, dependencies.NoteProjector, dependencies.IndexRebuilder),
			CreateHabit:    newCreateHabitHandler(dependencies.Habits, dependencies.IDGenerator),
			CreatePage:     newCreatePageHandler(dependencies.Pages, dependencies.IDGenerator),
			CreatePerson:   newCreatePersonHandler(dependencies.Persons, dependencies.IDGenerator),
			CreateJournal:  newCreateJournalHandler(dependencies.Journals, dependencies.IDGenerator),
			CreateInbox:    newCreateInboxHandler(dependencies.Inbox, dependencies.IDGenerator),
			CreateBookmark: newCreateBookmarkHandler(dependencies.Bookmarks, dependencies.IDGenerator),
		},
	}
}
