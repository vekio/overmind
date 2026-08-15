package cli

import (
	"context"
	"fmt"

	"git.casta.me/alberto/overmind/pkg/asciidoc"
	"git.casta.me/alberto/overmind/pkg/asciidoc/format"
	urfavecli "github.com/urfave/cli/v3"
)

func newParserCommand() *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "parser",
		Usage:     "print the parsed AsciiDoc result as JSON",
		ArgsUsage: "PATH",
		Arguments: asciiDocPathArgument(),
		Action:    runParser,
	}
}

func runParser(_ context.Context, command *urfavecli.Command) error {
	path, file, err := openAsciiDoc(command, "parser")
	if err != nil {
		return err
	}
	defer file.Close()

	result := asciidoc.ParseReader(file)
	if err := format.WriteJSON(command.Writer, result); err != nil {
		return fmt.Errorf("write parser result for %q: %w", path, err)
	}
	if result.HasErrors() {
		return fmt.Errorf("parse AsciiDoc file %q: completed with errors", path)
	}
	return nil
}
