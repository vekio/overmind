package cli

import urfavecli "github.com/urfave/cli/v3"

func titleArgument() []urfavecli.Argument {
	return stringArgument("title", "TITLE", 1)
}

func nameArgument() []urfavecli.Argument {
	return stringArgument("name", "NAME", 1)
}

func urlArgument() []urfavecli.Argument {
	return stringArgument("url", "URL", 1)
}

func noteIDArgument() []urfavecli.Argument {
	return stringArgument("id", "ID", 1)
}

func inboxTextArgument() []urfavecli.Argument {
	return stringArgument("text", "[TEXT]", 0)
}

// stringArgument returns fresh positional definitions so commands do not share mutable parser state.
func stringArgument(name, usage string, minimum int) []urfavecli.Argument {
	return []urfavecli.Argument{
		&urfavecli.StringArgs{
			Name:      name,
			UsageText: usage,
			Min:       minimum,
			Max:       1,
		},
	}
}
