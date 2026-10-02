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
	"github.com/vekio/overmind/internal/domain"
)

type otherCommandsClient struct {
	Client
	bookmarkURL domain.URL
	journalTags domain.Tags
	captureText string
	rebuilds    int
}

func (client *otherCommandsClient) CreateBookmark(_ context.Context, url domain.URL, tags domain.Tags) (app.CreateBookmarkResult, error) {
	client.bookmarkURL = url
	bookmark, err := domain.NewBookmark(uuid.New(), url, tags, time.Now())
	return app.CreateBookmarkResult{Bookmark: bookmark, Path: "/vault/bookmark.adoc"}, err
}

func (client *otherCommandsClient) CreateJournal(_ context.Context, tags domain.Tags) (app.CreateJournalResult, error) {
	client.journalTags = tags
	journal, err := domain.NewJournal(uuid.New(), domain.DateFromTime(time.Now()), tags, time.Now())
	return app.CreateJournalResult{Journal: journal, Path: "/vault/journal.adoc"}, err
}

func (client *otherCommandsClient) Capture(_ context.Context, content string) (app.CaptureResult, error) {
	client.captureText = content
	inbox, err := domain.NewInbox(uuid.New(), content, time.Now())
	return app.CaptureResult{Inbox: inbox, Path: "/vault/inbox.adoc"}, err
}

func (client *otherCommandsClient) RebuildIndex(context.Context) (int, error) {
	client.rebuilds++
	return 3, nil
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
