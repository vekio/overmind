package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app"
)

func newHabitCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:        "habit",
		Usage:       "create a habit with a fixed quantity per calendar period",
		Description: `Example: overmind habit --amount 2 --unit litros --period day --tag salud "Beber agua"`,
		ArgsUsage:   "TITLE",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{Name: "title", UsageText: "TITLE", Min: 1, Max: 1},
		},
		Flags: []urfavecli.Flag{
			&urfavecli.Float64Flag{Name: "amount", Usage: "positive target `AMOUNT`; decimals allowed", Required: true},
			&urfavecli.StringFlag{Name: "unit", Usage: "target `UNIT`, such as litros or entrenamientos", Required: true},
			&urfavecli.StringFlag{Name: "period", Usage: "calendar `PERIOD`: day, week or month", Required: true},
			tagFlag("add `TAG`; repeatable and optional"),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after title: %q; quote a title containing spaces", command.Args().Slice())
			}
			input := app.CreateHabitCommand{
				Title:  command.StringArgs("title")[0],
				Amount: command.Float64("amount"),
				Unit:   command.String("unit"),
				Period: command.String("period"),
				Tags:   command.StringSlice("tag"),
			}
			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.CreateHabit(ctx, input)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(command.Writer, "%s\nID: %s\n", result.Habit.Summary(), result.Habit.ID()); err != nil {
				return fmt.Errorf("write created habit summary: %w", err)
			}
			return nil
		},
	}
}
