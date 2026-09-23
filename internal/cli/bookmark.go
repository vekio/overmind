package cli

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"git.casta.me/alberto/overmind/internal/id"
	"git.casta.me/alberto/overmind/internal/renderer"
	"git.casta.me/alberto/overmind/internal/storage"
	urfavecli "github.com/urfave/cli/v3"
)

func newBookmarkCommand(documentRenderer *renderer.Renderer, noteWriter *storage.FileWriter) *urfavecli.Command {
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
			&urfavecli.StringFlag{
				Name:   "title",
				Usage:  "use `TITLE` instead of the URL",
				Config: urfavecli.StringConfig{TrimSpace: true},
			},
			tagFlag("add `TAG`; repeatable"),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			urlArguments := command.StringArgs("url")
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after URL: %q", command.Args().Slice())
			}

			bookmarkURL, err := normalizeBookmarkURL(urlArguments[0])
			if err != nil {
				return err
			}
			title := command.String("title")
			if title == "" {
				title = bookmarkURL
			}

			now := time.Now()
			noteID := id.New()
			content, err := documentRenderer.Render(ctx, renderer.BookmarkTemplate, renderer.Bookmark{
				ID:        noteID.String(),
				Title:     title,
				URL:       bookmarkURL,
				Tags:      command.StringSlice("tag"),
				CreatedAt: now,
				UpdatedAt: now,
			})
			if err != nil {
				return err
			}

			path, err := noteWriter.Write(ctx, noteID, content)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(command.Writer, path); err != nil {
				return fmt.Errorf("write created bookmark path: %w", err)
			}
			return nil
		},
	}
}

func normalizeBookmarkURL(value string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("invalid bookmark URL: %w", err)
	}
	if parsed.Host == "" || parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid bookmark URL: expected an HTTP or HTTPS URL")
	}
	return parsed.String(), nil
}
