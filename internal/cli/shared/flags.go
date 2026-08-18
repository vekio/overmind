package shared

import urfavecli "github.com/urfave/cli/v3"

// DebugFlag creates the common diagnostic logging flag.
func DebugFlag(environmentName string) urfavecli.Flag {
	return &urfavecli.BoolFlag{
		Name:    "debug",
		Usage:   "show diagnostic logs",
		Sources: urfavecli.EnvVars(environmentName),
	}
}
