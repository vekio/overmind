package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
	configurfave "github.com/vekio/config/urfave"
	"github.com/vekio/overmind/internal/app"
	appconfig "github.com/vekio/overmind/internal/config"
	"github.com/vekio/overmind/internal/tui"
)

func New(configFile *configlib.ConfigFile[appconfig.Settings], newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:                  "overmind",
		Version:               buildVersion(),
		Usage:                 "manage the overmind knowledge base",
		Description:           "Run without a command to open the terminal interface.",
		EnableShellCompletion: true,
		Flags:                 []urfavecli.Flag{configurfave.NewConfigFlag(configFile)},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			return tui.Run(ctx, func(ctx context.Context, query app.ListNotesQuery) (app.ListNotesResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return app.ListNotesResult{}, err
				}
				return client.ListNotes(ctx, query)
			}, func(ctx context.Context, command app.ReindexCommand) (app.ReindexResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return app.ReindexResult{}, err
				}
				return client.Reindex(ctx, command)
			}, func(ctx context.Context, command app.DeleteNoteCommand) (app.DeleteNoteResult, error) {
				client, err := newClient(ctx)
				if err != nil {
					return app.DeleteNoteResult{}, err
				}
				return client.DeleteNote(ctx, command)
			})
		},
		Commands: []*urfavecli.Command{
			newListCommand(newClient),
			newReindexCommand(newClient),
			newDeleteCommand(newClient),
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
