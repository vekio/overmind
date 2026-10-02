package cli

import (
	"context"
	"fmt"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/pkg/asciidoc/lexer"
)

func newLexerCommand() *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "lexer",
		Usage:     "print the line tokens produced for an AsciiDoc file",
		ArgsUsage: "PATH",
		Arguments: asciiDocPathArgument(),
		Action:    runLexer,
	}
}

func runLexer(_ context.Context, command *urfavecli.Command) error {
	path, file, err := openAsciiDoc(command, "lexer")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := lexer.New(file)
	for {
		token, err := scanner.Next()
		if err != nil {
			return fmt.Errorf("lex AsciiDoc file %q: %w", path, err)
		}
		if token.Kind == lexer.LineEOF {
			return nil
		}
		if _, err := fmt.Fprintln(command.Writer, token); err != nil {
			return fmt.Errorf("write lexer token: %w", err)
		}
	}
}
