package server

import urfavecli "github.com/urfave/cli/v3"

func addressFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:    "address",
		Aliases: []string{"a"},
		Usage:   "HTTP listen address",
		Config:  urfavecli.StringConfig{TrimSpace: true},
		Sources: urfavecli.EnvVars("OVERMIND_SERVER_HTTP_ADDRESS"),
	}
}
