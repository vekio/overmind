package cli

import (
	urfavecli "github.com/urfave/cli/v3"
	configlib "github.com/vekio/config"
	configurfave "github.com/vekio/config/urfave"
	appconfig "github.com/vekio/overmind/internal/config"
)

func configFlag(configFile *configlib.ConfigFile[appconfig.Settings]) urfavecli.Flag {
	return configurfave.NewConfigFlag(configFile)
}

func areaFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:    "area",
		Aliases: []string{"a"},
		Usage:   "organize under `AREA`",
		Config:  urfavecli.StringConfig{TrimSpace: true},
	}
}

// tagFlag returns a fresh repeatable flag; normalization and duplicate checks belong to the use case.
func tagFlag() urfavecli.Flag {
	return &urfavecli.StringSliceFlag{
		Name:    "tag",
		Aliases: []string{"t"},
		Usage:   "add `TAG`; repeatable and optional",
	}
}

func groupFlag() urfavecli.Flag {
	return &urfavecli.StringSliceFlag{
		Name:  "group",
		Usage: "add `GROUP`; repeatable",
	}
}

func dateFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:  "date",
		Usage: "calendar date YYYY-MM-DD; defaults to today",
	}
}

func amountFlag() urfavecli.Flag {
	return &urfavecli.Float64Flag{
		Name:     "amount",
		Usage:    "positive target `AMOUNT`; decimals allowed",
		Required: true,
	}
}

func unitFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:     "unit",
		Usage:    "target `UNIT`, such as litros or entrenamientos",
		Required: true,
	}
}

func periodFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:     "period",
		Usage:    "calendar `PERIOD`: day, week or month",
		Required: true,
	}
}

func typeFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:  "type",
		Usage: "note `TYPE`: habit, person, bookmark, inbox, page or journal",
	}
}

func tagFilterFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:  "tag",
		Usage: "match normalized `TAG`",
	}
}

func limitFlag() urfavecli.Flag {
	return &urfavecli.IntFlag{
		Name:  "limit",
		Value: 100,
		Usage: "maximum number of notes (1–1000; 0 defaults to 100)",
	}
}

func offsetFlag() urfavecli.Flag {
	return &urfavecli.IntFlag{
		Name:  "offset",
		Usage: "skip this many notes",
	}
}
