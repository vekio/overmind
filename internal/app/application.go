// Package app exposes the application's use-case handlers to input adapters.
package app

// Application groups the use-case handlers exposed by Overmind.
type Application struct {
	Commands Commands
	Queries  Queries
}

// Commands contains the application's state-changing handlers.
type Commands struct {
	CreatePage     *CreatePageHandler
	CreateJournal  *CreateJournalHandler
	Capture        *CaptureHandler
	CreateBookmark *CreateBookmarkHandler
	RebuildIndex   *RebuildIndexHandler
	UpdateNote     *UpdateNoteHandler
	DeleteNote     *DeleteNoteHandler
}

// Queries contains the application's read-only handlers.
type Queries struct {
	ListNotes   *ListNotesHandler
	OpenNote    *OpenNoteHandler
	FindJournal *FindJournalHandler
}

// New creates the application handlers.
func New(dependencies Dependencies) *Application {
	saver := newNoteSaver(dependencies.Renderer, dependencies.Writer, dependencies.Index)

	return &Application{
		Commands: Commands{
			CreatePage:     newCreatePageHandler(saver, dependencies.IDGenerator),
			CreateJournal:  newCreateJournalHandler(saver, dependencies.Index, dependencies.IDGenerator),
			Capture:        newCaptureHandler(saver, dependencies.IDGenerator),
			CreateBookmark: newCreateBookmarkHandler(saver, dependencies.IDGenerator),
			RebuildIndex:   newRebuildIndexHandler(dependencies.Walker, dependencies.Parser, dependencies.Index),
			UpdateNote:     newUpdateNoteHandler(dependencies.Reader, dependencies.Writer, dependencies.Parser, dependencies.Index),
			DeleteNote:     newDeleteNoteHandler(dependencies.Deleter, dependencies.Index),
		},
		Queries: Queries{
			ListNotes:   newListNotesHandler(dependencies.Lister),
			OpenNote:    newOpenNoteHandler(dependencies.Reader),
			FindJournal: newFindJournalHandler(dependencies.Index),
		},
	}
}
