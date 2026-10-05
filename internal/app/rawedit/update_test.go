package rawedit_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/app/bookmark"
	"github.com/vekio/overmind/internal/app/habit"
	"github.com/vekio/overmind/internal/app/inbox"
	"github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/app/page"
	"github.com/vekio/overmind/internal/app/person"
	"github.com/vekio/overmind/internal/app/rawedit"
	"github.com/vekio/overmind/internal/infra/codecs"
	"github.com/vekio/overmind/internal/infra/idgenerator"
	noteindex "github.com/vekio/overmind/internal/infra/index"
	"github.com/vekio/overmind/internal/infra/notestore"
	"github.com/vekio/overmind/internal/infra/repositories"
	"github.com/vekio/overmind/internal/ports"
)

func fixture(t *testing.T, kind ports.NoteKind) (*app.Application, *notestore.Store, *noteindex.Index, rawedit.GetResult) {
	t.Helper()
	ctx := context.Background()
	store := notestore.New(t.TempDir())
	index, err := noteindex.New(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { index.Close() })
	application := app.New(app.Dependencies{
		NoteStore: store, Index: index, RawNoteCodec: codecs.RawNoteCodec{}, IDGenerator: idgenerator.New(),
		Pages:     repositories.NewPageRepository(store, index, codecs.PageCodec{}),
		Inbox:     repositories.NewInboxRepository(store, index, codecs.InboxCodec{}),
		Bookmarks: repositories.NewBookmarkRepository(store, index, codecs.BookmarkCodec{}),
		Persons:   repositories.NewPersonRepository(store, index, codecs.PersonCodec{}),
		Habits:    repositories.NewHabitRepository(store, index, codecs.HabitCodec{}),
		Journals:  repositories.NewJournalRepository(store, index, codecs.JournalCodec{}),
	})
	var id uuid.UUID
	switch kind {
	case ports.NoteKindPage:
		result, e := application.Commands.CreatePage.Handle(ctx, page.CreateCommand{Title: "Original", Content: "Original body\n"})
		err, id = e, result.Page.ID()
	case ports.NoteKindInbox:
		result, e := application.Commands.CreateInbox.Handle(ctx, inbox.CreateCommand{Content: "Original body\n"})
		err, id = e, result.Inbox.ID()
	case ports.NoteKindBookmark:
		result, e := application.Commands.CreateBookmark.Handle(ctx, bookmark.CreateCommand{URL: "https://example.com", Content: "Original body\n"})
		err, id = e, result.Bookmark.ID()
	case ports.NoteKindPerson:
		result, e := application.Commands.CreatePerson.Handle(ctx, person.CreateCommand{Name: "Original", Content: "Original body\n"})
		err, id = e, result.Person.ID()
	case ports.NoteKindHabit:
		result, e := application.Commands.CreateHabit.Handle(ctx, habit.CreateCommand{Title: "Original", Amount: 1, Unit: "hours", Period: "day", Content: "Original body\n"})
		err, id = e, result.Habit.ID()
	case ports.NoteKindJournal:
		result, e := application.Commands.CreateJournal.Handle(ctx, journal.CreateCommand{Date: "2026-10-05", Content: "Original body\n"})
		err, id = e, result.Journal.ID()
	}
	if err != nil {
		t.Fatal(err)
	}
	original, err := application.Queries.GetRawNote.Handle(ctx, rawedit.GetQuery{ID: id.String()})
	if err != nil {
		t.Fatal(err)
	}
	return application, store, index, original
}

