package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"uuid"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/app/inbox"
	"github.com/vekio/overmind/internal/app/rawedit"
	"github.com/vekio/overmind/internal/infra/codecs"
	"github.com/vekio/overmind/internal/infra/idgenerator"
	noteindex "github.com/vekio/overmind/internal/infra/index"
	"github.com/vekio/overmind/internal/infra/notestore"
	"github.com/vekio/overmind/internal/infra/repositories"
	"github.com/vekio/overmind/internal/ports"
)

func TestRawEditCLI(t *testing.T) {
	for _, tc := range []struct {
		name, script  string
		changed, fail bool
	}{
		{"edit", "printf '\\nEdited from external editor\\n' >> \"$1\"", true, false},
		{"no changes", "exit 0", false, false},
		{"invalid document", "printf '\\n----\\nUnclosed\\n' >> \"$1\"", false, true},
		{"editor failure", "printf '\\nRecover me\\n' >> \"$1\"; exit 1", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := notestore.New(t.TempDir())
			index, err := noteindex.New(ctx, ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer index.Close()
			application := app.New(app.Dependencies{NoteStore: store, Index: index, RawNoteCodec: codecs.RawNoteCodec{}, IDGenerator: idgenerator.New(), Inbox: repositories.NewInboxRepository(store, index, codecs.InboxCodec{})})
			created, err := application.Commands.CreateInbox.Handle(ctx, inbox.CreateCommand{Content: "Original content\n"})
			if err != nil {
				t.Fatal(err)
			}
			before, _ := store.Get(ctx, created.Inbox.ID())
			script := filepath.Join(t.TempDir(), "editor")
			if err := os.WriteFile(script, []byte("#!/bin/sh\n"+tc.script+"\n"), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("EDITOR", script)
			var output bytes.Buffer
			client := NewLocalClient(application)
			command := &urfavecli.Command{Name: "overmind", Reader: strings.NewReader("n\n"), Writer: &output, ErrWriter: &output, Commands: []*urfavecli.Command{newRawEditCommand(func(context.Context) (Client, error) { return client, nil })}}
			err = command.Run(ctx, []string{"overmind", "rawedit", created.Inbox.ID().String()})
			after, readErr := store.Get(ctx, created.Inbox.ID())
			if readErr != nil {
				t.Fatal(readErr)
			}
			if tc.fail {
				if err == nil || !bytes.Equal(before.Content, after.Content) {
					t.Fatalf("failed edit changed note: %v", err)
				}
				text := err.Error()
				start := strings.Index(text, "draft retained at ") + len("draft retained at ")
				end := strings.Index(text[start:], ": ") + start
				if start < len("draft retained at ") || end < start {
					t.Fatalf("missing recovery path: %v", err)
				}
				path := text[start:end]
				defer os.Remove(path)
				if _, err := os.Stat(path); err != nil {
					t.Fatalf("failed draft lost: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if tc.changed {
					if !bytes.Contains(after.Content, []byte("Edited from external editor")) || !strings.Contains(output.String(), "Note updated") {
						t.Fatal("editor content was not saved")
					}
				} else if !bytes.Equal(before.Content, after.Content) || !strings.Contains(output.String(), "No changes") {
					t.Fatal("no-op edit changed note")
				}
			}
		})
	}
}

type rawRetryClient struct {
	Client
	original  ports.Note
	calls     int
	indexOnly bool
}

func (client *rawRetryClient) GetRawNote(context.Context, rawedit.GetQuery) (rawedit.GetResult, error) {
	return rawedit.GetResult{Note: client.original, Kind: ports.NoteKindInbox}, nil
}

func (client *rawRetryClient) UpdateRawNote(_ context.Context, command rawedit.UpdateCommand) (rawedit.UpdateResult, error) {
	client.calls++
	client.indexOnly = command.IndexOnly
	if client.calls == 1 {
		client.original.Content = []byte("persisted source including updated timestamp")
		return rawedit.UpdateResult{Changed: true, Note: client.original, IndexPending: true}, rawedit.ErrIndexUpdate
	}
	if !bytes.Equal(command.Original, client.original.Content) || !command.IndexOnly {
		return rawedit.UpdateResult{}, errors.New("retry rewrote the note")
	}
	return rawedit.UpdateResult{Changed: true, Note: client.original}, nil
}

func TestRawEditCLIRetriesIndexWithoutReopeningEditor(t *testing.T) {
	root := t.TempDir()
	count := filepath.Join(root, "editor-count")
	script := filepath.Join(root, "editor")
	text := "#!/bin/sh\nprintf 'called\\n' >> '" + count + "'\nprintf 'Edited\\n' >> \"$1\"\n"
	if err := os.WriteFile(script, []byte(text), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", script)
	client := &rawRetryClient{original: ports.Note{ID: uuid.New(), Content: []byte("initial source")}}
	input, err := os.CreateTemp(t.TempDir(), "input")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	if _, err := input.WriteString("r\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := input.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	command := &urfavecli.Command{Name: "overmind", Reader: input, Writer: &output, ErrWriter: &output, Commands: []*urfavecli.Command{newRawEditCommand(func(context.Context) (Client, error) { return client, nil })}}
	if err := command.Run(context.Background(), []string{"overmind", "rawedit", client.original.ID.String()}); err != nil {
		t.Fatal(err)
	}
	calls, err := os.ReadFile(count)
	if err != nil || string(calls) != "called\n" || !client.indexOnly || client.calls != 2 {
		t.Fatal("index retry reopened editor or saved document")
	}
}
