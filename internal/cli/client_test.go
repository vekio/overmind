package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
	"uuid"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/shared"
)

type otherCommandsClient struct {
	Client
	bookmarkURL  string
	bookmarkTags []string
	inboxTags    []string
	journalTags  []string
	inboxContent string
}

func (client *otherCommandsClient) CreateBookmark(_ context.Context, command app.CreateBookmarkCommand) (app.CreateBookmarkResult, error) {
	client.bookmarkURL, client.bookmarkTags = command.URL, command.Tags
	url, err := bookmarks.NewURL(command.URL)
	if err != nil {
		return app.CreateBookmarkResult{}, err
	}
	now := time.Now()
	metadata, _ := shared.NewEntityMetadata(now, now)
	bookmark, err := bookmarks.NewBookmark(uuid.MustParse("11111111-1111-4111-8111-111111111111"), url, shared.Tags{}, metadata)
	return app.CreateBookmarkResult{Bookmark: bookmark}, err
}

func (client *otherCommandsClient) CreateJournal(_ context.Context, command app.CreateJournalCommand) (app.CreateJournalResult, error) {
	client.journalTags = command.Tags
	text := command.Date
	if text == "" {
		text = "2026-10-04"
	}
	date, _ := calendar.NewDate(text)
	now := time.Now()
	metadata, _ := shared.NewEntityMetadata(now, now)
	entity, err := journals.NewJournal(uuid.MustParse("11111111-1111-4111-8111-111111111111"), date, shared.Tags{}, metadata)
	return app.CreateJournalResult{Journal: entity}, err
}

func (client *otherCommandsClient) CreateInbox(_ context.Context, command app.CreateInboxCommand) (app.CreateInboxResult, error) {
	client.inboxContent, client.inboxTags = command.Content, command.Tags
	now := time.Now()
	metadata, _ := shared.NewEntityMetadata(now, now)
	note, err := inbox.NewInbox(uuid.MustParse("11111111-1111-4111-8111-111111111111"), command.Content, shared.Tags{}, metadata)
	return app.CreateInboxResult{Inbox: note}, err
}

func runCLICommand(t *testing.T, command *urfavecli.Command, args []string, input string) string {
	t.Helper()
	var output bytes.Buffer
	command.Writer = &output
	command.Reader = strings.NewReader(input)
	if err := command.Run(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func fixedClient(client Client) ClientFactory {
	return func(context.Context) (Client, error) { return client, nil }
}
