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
