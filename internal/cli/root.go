// Package cli defines the Overmind command-line application.
package cli

import (
	"git.casta.me/alberto/overmind/internal/config"
	asciidoccli "git.casta.me/alberto/overmind/pkg/asciidoc/cli"
	urfavecli "github.com/urfave/cli/v3"
	configurfave "github.com/vekio/config/urfave"
)

// New creates the complete Overmind command tree.
func New() (*urfavecli.Command, error) {
	configFile, defaults, err := config.NewFile()
	if err != nil {
		return nil, err
	}
	return &urfavecli.Command{
		Name:                  "overmind",
		Usage:                 "manage the overmind knowledge base",
		EnableShellCompletion: true,
		Flags: []urfavecli.Flag{
			configFlag(configFile),
		},
		Commands: []*urfavecli.Command{
			newSetupCommand(configFile, defaults),
			configurfave.NewConfigCommand(configFile, defaults),
			asciidoccli.Command(),
			newIndexCommand(configFile),
			newListCommand(configFile),
			newPageCommand(configFile),
		},
	}, nil
}
