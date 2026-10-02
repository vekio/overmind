// Package tui provides Overmind's interactive terminal interface.
package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
	"uuid"
)

// Client provides the operations available in the terminal interface.
type Client interface {
	CreateBookmark(context.Context, domain.URL, domain.Tags) (app.CreateBookmarkResult, error)
	Capture(context.Context, string) (app.CaptureResult, error)
	CreateJournal(context.Context, domain.Tags) (app.CreateJournalResult, error)
	CreatePerson(context.Context, domain.Title, domain.Groups, domain.Tags) (app.CreatePersonResult, error)
	CreatePage(context.Context, domain.Title, domain.Area, domain.Tags) (app.CreatePageResult, error)
	RebuildIndex(context.Context) (int, error)
	ListNotes(context.Context) ([]app.ListedNote, error)
	FindJournal(context.Context, domain.Date) (uuid.UUID, bool, error)
	OpenNote(context.Context, uuid.UUID) ([]byte, error)
	UpdateNote(context.Context, uuid.UUID, []byte, []byte) (app.UpdateNoteResult, error)
	DeleteNote(context.Context, uuid.UUID) error
}

// Run opens the terminal interface until the user quits.
func Run(ctx context.Context, client Client) error {
	_, err := tea.NewProgram(newModel(ctx, client), tea.WithContext(ctx)).Run()
	return err
}
