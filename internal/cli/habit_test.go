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

	apphabit "github.com/vekio/overmind/internal/app/habit"
	"github.com/vekio/overmind/internal/bootstrap"
	appconfig "github.com/vekio/overmind/internal/config"
	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/infra/index/sqlitedb"
)

type habitClient struct {
	Client
	input apphabit.CreateCommand
	calls int
	err   error
}

func (client *habitClient) CreateHabit(_ context.Context, input apphabit.CreateCommand) (apphabit.CreateResult, error) {
	client.calls++
	client.input = input
	if client.err != nil {
		return apphabit.CreateResult{}, client.err
	}
	title, err := shared.NewTitle(input.Title)
	if err != nil {
		return apphabit.CreateResult{}, err
	}
	unit, err := habits.NewUnit(input.Unit)
	if err != nil {
		return apphabit.CreateResult{}, err
	}
	goal, err := habits.NewGoal(input.Amount, unit, calendar.Day)
	if err != nil {
		return apphabit.CreateResult{}, err
	}
	habit, err := habits.NewHabit(uuid.MustParse("11111111-1111-4111-8111-111111111111"), title, goal, shared.Tags{}, fixtureMetadata())
	return apphabit.CreateResult{Habit: habit}, err
}

func TestHabitCommandPassesFlagsAndTitleToUseCase(t *testing.T) {
	client := &habitClient{}
	output := runCLICommand(t, newHabitCommand(fixedClient(client)), []string{"habit", "--amount", "0.5", "--unit", "litros", "--period", "day", "--tag", "Salud", "--tag", "Bienestar", "Beber agua"}, "")
	expected := apphabit.CreateCommand{Title: "Beber agua", Amount: 0.5, Unit: "litros", Period: "day", Tags: []string{"Salud", "Bienestar"}}
	if client.calls != 1 || !reflect.DeepEqual(client.input, expected) {
		t.Fatalf("use case input = %+v", client.input)
	}
	if output != "Beber agua: 0.5 litros/day\nID: 11111111-1111-4111-8111-111111111111\n" {
		t.Fatalf("output = %q", output)
	}
}

func TestHabitCommandRequiresFlagsAndExactlyOneTitle(t *testing.T) {
	cases := [][]string{
		{"habit", "--unit", "litros", "--period", "day", "Agua"},
		{"habit", "--amount", "2", "--period", "day", "Agua"},
		{"habit", "--amount", "2", "--unit", "litros", "Agua"},
		{"habit", "--amount", "2", "--unit", "litros", "--period", "day"},
		{"habit", "--amount", "2", "--unit", "litros", "--period", "day", "Beber", "agua"},
		{"habit", "--amount", "two", "--unit", "litros", "--period", "day", "Agua"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			calls := 0
			command := newHabitCommand(func(context.Context) (Client, error) { calls++; return &habitClient{}, nil })
			command.Writer, command.ErrWriter = io.Discard, io.Discard
			if err := command.Run(context.Background(), args); err == nil {
				t.Fatal("incomplete or malformed input accepted")
			}
			if calls != 0 {
				t.Fatal("client initialized for invalid command syntax")
			}
		})
	}
}

func TestHabitCommandPropagatesUseCaseFailureWithoutSuccessOutput(t *testing.T) {
	failure := errors.New("save failed")
	client := &habitClient{err: failure}
	command := newHabitCommand(fixedClient(client))
	var output bytes.Buffer
	command.Writer = &output
	err := command.Run(context.Background(), []string{"habit", "--amount", "2", "--unit", "litros", "--period", "day", "Agua"})
	if !errors.Is(err, failure) || output.Len() != 0 {
		t.Fatalf("failure = %v, output = %q", err, output.String())
	}
}

func TestRootHabitCommandPersistsViaLocalUseCase(t *testing.T) {
	ctx := context.Background()
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
	defer func() {
		if err := runtime.Close(); err != nil {
			t.Error(err)
		}
	}()
	factory := func(ctx context.Context) (Client, error) {
		application, err := runtime.Application(ctx)
		if err != nil {
			return nil, err
		}
		return NewLocalClient(application), nil
	}
	for _, period := range []string{"day", "week", "month"} {
		t.Run(period, func(t *testing.T) {
			command := New(config, factory)
			var output bytes.Buffer
			command.Writer, command.ErrWriter = &output, io.Discard
			if err := command.Run(ctx, []string{"overmind", "habit", "--amount", "2", "--unit", "litros", "--period", period, "Beber agua"}); err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(strings.TrimSpace(output.String()), "\n")
			if len(lines) != 2 || lines[0] != "Beber agua: 2 litros/"+period {
				t.Fatalf("output = %q", output.String())
			}
			id, err := uuid.Parse(strings.TrimPrefix(lines[1], "ID: "))
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(filepath.Join(vault, "notes", id.String()+".adoc"))
			if err != nil || !bytes.Contains(source, []byte(":overmind-period: "+period)) {
				t.Fatalf("stored document = %s, %v", source, err)
			}
			db, err := sql.Open("sqlite", filepath.Join(vault, "index.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			row, err := sqlitedb.New(db).HabitByID(ctx, id.String())
			if err != nil || row.Title != "Beber agua" || row.Period != period {
				t.Fatalf("indexed habit = %+v, %v", row, err)
			}
			indexedTags, err := sqlitedb.New(db).TagsByNoteID(ctx, id.String())
			if err != nil || len(indexedTags) != 0 {
				t.Fatalf("optional tags = %v, %v", indexedTags, err)
			}
		})
	}
	// Domain validation stays in CreateHabit, including duplicate normalized tags.
	command := New(config, factory)
	var output bytes.Buffer
	command.Writer, command.ErrWriter = &output, io.Discard
	err = command.Run(ctx, []string{"overmind", "habit", "--amount", "2", "--unit", "litros", "--period", "day", "--tag", "Salud", "--tag", "salud", "Agua"})
	if !errors.Is(err, shared.ErrDuplicateTag) || output.Len() != 0 {
		t.Fatalf("duplicate tags = %v, output = %q", err, output.String())
	}
}

func fixtureMetadata() shared.EntityMetadata {
	at := time.Date(2026, 10, 3, 12, 0, 0, 123456789, time.UTC)
	metadata, err := shared.NewEntityMetadata(at, at)
	if err != nil {
		panic(err)
	}
	return metadata
}
