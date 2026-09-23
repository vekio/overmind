package cli

import (
	"context"
	"fmt"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/id"
	"git.casta.me/alberto/overmind/internal/infra/sqliteindex"
	"git.casta.me/alberto/overmind/internal/renderer"
	"git.casta.me/alberto/overmind/internal/storage"
	urfavecli "github.com/urfave/cli/v3"
)

func newBookmarkCommand(
	documentRenderer *renderer.Renderer,
	noteWriter *storage.FileWriter,
	noteIndex *sqliteindex.Store,
) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "bookmark",
		Usage:     "save a link for later",
		ArgsUsage: "URL",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{
				Name:      "url",
				UsageText: "URL",
				Min:       1,
				Max:       1,
			},
		},
		Flags: []urfavecli.Flag{
			tagFlag("add `TAG`; repeatable"),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			urlArguments := command.StringArgs("url")
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after URL: %q", command.Args().Slice())
			}

			bookmarkURL, err := domain.NewURL(urlArguments[0])
			if err != nil {
				return err
			}
			tags, err := parseTags(command.StringSlice("tag"))
			if err != nil {
				return err
			}

			now := time.Now()
			noteID := id.New()
			bookmark, err := domain.NewBookmark(noteID, bookmarkURL, tags, now)
			if err != nil {
				return err
			}

			content, err := documentRenderer.Render(ctx, renderer.BookmarkTemplate, bookmark)
			if err != nil {
				return err
			}

			path, err := noteWriter.Write(ctx, noteID, content)
			if err != nil {
				return err
			}
			if err := noteIndex.Upsert(ctx, bookmark); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(command.Writer, path); err != nil {
				return fmt.Errorf("write created bookmark path: %w", err)
			}
			return nil
		},
	}
}
