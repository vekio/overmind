package app_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/domain/journals"
)

type journalTestRepository struct {
	saved *journals.Journal
	err   error
}

func (repo *journalTestRepository) Save(_ context.Context, entity *journals.Journal) error {
	if repo.err != nil {
		return repo.err
	}
	repo.saved = entity
	return nil
}
func TestCreateJournal(t *testing.T) {
	repo := &journalTestRepository{}
	handler := app.New(app.Dependencies{Journals: repo, IDGenerator: testIDGenerator{}}).Commands.CreateJournal
	result, err := handler.Handle(context.Background(), app.CreateJournalCommand{Date: "2024-02-29", Tags: []string{"Daily"}})
	if err != nil {
		t.Fatal(err)
	}
	entity := result.Journal
	if entity != repo.saved || entity.ID() == uuid.Nil() || entity.Date().String() != "2024-02-29" || !reflect.DeepEqual(entity.Tags().Strings(), []string{"daily"}) {
		t.Fatal("raw input was not validated and saved")
	}
	if entity.Metadata().IsZero() || !entity.Metadata().CreatedAt().Equal(entity.Metadata().UpdatedAt()) {
		t.Fatal("invalid initial metadata")
	}
}
func TestCreateJournalRejectsInvalidInputWithoutSaving(t *testing.T) {
	for _, command := range []app.CreateJournalCommand{{Date: "2026-02-29"}, {Date: "yesterday"}, {Tags: []string{"!!!"}}, {Tags: []string{"One", "one"}}} {
		repo := &journalTestRepository{}
		handler := app.New(app.Dependencies{Journals: repo, IDGenerator: testIDGenerator{}}).Commands.CreateJournal
		result, err := handler.Handle(context.Background(), command)
		if err == nil || result.Journal != nil || repo.saved != nil {
			t.Fatal("invalid command saved")
		}
	}
}
func TestCreateJournalDefaultsToTodayAndPropagatesFailures(t *testing.T) {
	repo := &journalTestRepository{}
	handler := app.New(app.Dependencies{Journals: repo, IDGenerator: testIDGenerator{}}).Commands.CreateJournal
	before := time.Now().Format(time.DateOnly)
	result, err := handler.Handle(context.Background(), app.CreateJournalCommand{})
	after := time.Now().Format(time.DateOnly)
	if err != nil {
		t.Fatal(err)
	}
	if date := result.Journal.Date().String(); date != before && date != after {
		t.Fatalf("today=%s", date)
	}
	if !result.Journal.Tags().IsEmpty() {
		t.Fatal("default tags should be empty")
	}
	for _, failure := range []error{app.ErrJournalAlreadyExists, errors.New("disk failure")} {
		repo.err = failure
		if result, err := handler.Handle(context.Background(), app.CreateJournalCommand{}); !errors.Is(err, failure) || result.Journal != nil {
			t.Fatal("save failure not propagated")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := handler.Handle(ctx, app.CreateJournalCommand{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := app.New(app.Dependencies{}).Commands.CreateJournal.Handle(context.Background(), app.CreateJournalCommand{}); err == nil {
		t.Fatal("missing dependencies accepted")
	}
}
