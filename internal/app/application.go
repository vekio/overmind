// Package app exposes the application's use-case handlers to input adapters.
package app

// Application groups the use-case handlers exposed by Overmind.
type Application struct {
	Commands Commands
}

// Commands contains the application's state-changing handlers.
type Commands struct {
	CreatePage     *CreatePageHandler
	CreateJournal  *CreateJournalHandler
	Capture        *CaptureHandler
	CreateBookmark *CreateBookmarkHandler
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
		},
	}
}
