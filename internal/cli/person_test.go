package cli

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
	"uuid"

	appperson "github.com/vekio/overmind/internal/app/person"
	"github.com/vekio/overmind/internal/bootstrap"
	appconfig "github.com/vekio/overmind/internal/config"
	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/infra/index/sqlitedb"
)

type personClient struct {
	Client
	input appperson.CreateCommand
	calls int
	err   error
}

func (client *personClient) CreatePerson(_ context.Context, command appperson.CreateCommand) (appperson.CreateResult, error) {
	client.input = command
	client.calls++
	if client.err != nil {
		return appperson.CreateResult{}, client.err
	}
	name, _ := shared.NewTitle(command.Name)
	now := time.Now()
	metadata, _ := shared.NewEntityMetadata(now, now)
	person, err := persons.NewPerson(uuid.MustParse("11111111-1111-4111-8111-111111111111"), name, persons.Groups{}, shared.Tags{}, metadata)
	return appperson.CreateResult{Person: person}, err
}
func TestPersonCommandPassesRawInput(t *testing.T) {
	client := &personClient{}
	got := runCLICommand(t, newPersonCommand(fixedClient(client)), []string{"person", "--group", "Work", "--group", "University", "--tag", "One", "--tag", "Two", "Ana García"}, "")
	expected := appperson.CreateCommand{Name: "Ana García", Groups: []string{"Work", "University"}, Tags: []string{"One", "Two"}}
	if client.calls != 1 || !reflect.DeepEqual(client.input, expected) || got != "Ana García\nID: 11111111-1111-4111-8111-111111111111\n" {
		t.Fatalf("person input=%+v output=%q", client.input, got)
	}
}
func TestPersonCommandRequiresOneName(t *testing.T) {
	for _, args := range [][]string{{"person"}, {"person", "Ana", "García"}} {
		client := &personClient{}
		command := newPersonCommand(fixedClient(client))
		command.Writer = io.Discard
		if err := command.Run(context.Background(), args); err == nil || client.calls != 0 {
			t.Fatal("missing or unquoted name accepted")
		}
	}
}
func TestPersonCommandPropagatesUseCaseFailure(t *testing.T) {
	failure := errors.New("save failure")
	command := newPersonCommand(fixedClient(&personClient{err: failure}))
	command.Writer = io.Discard
	if err := command.Run(context.Background(), []string{"person", "Ana"}); !errors.Is(err, failure) {
		t.Fatal(err)
	}
}

func TestRootPersonCommandUsesCurrentPersistence(t *testing.T) {
	root := t.TempDir()
	config, err := appconfig.NewFile()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.SetPath(filepath.Join(root, "config.yml")); err != nil {
		t.Fatal(err)
	}
	vault := filepath.Join(root, "vault")
	if err := config.Create(appconfig.Settings{Mode: appconfig.ModeLocal, VaultPath: vault}); err != nil {
		t.Fatal(err)
	}
	runtime := bootstrap.New(config)
	defer runtime.Close()
	factory := func(ctx context.Context) (Client, error) {
		application, err := runtime.Application(ctx)
		if err != nil {
			return nil, err
		}
		return NewLocalClient(application), nil
	}
	command := New(config, factory)
	var output bytes.Buffer
	command.Writer, command.ErrWriter = &output, io.Discard
	if err := command.Run(context.Background(), []string{"overmind", "person", "--group", "Work", "--group", "University", "--tag", "Friend", "Ana García"}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 || lines[0] != "Ana García" {
		t.Fatalf("output = %q", output.String())
	}
	id, err := uuid.Parse(strings.TrimPrefix(lines[1], "ID: "))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join(vault, "notes", id.String()+".adoc"))
	if err != nil || !bytes.Contains(source, []byte(":overmind-groups: work, university")) {
		t.Fatalf("stored person = %s, %v", source, err)
	}
	client, err := factory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	listed, err := client.ListGroups(context.Background(), appperson.ListGroupsQuery{})
	if err != nil || !reflect.DeepEqual(listed.Groups, []string{"university", "work"}) {
		t.Fatalf("listed groups=%v, error=%v", listed.Groups, err)
	}
	db, err := sql.Open("sqlite", filepath.Join(vault, "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	row, err := sqlitedb.New(db).PersonByID(context.Background(), id.String())
	if err != nil || row.Name != "Ana García" || row.Type != "person" {
		t.Fatalf("indexed person = %+v, %v", row, err)
	}
	command = New(config, factory)
	command.Writer, command.ErrWriter = io.Discard, io.Discard
	if err := command.Run(context.Background(), []string{"overmind", "person", "--group", "Work", "--group", "work", "Ana"}); !errors.Is(err, persons.ErrDuplicateGroup) {
		t.Fatalf("duplicate groups = %v", err)
	}
}
