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

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
)

type personClient struct {
	Client
	name   domain.Title
	groups domain.Groups
	tags   domain.Tags
	calls  int
}

func (client *personClient) CreatePerson(_ context.Context, name domain.Title, groups domain.Groups, tags domain.Tags) (app.CreatePersonResult, error) {
	client.name, client.groups, client.tags = name, groups, tags
	client.calls++
	person, err := domain.NewPerson(uuid.New(), name, groups, tags, time.Now())
	return app.CreatePersonResult{Person: person, Path: "/vault/person.adoc"}, err
}

func TestPersonCommandValidatesBeforeCreatingClientAndPassesValues(t *testing.T) {
	client := &personClient{}
	factoryCalls := 0
	factory := func(context.Context) (Client, error) {
		factoryCalls++
		return client, nil
	}
	var output bytes.Buffer
	command := newPersonCommand(factory)
	command.Writer = &output
	if err := command.Run(context.Background(), []string{"person", "--group", "Work", "--group", "University", "--tag", "One", "--tag", "Two", "A name"}); err != nil {
		t.Fatal(err)
	}
	if output.String() != "/vault/person.adoc\n" || client.name.String() != "A name" || client.groups.Len() != 2 || !reflect.DeepEqual(client.groups.Strings(), []string{"work", "university"}) || !reflect.DeepEqual(client.tags.Strings(), []string{"one", "two"}) {
		t.Fatalf("person command output=%q name=%q groups=%v tags=%v", output.String(), client.name, client.groups.Strings(), client.tags.Strings())
	}
	if client.calls != 1 || factoryCalls != 1 {
		t.Fatalf("person calls=%d factory calls=%d", client.calls, factoryCalls)
	}
	command = newPersonCommand(factory)
	command.Writer = io.Discard
	if err := command.Run(context.Background(), []string{"person", "--group", "Work", "--group", "work", "A name"}); !errors.Is(err, domain.ErrDuplicateGroup) {
		t.Fatalf("duplicate groups = %v", err)
	}
	if factoryCalls != 1 {
		t.Fatal("client created for invalid person input")
	}
}
