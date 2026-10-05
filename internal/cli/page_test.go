package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/app/page"
	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/shared"
)

type pageClient struct {
	Client
	input page.CreateCommand
	calls int
}

func (client *pageClient) CreatePage(_ context.Context, command page.CreateCommand) (page.CreateResult, error) {
	client.input = command
	client.calls++
	title, _ := shared.NewTitle(command.Title)
	now := time.Now()
	metadata, _ := shared.NewEntityMetadata(now, now)
	entity, err := pages.NewPage(uuid.MustParse("11111111-1111-4111-8111-111111111111"), title, pages.Area{}, shared.Tags{}, metadata)
	return page.CreateResult{Page: entity}, err
}
func TestPageCommandPassesRawInput(t *testing.T) {
	client := &pageClient{}
	output := runCLICommand(t, newPageCommand(fixedClient(client)), []string{"page", "--area", "Work/Ideas", "--tag", "One", "--tag", "Two", "A title"}, "")
	expected := page.CreateCommand{Title: "A title", Area: "Work/Ideas", Tags: []string{"One", "Two"}}
	if client.calls != 1 || !reflect.DeepEqual(client.input, expected) || output != "A title\nID: 11111111-1111-4111-8111-111111111111\n" {
		t.Fatalf("input=%+v output=%q", client.input, output)
	}
}
func TestPageCommandRequiresOneTitle(t *testing.T) {
	for _, args := range [][]string{{"page"}, {"page", "A", "title"}} {
		client := &pageClient{}
		command := newPageCommand(fixedClient(client))
		command.Writer = io.Discard
		if err := command.Run(context.Background(), args); err == nil || client.calls != 0 {
			t.Fatal("invalid argument count accepted")
		}
	}
}

func TestRootPageCommandPersistsDocumentAndProjection(t *testing.T) {
	fixture := newCommandFixture(t)
	note := fixture.create("page", []string{"page", "--area", "Work/Ideas", "--tag", "Reading", "--tag", "Work", "Plan"}, "", []string{"reading", "work"})
	var title, area string
	if err := fixture.db().QueryRow("SELECT title,area FROM pages WHERE note_id=?", note.ID.String()).Scan(&title, &area); err != nil || title != "Plan" || area != "work/ideas" {
		t.Fatalf("page projection=%q %q err=%v", title, area, err)
	}
	if !bytes.HasPrefix(note.Source, []byte("= Plan\n")) || !bytes.Contains(note.Source, []byte(":overmind-area: work/ideas")) {
		t.Fatalf("page template=%s", note.Source)
	}
	if _, err := fixture.run([]string{"page", "!!!"}, ""); !errors.Is(err, shared.ErrInvalidTitle) {
		t.Fatalf("invalid title=%v", err)
	}
	fixture.requireSingleNote()
}
