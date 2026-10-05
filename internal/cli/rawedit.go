package cli

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app/rawedit"
	"github.com/vekio/overmind/internal/editor"
)

func newRawEditCommand(newClient ClientFactory) *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "rawedit",
		Usage:     "edit a complete AsciiDoc note with $EDITOR",
		ArgsUsage: "ID",
		Arguments: noteIDArgument(),
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return fmt.Errorf("unexpected arguments: %q", command.Args().Slice())
			}
			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			original, err := client.GetRawNote(ctx, rawedit.GetQuery{ID: command.StringArgs("id")[0]})
			if err != nil {
				return err
			}
			draft, err := editor.New(original.Note.Content)
			if err != nil {
				return err
			}
			// Success or explicit discard removes the draft; every other exit preserves it.
			discard := false
			defer func() {
				if discard {
					draft.Close()
				} else {
					_, _ = draft.Finish()
				}
			}()
			input := bufio.NewReader(command.Reader)
			openEditor, indexPending := true, false
			for {
				var editErr error
				if openEditor {
					process := draft.Command(ctx)
					process.Stdin, process.Stdout, process.Stderr = command.Reader, command.Writer, command.ErrWriter
					editErr = process.Run()
				}
				var result rawedit.UpdateResult
				if editErr == nil {
					source, readErr := draft.Read()
					editErr = readErr
					if editErr == nil {
						result, editErr = client.UpdateRawNote(ctx, rawedit.UpdateCommand{
							ID: original.Note.ID.String(), Kind: original.Kind, Original: original.Note.Content, Source: source,
							IndexOnly: indexPending && bytes.Equal(source, original.Note.Content),
						})
					}
				}
				if editErr == nil {
					discard = true
					text := "No changes."
					if result.Changed {
						text = "Note updated."
					}
					_, err := fmt.Fprintln(command.Writer, text)
					return err
				}
				draft.Keep()
				if result.IndexPending {
					// Retry from the bytes already saved, including the managed timestamp.
					original.Note = result.Note
					indexPending = true
					if err := draft.Write(result.Note.Content); err != nil {
						return fmt.Errorf("%v; draft retained at %s: %w", editErr, draft.Path, err)
					}
				}
				if ctx.Err() != nil {
					return fmt.Errorf("draft retained at %s: %w", draft.Path, ctx.Err())
				}
				fmt.Fprintf(command.ErrWriter, "%v\nDraft: %s\n", editErr, draft.Path)
				if indexPending {
					fmt.Fprint(command.ErrWriter, "[r] Retry indexing, [k] Keep draft, [d] Discard draft (default keep): ")
				} else {
					fmt.Fprint(command.ErrWriter, "[e] Reopen editor, [r] Retry save, [k] Keep draft, [d] Discard draft (default keep): ")
				}
				answer, readErr := input.ReadString('\n')
				if readErr == nil {
					switch strings.ToLower(strings.TrimSpace(answer)) {
					case "r":
						openEditor = false
						continue
					case "e", "y":
						openEditor = true
						continue
					case "d":
						discard = true
						return fmt.Errorf("draft discarded: %w", editErr)
					}
				}
				return fmt.Errorf("draft retained at %s: %w", draft.Path, editErr)
			}
		},
	}
}
