package cli

import (
	"context"
	"errors"
	"fmt"

	"git.casta.me/alberto/overmind/internal/app/createpage"
	"git.casta.me/alberto/overmind/internal/config"
	"git.casta.me/alberto/overmind/internal/domain"
	urfavecli "github.com/urfave/cli/v3"
)

func newPageCommand(state *applicationState) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "page",
		Usage:     "create an AsciiDoc page",
		ArgsUsage: "TITLE",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{
				Name:      "title",
				UsageText: "TITLE",
				Min:       0,
				Max:       1,
			},
		},
		Flags: []urfavecli.Flag{areaFlag(), tagFlag()},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			titleArguments := command.StringArgs("title")
			if len(titleArguments) == 0 {
				return pageUsageError("missing TITLE")
			}
			if command.NArg() != 0 {
				return pageUsageError(fmt.Sprintf(
					"unexpected arguments after TITLE: %q; quote titles containing spaces",
					command.Args().Slice(),
				))
			}
			application, err := state.get()
			if err != nil {
				return err
			}

			result, err := application.Commands.CreatePage.Handle(ctx, createpage.CreatePageCommand{
				Title: titleArguments[0],
				Area:  command.String("area"),
				Tags:  command.StringSlice("tag"),
			})
			if err != nil {
				if state.config.CLI.Mode == config.CLIModeRemote {
					state.runtime.Logger().DebugContext(ctx, "remote create page command failed", "error", err)
				}
				return presentCreatePageError(err)
			}

			if _, err := fmt.Fprintln(command.Writer, result.Path); err != nil {
				return fmt.Errorf("write created page path: %w", err)
			}
			return nil
		},
	}
}

func presentCreatePageError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidTitle):
		return errors.New("invalid page title")
	case errors.Is(err, domain.ErrInvalidArea):
		return errors.New("invalid page area")
	case errors.Is(err, domain.ErrInvalidTag):
		return errors.New("invalid page tag")
	case errors.Is(err, domain.ErrDuplicateTag):
		return errors.New("duplicate page tag")
	case errors.Is(err, context.Canceled):
		return errors.New("page creation canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return errors.New("page creation timed out")
	}

	if _, ok := errors.AsType[*createpage.PageAlreadyExistsError](err); ok {
		return fmt.Errorf("cannot create page; choose a different title or area: %w", err)
	}
	if incomplete, ok := errors.AsType[*createpage.PageCreationIncompleteError](err); ok {
		return fmt.Errorf(
			"page creation did not finish cleanly at %q; check the document and run 'overmind index rebuild': %w",
			incomplete.Path,
			err,
		)
	}
	return errors.New("could not create page because of an internal error; rerun with --debug for details")
}

func pageUsageError(problem string) error {
	return fmt.Errorf("%s\nRun 'overmind page --help' to see usage and options", problem)
}

func areaFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:    "area",
		Aliases: []string{"a"},
		Usage:   "page area",
		Config:  urfavecli.StringConfig{TrimSpace: true},
	}
}
