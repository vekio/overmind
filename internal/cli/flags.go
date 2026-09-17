package cli

import (
	"git.casta.me/alberto/overmind/internal/config"
	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
	configurfave "github.com/vekio/config/urfave"
)

func configFlag(configFile *configlib.ConfigFile[config.Config]) urfavecli.Flag {
	return configurfave.NewConfigFlag(configFile)
}

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

func titleFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:   "title",
		Usage:  "filter by title text",
		Config: urfavecli.StringConfig{TrimSpace: true},
	}
}

func areaFlag(usage string) urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:    "area",
		Aliases: []string{"a"},
		Usage:   usage,
		Config:  urfavecli.StringConfig{TrimSpace: true},
	}
}
