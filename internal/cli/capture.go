package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"git.casta.me/alberto/overmind/internal/app"
	urfavecli "github.com/urfave/cli/v3"
)

func newCaptureCommand(application *app.App) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "capture",
		Usage:     "capture a quick inbox note from text or standard input",
		ArgsUsage: "[TEXT]",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{
				Name:      "text",
				UsageText: "[TEXT]",
				Min:       0,
				Max:       1,
			},
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			textArguments := command.StringArgs("text")
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after text: %q", command.Args().Slice())
			}
			text, err := captureText(command.Reader, textArguments)
			if err != nil {
				return err
			}

			path, err := application.Capture(ctx, text)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(command.Writer, path); err != nil {
				return fmt.Errorf("write captured note path: %w", err)
			}
			return nil
		},
	}
}

func captureText(reader io.Reader, arguments []string) (string, error) {
	if len(arguments) != 0 {
		if strings.TrimSpace(arguments[0]) == "" {
			return "", fmt.Errorf("capture text must not be empty")
		}
		return arguments[0], nil
	}

	if reader == nil {
		return "", fmt.Errorf("capture text is required as an argument or standard input")
	}
	if input, ok := reader.(*os.File); ok {
		info, err := input.Stat()
		if err != nil {
			return "", fmt.Errorf("inspect standard input: %w", err)
		}
		if info.Mode()&os.ModeCharDevice != 0 {
			return "", fmt.Errorf("capture text is required as an argument or standard input")
		}
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read standard input: %w", err)
	}
	text := strings.TrimRight(string(content), "\r\n")
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("capture text is required as an argument or standard input")
	}
	return text, nil
}
