package repositories_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/domain/bookmarks"
	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/habits"
	"github.com/vekio/overmind/internal/domain/inbox"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/infra/codecs"
	"github.com/vekio/overmind/internal/infra/idgenerator"
	noteindex "github.com/vekio/overmind/internal/infra/index"
	notestore "github.com/vekio/overmind/internal/infra/notestore"
	repository "github.com/vekio/overmind/internal/infra/repositories"
	"github.com/vekio/overmind/internal/ports"
)

func newRepository(t *testing.T, store ports.NoteStore) *repository.HabitRepository {
	t.Helper()
	index, err := noteindex.New(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := index.Close(); err != nil {
			t.Error(err)
		}
	})
	codec := codecs.HabitCodec{}
	return repository.NewHabitRepository(store, index, codec)
}

func TestUseCasesPersistAndReloadAsciiDoc(t *testing.T) {
	ctx := context.Background()
	store := notestore.New(t.TempDir())
	repo := newRepository(t, store)
	application := app.New(app.Dependencies{Habits: repo, IDGenerator: idgenerator.New()})
	for _, period := range []string{"day", "week", "month"} {
		t.Run(period, func(t *testing.T) {
			command := app.CreateHabitCommand{Title: "Beber *agua* {literal}", Amount: 2.12345678912345, Unit: "litros: {literal}", Period: period, Tags: []string{"Bienestar", "Salud"}}
			created, err := application.Commands.CreateHabit.Handle(ctx, command)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := repo.ByID(ctx, created.Habit.ID())
			if err != nil {
				t.Fatal(err)
			}
			if loaded.ID() != created.Habit.ID() || loaded.Title() != created.Habit.Title() || loaded.Goal() != created.Habit.Goal() || loaded.Metadata() != created.Habit.Metadata() || !reflect.DeepEqual(loaded.Tags().Strings(), created.Habit.Tags().Strings()) {
				t.Fatal("habit did not survive document mapping")
			}

		})
	}
}

func TestByIDRejectsMissingAndMalformedNotes(t *testing.T) {
	ctx := context.Background()
	store := notestore.New(t.TempDir())
	repo := newRepository(t, store)
	if _, err := repo.ByID(ctx, uuid.New()); !errors.Is(err, ports.ErrHabitNotFound) {
		t.Fatalf("missing habit = %v", err)
	}
	id := uuid.New()
	valid := "= Agua\n:overmind-id: " + id.String() + "\n:overmind-type: habit\n:overmind-amount: 2\n:overmind-unit: litros\n:overmind-period: day\n:overmind-created-at: 2026-10-03T12:00:00Z\n:overmind-updated-at: 2026-10-03T12:00:00Z\n\n"
	for _, source := range []string{
		strings.Replace(valid, "habit\n", "page\n", 1),
		strings.Replace(valid, id.String(), uuid.New().String(), 1),
		strings.Replace(valid, ":overmind-amount: 2", ":overmind-amount: NaN", 1),
		strings.Replace(valid, ":overmind-period: day", ":overmind-period: year", 1),
		strings.Replace(valid, ":overmind-created-at: 2026-10-03T12:00:00Z", ":overmind-created-at: invalid", 1),
		strings.Replace(valid, ":overmind-updated-at: 2026-10-03T12:00:00Z", ":overmind-updated-at: 2026-10-02T12:00:00Z", 1),
		strings.Replace(valid, ":overmind-unit: litros\n", "", 1),
		strings.Replace(valid, "= Agua\n", "", 1),
		strings.Replace(valid, "\n\n", "\n:overmind-tags: salud, Salud\n\n", 1),
	} {
		if _, err := store.Put(ctx, ports.Note{ID: id, Content: []byte(source)}); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.ByID(ctx, id); err == nil {
			t.Fatalf("invalid note accepted: %s", source)
		}
	}
}

type failingStore struct{ err error }

func (s failingStore) Get(context.Context, uuid.UUID) (ports.Note, error) { return ports.Note{}, s.err }
func (s failingStore) Put(context.Context, ports.Note) (string, error)    { return "", s.err }
func (s failingStore) Delete(context.Context, uuid.UUID) error            { return s.err }

