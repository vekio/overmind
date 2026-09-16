package cli

import urfavecli "github.com/urfave/cli/v3"

func tagFlag(usage string) urfavecli.Flag {
	return &urfavecli.StringSliceFlag{
		Name:    "tag",
		Aliases: []string{"t"},
		Usage:   usage,
	}
}

func documentTypeFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:  "type",
		Usage: "filter by document `TYPE`",
		Config: urfavecli.StringConfig{
			TrimSpace: true,
		},
	}
}