func TestRawEditPersistsAllKindsAndUpdatesIndex(t *testing.T) {
	for _, kind := range []ports.NoteKind{ports.NoteKindPage, ports.NoteKindInbox, ports.NoteKindBookmark, ports.NoteKindPerson, ports.NoteKindHabit, ports.NoteKindJournal} {
		t.Run(kind.String(), func(t *testing.T) {
			application, store, index, original := fixture(t, kind)
			source := strings.ReplaceAll(string(original.Note.Content), "Original body\n", "== Edited\n\nUnicode: ñ 📝  \n")
			source = strings.ReplaceAll(source, ":overmind-tags: ", ":overmind-tags: edited")
			source = strings.ReplaceAll(source, "= Original\n", "= Renamed\n")
			source = strings.ReplaceAll(source, "https://example.com", "https://example.org")
			source = strings.ReplaceAll(source, ":overmind-amount: 1", ":overmind-amount: 2")
			result, err := application.Commands.UpdateRawNote.Handle(context.Background(), rawedit.UpdateCommand{ID: original.Note.ID.String(), Kind: kind, Original: original.Note.Content, Source: []byte(source)})
			if err != nil || !result.Changed {
				t.Fatalf("update: %+v, %v", result, err)
			}
			stored, err := store.Get(context.Background(), original.Note.ID)
			if err != nil {
				t.Fatal(err)
			}
			decoder := codecs.RawNoteCodec{}
			before, _ := decoder.Decode(kind, original.Note.Content)
			after, err := decoder.Decode(kind, stored.Content)
			if err != nil {
				t.Fatal(err)
			}
			if after.ID != before.ID || !after.Metadata.CreatedAt().Equal(before.Metadata.CreatedAt()) || !after.Metadata.UpdatedAt().After(before.Metadata.UpdatedAt()) {
				t.Fatal("identity or timestamps incorrect")
			}
			rows, err := index.FindNotes(context.Background(), ports.NoteFilter{Tag: "edited", Limit: 10})
			if err != nil || len(rows) != 1 || rows[0].ID != before.ID || !rows[0].UpdatedAt.Equal(after.Metadata.UpdatedAt()) {
				t.Fatalf("index not updated: %+v, %v", rows, err)
			}
			if (kind == ports.NoteKindPage || kind == ports.NoteKindPerson || kind == ports.NoteKindHabit) && rows[0].Label != "Renamed" {
				t.Fatal("updated title was not indexed")
			}
		})
	}
}

func TestRawEditRejectsInvalidChangesWithoutWriting(t *testing.T) {
	cases := []struct {
		name   string
		kind   ports.NoteKind
		change func(string) string
	}{
		{"ID", ports.NoteKindPage, func(s string) string {
			i := strings.Index(s, ":overmind-id: ")
			start := i + len(":overmind-id: ")
			return s[:start] + "11111111-1111-4111-8111-111111111111" + s[start+36:]
		}},
		{"type", ports.NoteKindPage, func(s string) string { return strings.Replace(s, ":overmind-type: page", ":overmind-type: inbox", 1) }},
		{"creation time", ports.NoteKindPage, func(s string) string {
			i := strings.Index(s, ":overmind-created-at: ")
			start := i + len(":overmind-created-at: ")
			end := start + strings.IndexByte(s[start:], '\n')
			return s[:start] + "2020-01-01T00:00:00Z" + s[end:]
		}},
		{"journal date", ports.NoteKindJournal, func(s string) string {
			return strings.Replace(s, ":overmind-date: 2026-10-05", ":overmind-date: 2026-10-06", 1)
		}},
		{"title", ports.NoteKindPage, func(s string) string { return strings.Replace(s, "= Original", "= ", 1) }},
		{"URL", ports.NoteKindBookmark, func(s string) string { return strings.Replace(s, "https://example.com", "invalid-url", 1) }},
		{"goal", ports.NoteKindHabit, func(s string) string { return strings.Replace(s, ":overmind-amount: 1", ":overmind-amount: 0", 1) }},
		{"AsciiDoc", ports.NoteKindPage, func(s string) string { return s + "\n----\nUnclosed listing\n" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			application, store, _, original := fixture(t, tc.kind)
			_, err := application.Commands.UpdateRawNote.Handle(context.Background(), rawedit.UpdateCommand{ID: original.Note.ID.String(), Kind: tc.kind, Original: original.Note.Content, Source: []byte(tc.change(string(original.Note.Content)))})
			if err == nil {
				t.Fatal("invalid change accepted")
			}
			actual, _ := store.Get(context.Background(), original.Note.ID)
			if !bytes.Equal(actual.Content, original.Note.Content) {
				t.Fatal("invalid edit changed vault")
			}
		})
	}
}

