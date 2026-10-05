package journal_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/ports"
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

func (repo *journalTestRepository) ByID(_ context.Context, id uuid.UUID) (*journals.Journal, error) {
	if repo.err != nil {
		return nil, repo.err
	}
	if repo.saved == nil || repo.saved.ID() != id {
		return nil, ports.ErrNoteNotFound
	}
	return repo.saved, nil
}

func (repo *journalTestRepository) Update(ctx context.Context, entity *journals.Journal) error {
	return repo.Save(ctx, entity)
}
func TestCreateJournal(t *testing.T) {
	repo := &journalTestRepository{}
	handler := app.New(app.Dependencies{Journals: repo, IDGenerator: testIDGenerator{}}).Commands.CreateJournal
	result, err := handler.Handle(context.Background(), journal.CreateCommand{Date: "2024-02-29", Content: "My day\nA second line", Tags: []string{"Daily"}})
	if err != nil {
		t.Fatal(err)
	}
	entity := result.Journal
	if entity != repo.saved || entity.ID() == uuid.Nil() || entity.Date().String() != "2024-02-29" || entity.Content() != "My day\nA second line" || !reflect.DeepEqual(entity.Tags().Strings(), []string{"daily"}) {
		t.Fatal("raw input was not validated and saved")
	}
	if entity.Metadata().IsZero() || !entity.Metadata().CreatedAt().Equal(entity.Metadata().UpdatedAt()) {
		t.Fatal("invalid initial metadata")
	}
}
func TestCreateJournalRejectsInvalidInputWithoutSaving(t *testing.T) {
	for _, command := range []journal.CreateCommand{{Date: "2026-02-29"}, {Date: "yesterday"}, {Tags: []string{"!!!"}}, {Tags: []string{"One", "one"}}} {
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
	result, err := handler.Handle(context.Background(), journal.CreateCommand{})
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
	for _, failure := range []error{journal.ErrAlreadyExists, errors.New("disk failure")} {
		repo.err = failure
		if result, err := handler.Handle(context.Background(), journal.CreateCommand{}); !errors.Is(err, failure) || result.Journal != nil {
			t.Fatal("save failure not propagated")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := handler.Handle(ctx, journal.CreateCommand{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := app.New(app.Dependencies{}).Commands.CreateJournal.Handle(context.Background(), journal.CreateCommand{}); err == nil {
		t.Fatal("missing dependencies accepted")
	}
}

// testIDGenerator keeps ID generation independent of infrastructure.
type testIDGenerator struct{}

func (testIDGenerator) Generate() uuid.UUID { return uuid.New() }