func TestRepositoryPropagatesStorageFailuresAndCancellation(t *testing.T) {
	failure := errors.New("storage failure")
	repo := newRepository(t, failingStore{failure})
	if _, err := repo.ByID(context.Background(), uuid.New()); !errors.Is(err, failure) {
		t.Fatalf("read failure = %v", err)
	}
	if _, err := repo.ByID(context.Background(), uuid.New()); errors.Is(err, ports.ErrHabitNotFound) {
		t.Fatal("storage failure was reported as missing habit")
	}
	application := app.New(app.Dependencies{Habits: repo, IDGenerator: idgenerator.New()})
	_, err := application.Commands.CreateHabit.Handle(context.Background(), app.CreateHabitCommand{Title: "Agua", Amount: 2, Unit: "litros", Period: "day"})
	if !errors.Is(err, failure) {
		t.Fatalf("write failure = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.ByID(ctx, uuid.New()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation = %v", err)
	}
	if err := repo.Save(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled save = %v", err)
	}
	if err := repo.Save(context.Background(), nil); err == nil {
		t.Fatal("nil habit accepted")
	}
	missing := newRepository(t, failingStore{ports.ErrNoteNotFound})
	if _, err := missing.ByID(context.Background(), uuid.New()); !errors.Is(err, ports.ErrHabitNotFound) {
		t.Fatalf("missing note = %v", err)
	}
}

type observingIndex struct {
	err    error
	habits int
}

func (index *observingIndex) UpsertHabit(context.Context, *habits.Habit, string) error {
	index.habits++
	return index.err
}
func TestIndexFailuresLeaveSavedDocumentsAndCanBeRetried(t *testing.T) {
	ctx := context.Background()
	store := notestore.New(t.TempDir())
	failure := errors.New("index unavailable")
	index := &observingIndex{err: failure}
	codec := codecs.HabitCodec{}
	repo := repository.NewHabitRepository(store, index, codec)
	title, _ := shared.NewTitle("Agua")
	unit, _ := habits.NewUnit("litros")
	goal, _ := habits.NewGoal(2, unit, calendar.Day)
	habit, _ := habits.NewHabit(uuid.New(), title, goal, shared.Tags{}, fixtureMetadata())
	if err := repo.Save(ctx, habit); !errors.Is(err, failure) {
		t.Fatalf("index failure = %v", err)
	}
	if _, err := repo.ByID(ctx, habit.ID()); err != nil {
		t.Fatalf("saved document must remain readable when index fails: %v", err)
	}
	index.err = nil
	if err := repo.Save(ctx, habit); err != nil {
		t.Fatal(err)
	}
	if index.habits != 2 {
		t.Fatal("save retry did not refresh index")
	}

}

func TestFailedDocumentWriteDoesNotUpdateIndex(t *testing.T) {
	failure := errors.New("disk unavailable")
	index := &observingIndex{}
	codec := codecs.HabitCodec{}
	repo := repository.NewHabitRepository(failingStore{failure}, index, codec)
	title, _ := shared.NewTitle("Agua")
	unit, _ := habits.NewUnit("litros")
	goal, _ := habits.NewGoal(2, unit, calendar.Day)
	habit, _ := habits.NewHabit(uuid.New(), title, goal, shared.Tags{}, fixtureMetadata())
	if err := repo.Save(context.Background(), habit); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if index.habits != 0 {
		t.Fatal("failed write must not update the index")
	}
	if err := repo.Save(context.Background(), nil); err == nil {
		t.Fatal("nil habit encoded")
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

func (index *observingIndex) UpsertPerson(context.Context, *persons.Person, string) error { return nil }

func (*observingIndex) UpsertBookmark(context.Context, *bookmarks.Bookmark, string) error { return nil }
func (*observingIndex) UpsertInbox(context.Context, *inbox.Inbox, string) error           { return nil }

func (*observingIndex) UpsertPage(context.Context, *pages.Page, string) error          { return nil }
func (*observingIndex) UpsertJournal(context.Context, *journals.Journal, string) error { return nil }
func (*observingIndex) JournalExists(context.Context, calendar.Date) (bool, error)     { return false, nil }