func TestRawEditNoopAndConflict(t *testing.T) {
	application, store, _, original := fixture(t, ports.NoteKindPage)
	command := rawedit.UpdateCommand{ID: original.Note.ID.String(), Kind: original.Kind, Original: original.Note.Content, Source: original.Note.Content}
	result, err := application.Commands.UpdateRawNote.Handle(context.Background(), command)
	actual, _ := store.Get(context.Background(), original.Note.ID)
	if err != nil || result.Changed || !bytes.Equal(actual.Content, original.Note.Content) {
		t.Fatal("unchanged edit wrote note")
	}
	concurrent := original.Note
	concurrent.Content = []byte(strings.Replace(string(concurrent.Content), "Original body", "Concurrent edit", 1))
	if _, err := store.Put(context.Background(), concurrent); err != nil {
		t.Fatal(err)
	}
	command.Source = []byte(strings.Replace(string(original.Note.Content), "Original body", "My edit", 1))
	if _, err := application.Commands.UpdateRawNote.Handle(context.Background(), command); err == nil {
		t.Fatal("concurrent edit overwritten")
	}
	actual, _ = store.Get(context.Background(), original.Note.ID)
	if !bytes.Equal(actual.Content, concurrent.Content) {
		t.Fatal("conflict changed vault")
	}
}

func TestRawEditPreservesExactSource(t *testing.T) {
	for _, kind := range []ports.NoteKind{ports.NoteKindPage, ports.NoteKindInbox, ports.NoteKindBookmark, ports.NoteKindPerson, ports.NoteKindHabit, ports.NoteKindJournal} {
		for _, ending := range []string{"\n", "\r\n"} {
			t.Run(kind.String()+"/"+strings.ReplaceAll(ending, "\n", "LF"), func(t *testing.T) {
				application, store, _, original := fixture(t, kind)
				source := string(original.Note.Content)
				source = strings.Replace(source, ":overmind-type:", "// Custom header comment\n:custom-option: preserve me\n:overmind-type:", 1)
				source = strings.Replace(source, "Original body\n", "== Contenido\n\nUnicode ñ 📝\t  \n:overmind-updated-at: body attribute\n\n// Body comment\nFinal without newline", 1)
				source = strings.ReplaceAll(source, "\n", ending)
				result, err := application.Commands.UpdateRawNote.Handle(context.Background(), rawedit.UpdateCommand{ID: original.Note.ID.String(), Kind: kind, Original: original.Note.Content, Source: []byte(source)})
				if err != nil {
					t.Fatal(err)
				}
				// Replace only the managed value in the expected source, independently of Stamp.
				header, _ := (codecs.RawNoteCodec{}).Inspect(result.Note.Content)
				// Target the updated declaration specifically: creation can have the same value.
				prefix := ":overmind-updated-at: "
				start := strings.Index(source, prefix) + len(prefix)
				end := start + strings.IndexAny(source[start:], "\r\n")
				expected := source[:start] + header.Metadata.UpdatedAt().Format(time.RFC3339Nano) + source[end:]
				actual, err := store.Get(context.Background(), original.Note.ID)
				if err != nil || !bytes.Equal(actual.Content, []byte(expected)) {
					t.Fatalf("source changed outside timestamp:\n%q\nexpected:\n%q, %v", actual.Content, expected, err)
				}
			})
		}
	}
}

