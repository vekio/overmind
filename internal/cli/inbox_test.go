package cli

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/vekio/overmind/internal/domain/shared"
)

func TestInboxTextUsesArgumentOrPipedInput(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		read io.Reader
		want string
	}{
		{"argument", []string{"first\nsecond"}, strings.NewReader("ignored"), "first\nsecond"},
		{"stdin", nil, strings.NewReader("first\nsecond\n\n"), "first\nsecond"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := inboxText(tc.read, tc.args)
			if err != nil || got != tc.want {
				t.Fatalf("inboxText() = %q, %v", got, err)
			}
		})
	}
	for _, args := range [][]string{nil, {"  "}} {
		if _, err := inboxText(strings.NewReader("\n  \n"), args); err == nil {
			t.Fatalf("empty inbox accepted with args %v", args)
		}
	}
	if _, err := inboxText(nil, nil); err == nil {
		t.Fatal("missing inbox input accepted")
	}
}

func TestInboxCommandReadsPipedTextAndPrintsID(t *testing.T) {
	client := &otherCommandsClient{}
	got := runCLICommand(t, newInboxCommand(fixedClient(client)), []string{"inbox"}, "piped text\n")
	if got != "Inbox created\nID: 11111111-1111-4111-8111-111111111111\n" || client.inboxContent != "piped text" {
		t.Fatalf("inbox output=%q text=%q", got, client.inboxContent)
	}
}

func TestInboxCommandPreservesArgumentContent(t *testing.T) {
	client := &otherCommandsClient{}
	content := "Una idea\n:overmind-type: body\n\nOtra línea.\n"
	runCLICommand(t, newInboxCommand(fixedClient(client)), []string{"inbox", "--", content}, "")
	if client.inboxContent != content {
		t.Fatalf("argument changed: %q", client.inboxContent)
	}
}

func TestRootInboxCommandPreservesBodyAndProjection(t *testing.T) {
	for _, tc := range []struct {
		name, content, input string
		args                 []string
	}{
		{"argument", "Una idea\n:overmind-type: body\n\nOtra línea.\n", "", []string{"inbox", "--tag", "Reading", "--tag", "Work", "--", "Una idea\n:overmind-type: body\n\nOtra línea.\n"}},
		{"stdin", "Una idea\nOtra línea.", "Una idea\nOtra línea.\n", []string{"inbox", "--tag", "Reading", "--tag", "Work"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newCommandFixture(t)
			note := fixture.create("inbox", tc.args, tc.input, []string{"reading", "work"})
			if !bytes.HasSuffix(note.Source, []byte("\n\n"+tc.content)) {
				t.Fatalf("inbox body changed: %q", note.Source)
			}
			var content string
			if err := fixture.db().QueryRow("SELECT content FROM inbox WHERE note_id=?", note.ID.String()).Scan(&content); err != nil || content != tc.content {
				t.Fatalf("inbox projection=%q err=%v", content, err)
			}
			if _, err := fixture.run([]string{"inbox", "--tag", "!!!", "idea"}, ""); !errors.Is(err, shared.ErrInvalidTag) {
				t.Fatalf("invalid tag=%v", err)
			}
			fixture.requireSingleNote()
		})
	}
}
