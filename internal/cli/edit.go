package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"git.casta.me/alberto/overmind/internal/app/getdocument"
	"git.casta.me/alberto/overmind/internal/app/updatedocument"
	urfavecli "github.com/urfave/cli/v3"
)

func newEditCommand(state *applicationState) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "edit",
		Usage:     "edit an existing document",
		ArgsUsage: "[PATH]",
		Arguments: []urfavecli.Argument{
			&urfavecli.StringArgs{Name: "path", UsageText: "PATH", Min: 0, Max: 1},
		},
		ShellComplete: func(ctx context.Context, command *urfavecli.Command) {
			completeDocumentPaths(ctx, command, state)
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments after path: %q", command.Args().Slice())
			}
			path, err := documentPath(command, "path")
			if err != nil {
				return err
			}
			application, err := state.Application()
			if err != nil {
				return err
			}
			current, err := application.Queries.GetDocument.Handle(ctx, getdocument.GetDocumentQuery{Path: path})
			if err != nil {
				return err
			}
			edited, err := editDocument(ctx, current.Content, command.ErrWriter)
			if err != nil {
				return err
			}
			if bytes.Equal(current.Content, edited) {
				return nil
			}
			_, err = application.Commands.UpdateDocument.Handle(ctx, updatedocument.UpdateDocumentCommand{
				Path:             current.Path,
				Content:          edited,
				ExpectedRevision: current.Revision,
			})
			return err
		},
	}
}

func editDocument(ctx context.Context, content []byte, errorOutput io.Writer) ([]byte, error) {
	temporary, err := os.CreateTemp("", "overmind-*.adoc")
	if err != nil {
		return nil, fmt.Errorf("create temporary document: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return nil, fmt.Errorf("write temporary document: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return nil, fmt.Errorf("close temporary document: %w", err)
	}

	editor, err := resolveEditor()
	if err != nil {
		return nil, err
	}
	arguments := append(append([]string(nil), editor.arguments...), temporaryPath)
	process := exec.CommandContext(ctx, editor.executable, arguments...)
	process.Stderr = errorOutput
	terminal, terminalErr := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if terminalErr == nil {
		defer terminal.Close()
		process.Stdin = terminal
		process.Stdout = terminal
	} else {
		process.Stdin = os.Stdin
		process.Stdout = errorOutput
	}
	if err := process.Run(); err != nil {
		return nil, fmt.Errorf("run editor %q: %w", editor.executable, err)
	}

	edited, err := os.ReadFile(temporaryPath)
	if err != nil {
		return nil, fmt.Errorf("read edited document: %w", err)
	}
	return edited, nil
}

type editorCommand struct {
	executable string
	arguments  []string
}

func resolveEditor() (editorCommand, error) {
	candidates := []string{os.Getenv("VISUAL"), os.Getenv("EDITOR"), "vi"}
	for _, value := range candidates {
		fields := strings.Fields(value)
		if len(fields) == 0 {
			continue
		}
		executable, err := exec.LookPath(fields[0])
		if err != nil {
			continue
		}
		return editorCommand{executable: executable, arguments: fields[1:]}, nil
	}
	return editorCommand{}, fmt.Errorf("no editor found; set VISUAL or EDITOR")
}
