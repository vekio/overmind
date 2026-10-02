package cli

import (
	"context"

	"uuid"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/domain"
)

// Client provides the note operations used by the CLI.
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

func (client *LocalClient) CreateBookmark(ctx context.Context, url domain.URL, tags domain.Tags) (app.CreateBookmarkResult, error) {
	return client.application.Commands.CreateBookmark.Handle(ctx, app.CreateBookmarkCommand{URL: url, Tags: tags})
}

func (client *LocalClient) Capture(ctx context.Context, content string) (app.CaptureResult, error) {
	return client.application.Commands.Capture.Handle(ctx, app.CaptureCommand{Content: content})
}

func (client *LocalClient) CreateJournal(ctx context.Context, tags domain.Tags) (app.CreateJournalResult, error) {
	return client.application.Commands.CreateJournal.Handle(ctx, app.CreateJournalCommand{Tags: tags})
}

func (client *LocalClient) CreatePage(ctx context.Context, title domain.Title, area domain.Area, tags domain.Tags) (app.CreatePageResult, error) {
	return client.application.Commands.CreatePage.Handle(ctx, app.CreatePageCommand{Title: title, Area: area, Tags: tags})
}

func (client *LocalClient) RebuildIndex(ctx context.Context) (int, error) {
	result, err := client.application.Commands.RebuildIndex.Handle(ctx)
	return result.Count, err
}

func (client *LocalClient) ListNotes(ctx context.Context) ([]app.ListedNote, error) {
	result, err := client.application.Queries.ListNotes.Handle(ctx)
	return result.Notes, err
}

func (client *LocalClient) FindJournal(ctx context.Context, date domain.Date) (uuid.UUID, bool, error) {
	return client.application.Queries.FindJournal.Handle(ctx, date)
}

func (client *LocalClient) OpenNote(ctx context.Context, id uuid.UUID) ([]byte, error) {
	return client.application.Queries.OpenNote.Handle(ctx, id)
}

func (client *LocalClient) UpdateNote(ctx context.Context, id uuid.UUID, original, source []byte) (app.UpdateNoteResult, error) {
	return client.application.Commands.UpdateNote.Handle(ctx, app.UpdateNoteCommand{ID: id, Original: original, Source: source})
}

func (client *LocalClient) DeleteNote(ctx context.Context, id uuid.UUID) error {
	return client.application.Commands.DeleteNote.Handle(ctx, id)
}

func (client *LocalClient) CreatePerson(ctx context.Context, name domain.Title, groups domain.Groups, tags domain.Tags) (app.CreatePersonResult, error) {
	return client.application.Commands.CreatePerson.Handle(ctx, app.CreatePersonCommand{Name: name, Groups: groups, Tags: tags})
}
