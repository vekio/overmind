package repositories_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
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
	"github.com/vekio/overmind/internal/infra/codecs"
	"github.com/vekio/overmind/internal/infra/idgenerator"
	noteindex "github.com/vekio/overmind/internal/infra/index"
	"github.com/vekio/overmind/internal/infra/index/sqlitedb"
	notestore "github.com/vekio/overmind/internal/infra/notestore"
	repository "github.com/vekio/overmind/internal/infra/repositories"
	"github.com/vekio/overmind/internal/ports"
)

func TestPersonRepositoryRoundTripAndIndex(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	path := filepath.Join(root, "index.db")
	index, err := noteindex.New(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	store := notestore.New(filepath.Join(root, "notes"))
	repo := repository.NewPersonRepository(store, index, codecs.PersonCodec{})
	application := app.New(app.Dependencies{Persons: repo, IDGenerator: idgenerator.New()})
	result, err := application.Commands.CreatePerson.Handle(ctx, app.CreatePersonCommand{Name: "Ana García", Groups: []string{"Trabajo", "Universidad"}, Tags: []string{"Amiga", "Contacto"}})
	if err != nil {
		t.Fatal(err)
	}
	person := result.Person
	loaded, err := repo.ByID(ctx, person.ID())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name() != person.Name() || loaded.Metadata() != person.Metadata() || !reflect.DeepEqual(loaded.Groups().Strings(), person.Groups().Strings()) || !reflect.DeepEqual(loaded.Tags().Strings(), person.Tags().Strings()) {
		t.Fatal("person did not survive codec mapping")
	}
	note, err := store.Get(ctx, person.ID())
	if err != nil {
		t.Fatal(err)
	}
	for _, section := range []string{"== Contact", "== Context", "== Notes"} {
		if !strings.Contains(string(note.Content), section) {
			t.Fatalf("missing %s", section)
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	queries := sqlitedb.New(db)
	row, err := queries.PersonByID(ctx, person.ID().String())
	if err != nil || row.Name != "Ana García" || row.Type != "person" || row.Path != note.Path {
		t.Fatalf("indexed person = %+v, %v", row, err)
	}
	groups, err := queries.GroupsByPersonID(ctx, person.ID().String())
	if err != nil || !reflect.DeepEqual(groups, []string{"trabajo", "universidad"}) {
		t.Fatalf("indexed groups = %v, %v", groups, err)
	}
	tags, err := queries.TagsByNoteID(ctx, person.ID().String())
	if err != nil || !reflect.DeepEqual(tags, []string{"amiga", "contacto"}) {
		t.Fatalf("indexed tags = %v, %v", tags, err)
	}
	if err := person.Regroup(persons.Groups{}, person.Metadata().UpdatedAt().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, person); err != nil {
		t.Fatal(err)
	}
	groups, err = queries.GroupsByPersonID(ctx, person.ID().String())
	if err != nil || len(groups) != 0 {
		t.Fatal("old memberships were not removed")
	}
	if _, err := repo.ByID(ctx, uuid.New()); !errors.Is(err, ports.ErrPersonNotFound) {
		t.Fatalf("missing person = %v", err)
	}

}

type failingIndex struct {
	err    error
	calls  int
	person *persons.Person
}

func (index *failingIndex) UpsertHabit(context.Context, *habits.Habit, string) error { return nil }
func (index *failingIndex) UpsertPerson(_ context.Context, person *persons.Person, _ string) error {
	index.person = person
	index.calls++
	return index.err
}

func TestIndexFailureKeepsSavedDocument(t *testing.T) {
	ctx := context.Background()
	store := notestore.New(t.TempDir())
	failure := errors.New("index unavailable")
	index := &failingIndex{err: failure}
	repo := repository.NewPersonRepository(store, index, codecs.PersonCodec{})
	application := app.New(app.Dependencies{Persons: repo, IDGenerator: idgenerator.New()})
	result, err := application.Commands.CreatePerson.Handle(ctx, app.CreatePersonCommand{Name: "Ana"})
	if !errors.Is(err, failure) || result.Person != nil || index.calls != 1 {
		t.Fatalf("index failure = %v", err)
	}
	loaded, readErr := repo.ByID(ctx, index.person.ID())
	if readErr != nil || loaded.Name().String() != "Ana" {
		t.Fatalf("saved document must remain readable: %v", readErr)
	}
	if err := repo.Save(ctx, nil); err == nil || index.calls != 1 {
		t.Fatal("invalid person reached index")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repo.ByID(canceled, uuid.New()); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestCodecRejectsWrongKindDuplicateMembershipsAndInvalidDates(t *testing.T) {
	codec := codecs.PersonCodec{}
	id := uuid.New()
	source := "= Ana\n:overmind-id: " + id.String() + "\n:overmind-type: person\n:overmind-groups: work\n:overmind-tags: amiga\n:overmind-created-at: 2026-10-04T12:00:00Z\n:overmind-updated-at: 2026-10-04T12:00:00Z\n\n"
	if _, err := codec.Decode([]byte(source)); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{
		strings.Replace(source, ":overmind-type: person", ":overmind-type: habit", 1),
		strings.Replace(source, ":overmind-groups: work", ":overmind-groups: work, Work", 1),
		strings.Replace(source, ":overmind-tags: amiga", ":overmind-tags: amiga, Amiga", 1),
		strings.Replace(source, ":overmind-updated-at: 2026-10-04T12:00:00Z", ":overmind-updated-at: 2026-10-03T12:00:00Z", 1),
	} {
		if _, err := codec.Decode([]byte(invalid)); err == nil {
			t.Fatalf("invalid document accepted: %s", invalid)
		}
	}
}

var _ ports.Index = (*failingIndex)(nil)

func (*failingIndex) UpsertBookmark(context.Context, *bookmarks.Bookmark, string) error { return nil }
func (*failingIndex) UpsertInbox(context.Context, *inbox.Inbox, string) error           { return nil }

func (*failingIndex) UpsertPage(context.Context, *pages.Page, string) error          { return nil }
func (*failingIndex) UpsertJournal(context.Context, *journals.Journal, string) error { return nil }
func (*failingIndex) JournalExists(context.Context, calendar.Date) (bool, error)     { return false, nil }
