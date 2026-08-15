package cli

import (
	urfavecli "github.com/urfave/cli/v3"
)

// Command creates the asciidoc command for registration in an application.
func Command() *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "asciidoc",
		Usage: "inspect AsciiDoc source files",
		Commands: []*urfavecli.Command{
			newLexerCommand(),
			newParserCommand(),
		},
	}
}
