package cli

import (
	urfavecli "github.com/urfave/cli/v3"
)

func areaFlag(usage string) urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:    "area",
		Aliases: []string{"a"},
		Usage:   usage,
		Config:  urfavecli.StringConfig{TrimSpace: true},
	}
}

func tagFlag(usage string) urfavecli.Flag {
	return &urfavecli.StringSliceFlag{
		Name:    "tag",
		Aliases: []string{"t"},
		Usage:   usage,
	}
}
