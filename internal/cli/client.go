package cli

import (
	"context"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
)

// Client provides the note operations used by the CLI.
type Client interface {
	CreateBookmark(context.Context, domain.URL, domain.Tags) (string, error)
	Capture(context.Context, string) (string, error)
	CreateJournal(context.Context, domain.Tags) (string, error)
	CreatePage(context.Context, domain.Title, domain.Area, domain.Tags) (string, error)
	RebuildIndex(context.Context) (int, error)
	ListNotes(context.Context) ([]app.ListedNote, error)
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

func (client *LocalClient) CreateBookmark(ctx context.Context, url domain.URL, tags domain.Tags) (string, error) {
	result, err := client.application.Commands.CreateBookmark.Handle(ctx, app.CreateBookmarkCommand{URL: url, Tags: tags})
	return result.Path, err
}

func (client *LocalClient) Capture(ctx context.Context, content string) (string, error) {
	result, err := client.application.Commands.Capture.Handle(ctx, app.CaptureCommand{Content: content})
	return result.Path, err
}

func (client *LocalClient) CreateJournal(ctx context.Context, tags domain.Tags) (string, error) {
	result, err := client.application.Commands.CreateJournal.Handle(ctx, app.CreateJournalCommand{Tags: tags})
	return result.Path, err
}

func (client *LocalClient) CreatePage(ctx context.Context, title domain.Title, area domain.Area, tags domain.Tags) (string, error) {
	result, err := client.application.Commands.CreatePage.Handle(ctx, app.CreatePageCommand{Title: title, Area: area, Tags: tags})
	return result.Path, err
}

func (client *LocalClient) RebuildIndex(ctx context.Context) (int, error) {
	result, err := client.application.Commands.RebuildIndex.Handle(ctx)
	return result.Count, err
}

func (client *LocalClient) ListNotes(ctx context.Context) ([]app.ListedNote, error) {
	result, err := client.application.Queries.ListNotes.Handle(ctx)
	return result.Notes, err
}
