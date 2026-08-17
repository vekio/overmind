package cli

import urfavecli "github.com/urfave/cli/v3"

func configFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:    "config",
		Usage:   "config file path",
		Config:  urfavecli.StringConfig{TrimSpace: true},
		Sources: urfavecli.EnvVars("OVERMIND_CONFIG"),
	}
}

func debugFlag() urfavecli.Flag {
	return &urfavecli.BoolFlag{
		Name:    "debug",
		Usage:   "show diagnostic logs",
		Sources: urfavecli.EnvVars("OVERMIND_DEBUG"),
	}
}

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
