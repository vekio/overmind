package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
	configurfave "github.com/vekio/config/urfave"
	"github.com/vekio/overmind/internal/app/bookmark"
	"github.com/vekio/overmind/internal/app/habit"
	"github.com/vekio/overmind/internal/app/inbox"
	"github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/app/notes"
	"github.com/vekio/overmind/internal/app/page"
	"github.com/vekio/overmind/internal/app/person"
	appconfig "github.com/vekio/overmind/internal/config"
	"github.com/vekio/overmind/internal/tui"
)

// New assembles CLI commands and opens the TUI when invoked without a subcommand.
// Client creation is deferred until an application operation is requested.
func New(configFile *configlib.ConfigFile[appconfig.Settings], newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:                  "overmind",
		Version:               buildVersion(),
		Usage:                 "manage the overmind knowledge base",
		Description:           "Run without a command to open the terminal interface.",
		EnableShellCompletion: true,
		Flags:                 []urfavecli.Flag{configFlag(configFile)},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			return tui.Run(ctx, func(ctx context.Context, query notes.ListQuery) (notes.ListResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return notes.ListResult{}, err
				}
				return client.ListNotes(ctx, query)
			}, func(ctx context.Context, command notes.ReindexCommand) (notes.ReindexResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return notes.ReindexResult{}, err
				}
				return client.Reindex(ctx, command)
			}, func(ctx context.Context, command notes.DeleteCommand) (notes.DeleteResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return notes.DeleteResult{}, err
				}
				return client.DeleteNote(ctx, command)
			}, func(ctx context.Context, command inbox.CreateCommand) (inbox.CreateResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return inbox.CreateResult{}, err
				}
				return client.CreateInbox(ctx, command)
			}, func(ctx context.Context, command page.CreateCommand) (page.CreateResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return page.CreateResult{}, err
				}
				return client.CreatePage(ctx, command)
			}, func(ctx context.Context, command bookmark.CreateCommand) (bookmark.CreateResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return bookmark.CreateResult{}, err
				}
				return client.CreateBookmark(ctx, command)
			}, func(ctx context.Context, command journal.CreateCommand) (journal.CreateResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return journal.CreateResult{}, err
				}
				return client.CreateJournal(ctx, command)
			}, func(ctx context.Context, query journal.GetQuery) (journal.GetResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return journal.GetResult{}, err
				}
				return client.GetJournal(ctx, query)
			}, func(ctx context.Context, command journal.UpdateCommand) (journal.UpdateResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return journal.UpdateResult{}, err
				}
				return client.UpdateJournal(ctx, command)
			}, func(ctx context.Context, command habit.CreateCommand) (habit.CreateResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return habit.CreateResult{}, err
				}
				return client.CreateHabit(ctx, command)
			}, func(ctx context.Context, command person.CreateCommand) (person.CreateResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return person.CreateResult{}, err
				}
				return client.CreatePerson(ctx, command)
			}, func(ctx context.Context, query person.ListGroupsQuery) (person.ListGroupsResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return person.ListGroupsResult{}, err
				}
				return client.ListGroups(ctx, query)
			}, func(ctx context.Context) (tui.EditClient, error) { return newClient(ctx) },
				func(ctx context.Context) (tui.RawEditClient, error) { return newClient(ctx) })
		},
		Commands: []*urfavecli.Command{
			newListCommand(newClient),
			newReindexCommand(newClient),
			newDeleteCommand(newClient),
			newRawEditCommand(newClient),
			newBookmarkCommand(newClient),
			newInboxCommand(newClient),
			newJournalCommand(newClient),
			newHabitCommand(newClient),
			newPageCommand(newClient),
			newPersonCommand(newClient),
			configurfave.NewConfigCommand(configFile),
			NewSetupCommand(configFile),
		},
	}
}
