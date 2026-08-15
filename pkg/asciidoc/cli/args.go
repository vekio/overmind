package cli

import urfavecli "github.com/urfave/cli/v3"

func asciiDocPathArgument() []urfavecli.Argument {
	return []urfavecli.Argument{
		&urfavecli.StringArgs{
			Name:      "path",
			UsageText: "PATH",
			Min:       1,
			Max:       1,
		},
	}
}
