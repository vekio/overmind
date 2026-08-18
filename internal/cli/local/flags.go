package local

import urfavecli "github.com/urfave/cli/v3"

func tagFlag() urfavecli.Flag {
	return &urfavecli.StringSliceFlag{
		Name:    "tag",
		Aliases: []string{"t"},
		Usage:   "document tag; repeat to specify multiple tags",
	}
}

func documentTypeFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:  "type",
		Usage: "document type",
		Config: urfavecli.StringConfig{
			TrimSpace: true,
		},
	}
}