func TestRawEditRepairsInvalidBody(t *testing.T) {
	application, store, _, original := fixture(t, ports.NoteKindPage)
	invalid := original.Note
	invalid.Content = append(append([]byte(nil), invalid.Content...), []byte("\n----\nUnclosed listing\n")...)
	if _, err := store.Put(context.Background(), invalid); err != nil {
		t.Fatal(err)
	}
	loaded, err := application.Queries.GetRawNote.Handle(context.Background(), rawedit.GetQuery{ID: invalid.ID.String()})
	if err != nil || !bytes.Equal(loaded.Note.Content, invalid.Content) {
		t.Fatalf("cannot open invalid body: %v", err)
	}
	result, err := application.Commands.UpdateRawNote.Handle(context.Background(), rawedit.UpdateCommand{ID: invalid.ID.String(), Kind: loaded.Kind, Original: loaded.Note.Content, Source: append(append([]byte(nil), loaded.Note.Content...), []byte("----\n")...)})
	if err != nil || !result.Changed {
		t.Fatalf("cannot repair invalid body: %v", err)
	}
}

type countingStore struct {
	ports.NoteStore
	writes             int
	writeErr           error
	readCount          int
	changeOnSecondRead []byte
}

func (store *countingStore) Put(ctx context.Context, note ports.Note) (string, error) {
	store.writes++
	if store.writeErr != nil {
		return "", store.writeErr
	}
	return store.NoteStore.Put(ctx, note)
}

func (store *countingStore) Get(ctx context.Context, id uuid.UUID) (ports.Note, error) {
	store.readCount++
	if store.readCount == 2 && store.changeOnSecondRead != nil {
		if _, err := store.NoteStore.Put(ctx, ports.Note{ID: id, Content: store.changeOnSecondRead}); err != nil {
			return ports.Note{}, err
		}
	}
	return store.NoteStore.Get(ctx, id)
}

type failingProjector struct {
	ports.Index
	calls int
	err   error
}

func (projector *failingProjector) Project(ctx context.Context, note ports.Note) error {
	projector.calls++
	if projector.calls == 1 {
		if projector.err != nil {
			return projector.err
		}
		return errors.New("database temporarily unavailable")
	}
	return projector.Index.Project(ctx, note)
}

func TestRawEditRetriesOnlyIndex(t *testing.T) {
	_, store, index, original := fixture(t, ports.NoteKindPage)
	counting := &countingStore{NoteStore: store}
	projector := &failingProjector{Index: index}
	handler := rawedit.NewUpdateHandler(counting, codecs.RawNoteCodec{}, projector)
	source := []byte(strings.Replace(string(original.Note.Content), "Original body", "Edited body", 1))
	first, err := handler.Handle(context.Background(), rawedit.UpdateCommand{ID: original.Note.ID.String(), Kind: original.Kind, Original: original.Note.Content, Source: source})
	if !errors.Is(err, rawedit.ErrIndexUpdate) || !first.IndexPending || !first.Changed || counting.writes != 1 {
		t.Fatalf("index failure was not distinguished: %+v, %v", first, err)
	}
	stored, _ := store.Get(context.Background(), original.Note.ID)
	if !bytes.Equal(stored.Content, first.Note.Content) {
		t.Fatal("saved result does not match vault")
	}
	second, err := handler.Handle(context.Background(), rawedit.UpdateCommand{ID: original.Note.ID.String(), Kind: original.Kind, Original: first.Note.Content, IndexOnly: true})
	if err != nil || second.IndexPending || counting.writes != 1 || !bytes.Equal(second.Note.Content, first.Note.Content) {
		t.Fatalf("retry rewrote note or changed timestamp: %+v, %v", second, err)
	}
	rows, err := index.FindNotes(context.Background(), ports.NoteFilter{Limit: 10})
	after, _ := (codecs.RawNoteCodec{}).Inspect(second.Note.Content)
	if err != nil || len(rows) != 1 || !rows[0].UpdatedAt.Equal(after.Metadata.UpdatedAt()) {
		t.Fatalf("retry did not repair index: %+v, %v", rows, err)
	}
}

