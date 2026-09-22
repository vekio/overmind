package cli

import (
	"context"

	urfavecli "github.com/urfave/cli/v3"
)

func New() (*urfavecli.Command, error) {
	return &urfavecli.Command{
		Name:                  "overmind",
		Usage:                 "manage the overmind knowledge base",
		EnableShellCompletion: true,
		Commands:              []*urfavecli.Command{},
		Action: func(ctx context.Context, c *urfavecli.Command) error {
			_, err := c.Writer.Write([]byte("overmind\n"))
			return err
		},
	}, nil
}
