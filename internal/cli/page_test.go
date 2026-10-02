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

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/domain"
)

type pageClient struct {
	Client
	title domain.Title
	area  domain.Area
	tags  domain.Tags
	calls int
}

func (client *pageClient) CreatePage(_ context.Context, title domain.Title, area domain.Area, tags domain.Tags) (app.CreatePageResult, error) {
	client.title, client.area, client.tags = title, area, tags
	client.calls++
	page, err := domain.NewPage(uuid.New(), title, area, tags, time.Now())
	return app.CreatePageResult{Page: page, Path: "/vault/page.adoc"}, err
}

func TestPageCommandValidatesBeforeCreatingClientAndPassesValues(t *testing.T) {
	client := &pageClient{}
	factoryCalls := 0
	factory := func(context.Context) (Client, error) {
		factoryCalls++
		return client, nil
	}
	var output bytes.Buffer
	command := newPageCommand(factory)
	command.Writer = &output
	if err := command.Run(context.Background(), []string{"page", "--area", "Work/Ideas", "--tag", "One", "--tag", "Two", "A title"}); err != nil {
		t.Fatal(err)
	}
	if output.String() != "/vault/page.adoc\n" || client.title.String() != "A title" || client.area.String() != "work/ideas" || !reflect.DeepEqual(client.tags.Strings(), []string{"one", "two"}) {
		t.Fatalf("page command output=%q title=%q area=%q tags=%v", output.String(), client.title, client.area, client.tags.Strings())
	}
	if client.calls != 1 || factoryCalls != 1 {
		t.Fatalf("page calls=%d factory calls=%d", client.calls, factoryCalls)
	}
	command = newPageCommand(factory)
	command.Writer = io.Discard
	if err := command.Run(context.Background(), []string{"page", "--tag", "same", "--tag", "same", "A title"}); !errors.Is(err, domain.ErrDuplicateTag) {
		t.Fatalf("duplicate tags = %v", err)
	}
	if factoryCalls != 1 {
		t.Fatal("client created for invalid page input")
	}
}
