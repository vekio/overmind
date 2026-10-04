package cli

import (
	"context"

	"github.com/vekio/overmind/internal/app"
)

// Client provides the note operations used by the CLI.
type Client interface {
	DeleteNote(context.Context, app.DeleteNoteCommand) (app.DeleteNoteResult, error)
	Reindex(context.Context, app.ReindexCommand) (app.ReindexResult, error)
	ListNotes(context.Context, app.ListNotesQuery) (app.ListNotesResult, error)
	CreateHabit(context.Context, app.CreateHabitCommand) (app.CreateHabitResult, error)
	CreateBookmark(context.Context, app.CreateBookmarkCommand) (app.CreateBookmarkResult, error)
	CreateInbox(context.Context, app.CreateInboxCommand) (app.CreateInboxResult, error)
	CreateJournal(context.Context, app.CreateJournalCommand) (app.CreateJournalResult, error)
	CreatePerson(context.Context, app.CreatePersonCommand) (app.CreatePersonResult, error)
	CreatePage(context.Context, app.CreatePageCommand) (app.CreatePageResult, error)
}

// ClientFactory creates the client after command flags have been parsed.
type ClientFactory func(context.Context) (Client, error)

// NewLocalClient adapts an application running in this process to the CLI.
func NewLocalClient(application *app.Application) *LocalClient {
	return &LocalClient{application: application}
}

// LocalClient runs CLI operations against the in-process application.
type LocalClient struct {
	application *app.Application
}

func (client *LocalClient) CreateBookmark(ctx context.Context, command app.CreateBookmarkCommand) (app.CreateBookmarkResult, error) {
	return client.application.Commands.CreateBookmark.Handle(ctx, command)
}
func (client *LocalClient) CreateInbox(ctx context.Context, command app.CreateInboxCommand) (app.CreateInboxResult, error) {
	return client.application.Commands.CreateInbox.Handle(ctx, command)
}

func (client *LocalClient) CreateJournal(ctx context.Context, command app.CreateJournalCommand) (app.CreateJournalResult, error) {
	return client.application.Commands.CreateJournal.Handle(ctx, command)
}

func (client *LocalClient) CreatePage(ctx context.Context, command app.CreatePageCommand) (app.CreatePageResult, error) {
	return client.application.Commands.CreatePage.Handle(ctx, command)
}

func (client *LocalClient) CreatePerson(ctx context.Context, command app.CreatePersonCommand) (app.CreatePersonResult, error) {
	return client.application.Commands.CreatePerson.Handle(ctx, command)
}

// CreateHabit forwards raw CLI input to the use case's domain validation.
func (client *LocalClient) CreateHabit(ctx context.Context, command app.CreateHabitCommand) (app.CreateHabitResult, error) {
	return client.application.Commands.CreateHabit.Handle(ctx, command)
}

func (client *LocalClient) ListNotes(ctx context.Context, query app.ListNotesQuery) (app.ListNotesResult, error) {
	return client.application.Queries.ListNotes.Handle(ctx, query)
}

func (client *LocalClient) Reindex(ctx context.Context, command app.ReindexCommand) (app.ReindexResult, error) {
	return client.application.Commands.Reindex.Handle(ctx, command)
}

func (client *LocalClient) DeleteNote(ctx context.Context, command app.DeleteNoteCommand) (app.DeleteNoteResult, error) {
	return client.application.Commands.DeleteNote.Handle(ctx, command)
}