func TestRawEditDetectsChangesDuringValidation(t *testing.T) {
	_, store, index, original := fixture(t, ports.NoteKindPage)
	concurrent := []byte(strings.Replace(string(original.Note.Content), "Original body", "Concurrent body", 1))
	counting := &countingStore{NoteStore: store, changeOnSecondRead: concurrent}
	handler := rawedit.NewUpdateHandler(counting, codecs.RawNoteCodec{}, index)
	_, err := handler.Handle(context.Background(), rawedit.UpdateCommand{ID: original.Note.ID.String(), Kind: original.Kind, Original: original.Note.Content, Source: append(append([]byte(nil), original.Note.Content...), []byte("Edited\n")...)})
	if !errors.Is(err, rawedit.ErrConflict) || counting.writes != 0 {
		t.Fatalf("late conflict was overwritten: %v", err)
	}
}

func TestRawEditStorageFailureDoesNotTouchIndex(t *testing.T) {
	_, store, index, original := fixture(t, ports.NoteKindPage)
	failure := errors.New("disk unavailable")
	counting := &countingStore{NoteStore: store, writeErr: failure}
	observing := &failingProjector{Index: index}
	handler := rawedit.NewUpdateHandler(counting, codecs.RawNoteCodec{}, observing)
	source := bytes.Replace(original.Note.Content, []byte("Original body"), []byte("Unsaved change"), 1)
	result, err := handler.Handle(context.Background(), rawedit.UpdateCommand{
		ID: original.Note.ID.String(), Kind: original.Kind, Original: original.Note.Content, Source: source,
	})
	if !errors.Is(err, failure) || result.Changed || result.IndexPending || observing.calls != 0 {
		t.Fatalf("storage failure was treated as saved or indexed: %+v, %v, calls=%d", result, err, observing.calls)
	}
	current, err := store.Get(context.Background(), original.Note.ID)
	if err != nil || !bytes.Equal(current.Content, original.Note.Content) {
		t.Fatalf("failed write altered source: %v", err)
	}
}

func TestRawEditIndexFailurePreservesCauseAndRetryRejectsNewSource(t *testing.T) {
	_, store, index, original := fixture(t, ports.NoteKindPage)
	failure := errors.New("database offline")
	counting := &countingStore{NoteStore: store}
	observing := &failingProjector{Index: index, err: failure}
	handler := rawedit.NewUpdateHandler(counting, codecs.RawNoteCodec{}, observing)
	source := bytes.Replace(original.Note.Content, []byte("Original body"), []byte("Saved change"), 1)
	saved, err := handler.Handle(context.Background(), rawedit.UpdateCommand{
		ID: original.Note.ID.String(), Kind: original.Kind, Original: original.Note.Content, Source: source,
	})
	if !errors.Is(err, rawedit.ErrIndexUpdate) || !errors.Is(err, failure) || !saved.IndexPending {
		t.Fatalf("partial save lost index cause: %+v, %v", saved, err)
	}
	concurrent := saved.Note
	concurrent.Content = bytes.Replace(saved.Note.Content, []byte("Saved change"), []byte("Another writer"), 1)
	if _, err := store.Put(context.Background(), concurrent); err != nil {
		t.Fatal(err)
	}
	_, err = handler.Handle(context.Background(), rawedit.UpdateCommand{
		ID: original.Note.ID.String(), Kind: original.Kind, Original: saved.Note.Content, IndexOnly: true,
	})
	if !errors.Is(err, rawedit.ErrConflict) || observing.calls != 1 || counting.writes != 1 {
		t.Fatalf("retry indexed a superseded snapshot: %v, calls=%d, writes=%d", err, observing.calls, counting.writes)
	}
	current, err := store.Get(context.Background(), original.Note.ID)
	if err != nil || !bytes.Equal(current.Content, concurrent.Content) {
		t.Fatalf("retry altered concurrent source: %v", err)
	}
}
