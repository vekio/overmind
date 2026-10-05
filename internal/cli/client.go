package cli

import (
	"context"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/app/bookmark"
	"github.com/vekio/overmind/internal/app/habit"
	"github.com/vekio/overmind/internal/app/inbox"
	"github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/app/notes"
	"github.com/vekio/overmind/internal/app/page"
	"github.com/vekio/overmind/internal/app/person"
	"github.com/vekio/overmind/internal/app/rawedit"
)

// Client provides the note operations used by the CLI.
type Client interface {
	// GetRawNote loads managed source without requiring a valid body.
	GetRawNote(context.Context, rawedit.GetQuery) (rawedit.GetResult, error)
	// UpdateRawNote validates and saves source, reporting partial index failures.
	UpdateRawNote(context.Context, rawedit.UpdateCommand) (rawedit.UpdateResult, error)
	// GetPage loads and verifies the requested page document.
	GetPage(context.Context, page.GetQuery) (page.GetResult, error)
	// UpdatePage replaces editable page values while preserving identity and creation time.
	UpdatePage(context.Context, page.UpdateCommand) (page.UpdateResult, error)
	// GetInbox loads and verifies the requested inbox document.
	GetInbox(context.Context, inbox.GetQuery) (inbox.GetResult, error)
	// UpdateInbox replaces editable inbox values while preserving identity and creation time.
	UpdateInbox(context.Context, inbox.UpdateCommand) (inbox.UpdateResult, error)
	// GetBookmark loads and verifies the requested bookmark document.
	GetBookmark(context.Context, bookmark.GetQuery) (bookmark.GetResult, error)
	// UpdateBookmark replaces editable bookmark values while preserving identity and creation time.
	UpdateBookmark(context.Context, bookmark.UpdateCommand) (bookmark.UpdateResult, error)
	// GetPerson loads and verifies the requested person document.
	GetPerson(context.Context, person.GetQuery) (person.GetResult, error)
	// UpdatePerson replaces editable person values while preserving identity and creation time.
	UpdatePerson(context.Context, person.UpdateCommand) (person.UpdateResult, error)
	// GetHabit loads and verifies the requested habit document.
	GetHabit(context.Context, habit.GetQuery) (habit.GetResult, error)
	// UpdateHabit replaces editable habit values while preserving identity and creation time.
	UpdateHabit(context.Context, habit.UpdateCommand) (habit.UpdateResult, error)
	// ListGroups lists normalized group names assigned to people.
	ListGroups(context.Context, person.ListGroupsQuery) (person.ListGroupsResult, error)
	// DeleteNote removes source before its derived projection.
	DeleteNote(context.Context, notes.DeleteCommand) (notes.DeleteResult, error)
	// Reindex rebuilds the complete index from vault documents.
	Reindex(context.Context, notes.ReindexCommand) (notes.ReindexResult, error)
	// ListNotes queries filtered and paginated index summaries.
	ListNotes(context.Context, notes.ListQuery) (notes.ListResult, error)
	// CreateHabit validates and persists a new habit.
	CreateHabit(context.Context, habit.CreateCommand) (habit.CreateResult, error)
	// CreateBookmark validates and persists a new bookmark.
	CreateBookmark(context.Context, bookmark.CreateCommand) (bookmark.CreateResult, error)
	// CreateInbox validates and persists a new inbox.
	CreateInbox(context.Context, inbox.CreateCommand) (inbox.CreateResult, error)
	// CreateJournal validates and persists a new journal.
	CreateJournal(context.Context, journal.CreateCommand) (journal.CreateResult, error)
	// GetJournal loads and verifies the requested journal document.
	GetJournal(context.Context, journal.GetQuery) (journal.GetResult, error)
	// UpdateJournal replaces editable journal values while preserving identity and creation time.
	UpdateJournal(context.Context, journal.UpdateCommand) (journal.UpdateResult, error)
	// CreatePerson validates and persists a new person.
	CreatePerson(context.Context, person.CreateCommand) (person.CreateResult, error)
	// CreatePage validates and persists a new page.
	CreatePage(context.Context, page.CreateCommand) (page.CreateResult, error)
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

// CreateBookmark delegates to the corresponding application use case.
func (client *LocalClient) CreateBookmark(ctx context.Context, command bookmark.CreateCommand) (bookmark.CreateResult, error) {
	return client.application.Commands.CreateBookmark.Handle(ctx, command)
}

// CreateInbox delegates to the corresponding application use case.
func (client *LocalClient) CreateInbox(ctx context.Context, command inbox.CreateCommand) (inbox.CreateResult, error) {
	return client.application.Commands.CreateInbox.Handle(ctx, command)
}

// CreateJournal delegates to the corresponding application use case.
func (client *LocalClient) CreateJournal(ctx context.Context, command journal.CreateCommand) (journal.CreateResult, error) {
	return client.application.Commands.CreateJournal.Handle(ctx, command)
}

// GetJournal delegates to the corresponding application use case.
func (client *LocalClient) GetJournal(ctx context.Context, query journal.GetQuery) (journal.GetResult, error) {
	return client.application.Queries.GetJournal.Handle(ctx, query)
}

// UpdateJournal delegates to the corresponding application use case.
func (client *LocalClient) UpdateJournal(ctx context.Context, command journal.UpdateCommand) (journal.UpdateResult, error) {
	return client.application.Commands.UpdateJournal.Handle(ctx, command)
}

// CreatePage delegates to the corresponding application use case.
func (client *LocalClient) CreatePage(ctx context.Context, command page.CreateCommand) (page.CreateResult, error) {
	return client.application.Commands.CreatePage.Handle(ctx, command)
}

// CreatePerson delegates to the corresponding application use case.
func (client *LocalClient) CreatePerson(ctx context.Context, command person.CreateCommand) (person.CreateResult, error) {
	return client.application.Commands.CreatePerson.Handle(ctx, command)
}

// CreateHabit forwards raw CLI input to the use case's domain validation.
func (client *LocalClient) CreateHabit(ctx context.Context, command habit.CreateCommand) (habit.CreateResult, error) {
	return client.application.Commands.CreateHabit.Handle(ctx, command)
}

// ListNotes delegates to the corresponding application use case.
func (client *LocalClient) ListNotes(ctx context.Context, query notes.ListQuery) (notes.ListResult, error) {
	return client.application.Queries.ListNotes.Handle(ctx, query)
}

// Reindex delegates to the corresponding application use case.
func (client *LocalClient) Reindex(ctx context.Context, command notes.ReindexCommand) (notes.ReindexResult, error) {
	return client.application.Commands.Reindex.Handle(ctx, command)
}

// DeleteNote delegates to the corresponding application use case.
func (client *LocalClient) DeleteNote(ctx context.Context, command notes.DeleteCommand) (notes.DeleteResult, error) {
	return client.application.Commands.DeleteNote.Handle(ctx, command)
}

// ListGroups delegates to the corresponding application use case.
func (client *LocalClient) ListGroups(ctx context.Context, query person.ListGroupsQuery) (person.ListGroupsResult, error) {
	return client.application.Queries.ListGroups.Handle(ctx, query)
}

// GetHabit delegates to the corresponding application use case.
func (client *LocalClient) GetHabit(ctx context.Context, query habit.GetQuery) (habit.GetResult, error) {
	return client.application.Queries.GetHabit.Handle(ctx, query)
}

// UpdateHabit delegates to the corresponding application use case.
func (client *LocalClient) UpdateHabit(ctx context.Context, command habit.UpdateCommand) (habit.UpdateResult, error) {
	return client.application.Commands.UpdateHabit.Handle(ctx, command)
}

// GetPerson delegates to the corresponding application use case.
func (client *LocalClient) GetPerson(ctx context.Context, query person.GetQuery) (person.GetResult, error) {
	return client.application.Queries.GetPerson.Handle(ctx, query)
}

// UpdatePerson delegates to the corresponding application use case.
func (client *LocalClient) UpdatePerson(ctx context.Context, command person.UpdateCommand) (person.UpdateResult, error) {
	return client.application.Commands.UpdatePerson.Handle(ctx, command)
}

// GetBookmark delegates to the corresponding application use case.
func (client *LocalClient) GetBookmark(ctx context.Context, query bookmark.GetQuery) (bookmark.GetResult, error) {
	return client.application.Queries.GetBookmark.Handle(ctx, query)
}

// UpdateBookmark delegates to the corresponding application use case.
func (client *LocalClient) UpdateBookmark(ctx context.Context, command bookmark.UpdateCommand) (bookmark.UpdateResult, error) {
	return client.application.Commands.UpdateBookmark.Handle(ctx, command)
}

// GetInbox delegates to the corresponding application use case.
func (client *LocalClient) GetInbox(ctx context.Context, query inbox.GetQuery) (inbox.GetResult, error) {
	return client.application.Queries.GetInbox.Handle(ctx, query)
}

// UpdateInbox delegates to the corresponding application use case.
func (client *LocalClient) UpdateInbox(ctx context.Context, command inbox.UpdateCommand) (inbox.UpdateResult, error) {
	return client.application.Commands.UpdateInbox.Handle(ctx, command)
}

// GetPage delegates to the corresponding application use case.
func (client *LocalClient) GetPage(ctx context.Context, query page.GetQuery) (page.GetResult, error) {
	return client.application.Queries.GetPage.Handle(ctx, query)
}

// UpdatePage delegates to the corresponding application use case.
func (client *LocalClient) UpdatePage(ctx context.Context, command page.UpdateCommand) (page.UpdateResult, error) {
	return client.application.Commands.UpdatePage.Handle(ctx, command)
}

// GetRawNote delegates to the corresponding application use case.
func (client *LocalClient) GetRawNote(ctx context.Context, query rawedit.GetQuery) (rawedit.GetResult, error) {
	return client.application.Queries.GetRawNote.Handle(ctx, query)
}

// UpdateRawNote delegates to the corresponding application use case.
func (client *LocalClient) UpdateRawNote(ctx context.Context, command rawedit.UpdateCommand) (rawedit.UpdateResult, error) {
	return client.application.Commands.UpdateRawNote.Handle(ctx, command)
}
