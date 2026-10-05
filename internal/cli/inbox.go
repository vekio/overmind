package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app/inbox"
)

func newInboxCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "inbox",
		Flags:     []urfavecli.Flag{tagFlag()},
		Usage:     "create a quick inbox note from text or standard input",
		ArgsUsage: "[TEXT]",
		Arguments: inboxTextArgument(),
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			textArguments := command.StringArgs("text")
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after text: %q", command.Args().Slice())
			}
			text, err := inboxText(command.Reader, textArguments)
			if err != nil {
				return err
			}

			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			result, err := client.CreateInbox(ctx, inbox.CreateCommand{Content: text, Tags: command.StringSlice("tag")})
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(command.Writer, "Inbox created\nID: %s\n", result.Inbox.ID()); err != nil {
				return fmt.Errorf("write inbox summary: %w", err)
			}
			return nil
		},
	}
}

func inboxText(reader io.Reader, arguments []string) (string, error) {
	if len(arguments) != 0 {
		if strings.TrimSpace(arguments[0]) == "" {
			return "", fmt.Errorf("inbox text must not be empty")
		}
		return arguments[0], nil
	}

	if reader == nil {
		return "", fmt.Errorf("inbox text is required as an argument or standard input")
	}
	if input, ok := reader.(*os.File); ok {
		info, err := input.Stat()
		if err != nil {
			return "", fmt.Errorf("inspect standard input: %w", err)
		}
		if info.Mode()&os.ModeCharDevice != 0 {
			return "", fmt.Errorf("inbox text is required as an argument or standard input")
		}
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read standard input: %w", err)
	}
	text := strings.TrimRight(string(content), "\r\n")
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("inbox text is required as an argument or standard input")
	}
	return text, nil
}
