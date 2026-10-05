package repositories_test

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/app/bookmark"
	"github.com/vekio/overmind/internal/app/habit"
	"github.com/vekio/overmind/internal/app/inbox"
	"github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/app/page"
	"github.com/vekio/overmind/internal/app/person"
	"github.com/vekio/overmind/internal/infra/codecs"
	"github.com/vekio/overmind/internal/infra/idgenerator"
	noteindex "github.com/vekio/overmind/internal/infra/index"
	"github.com/vekio/overmind/internal/infra/notestore"
	"github.com/vekio/overmind/internal/infra/repositories"
	"github.com/vekio/overmind/internal/ports"
)

func TestEditUseCasesPersistAllTypes(t *testing.T) {
	ctx := context.Background()
	store := notestore.New(t.TempDir())
	index, err := noteindex.New(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	application := app.New(app.Dependencies{
		IDGenerator: idgenerator.New(), Index: index,
		Habits:    repositories.NewHabitRepository(store, index, codecs.HabitCodec{}),
		Persons:   repositories.NewPersonRepository(store, index, codecs.PersonCodec{}),
		Bookmarks: repositories.NewBookmarkRepository(store, index, codecs.BookmarkCodec{}),
		Inbox:     repositories.NewInboxRepository(store, index, codecs.InboxCodec{}),
		Pages:     repositories.NewPageRepository(store, index, codecs.PageCodec{}),
		Journals:  repositories.NewJournalRepository(store, index, codecs.JournalCodec{}),
	})
	body := "\n== Content\n:overmind-type: body-only\n\nText with trailing spaces  \n"

	t.Run("habit", func(t *testing.T) {
		created, err := application.Commands.CreateHabit.Handle(ctx, habit.CreateCommand{Content: "Initial content\nSecond line\n", Title: "Old habit", Amount: 1, Unit: "hours", Period: "day", Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		original := created.Habit
		if original.Content() != "Initial content\nSecond line\n" {
			t.Fatal("creation dropped content")
		}
		note, err := store.Get(ctx, original.ID())
		if err != nil {
			t.Fatal(err)
		}
		note.Content = append(note.Content, []byte(body)...)
		if _, err := store.Put(ctx, note); err != nil {
			t.Fatal(err)
		}
		loadedBefore, err := application.Queries.GetHabit.Handle(ctx, habit.GetQuery{ID: original.ID().String()})
		if err != nil || !bytes.HasSuffix([]byte(loadedBefore.Habit.Content()), []byte(body)) {
			t.Fatal("existing document body was not loaded")
		}
		updated, err := application.Commands.UpdateHabit.Handle(ctx, habit.UpdateCommand{ID: original.ID().String(), Content: body, Title: "New habit", Amount: 2.5, Unit: "sessions", Period: "month", Tags: []string{"updated"}})
		if err != nil {
			t.Fatal(err)
		}
		entity := updated.Habit
		if entity.ID() != original.ID() || !entity.Metadata().CreatedAt().Equal(original.Metadata().CreatedAt()) || !entity.Metadata().UpdatedAt().After(original.Metadata().UpdatedAt()) || !reflect.DeepEqual(entity.Tags().Strings(), []string{"updated"}) {
			t.Fatal("update changed identity/creation time or failed to update metadata/tags")
		}
		if entity.Content() != body || entity.Title().String() != "New habit" || entity.Goal().Amount() != 2.5 || entity.Goal().Unit().String() != "sessions" || entity.Goal().Period().String() != "month" {
			t.Fatal("updated fields did not persist")
		}
		reloaded, err := application.Queries.GetHabit.Handle(ctx, habit.GetQuery{ID: original.ID().String()})
		if err != nil || !reflect.DeepEqual(reloaded.Habit, entity) {
			t.Fatalf("round trip changed entity: %v", err)
		}
		stored, err := store.Get(ctx, original.ID())
		if err != nil || !bytes.HasSuffix(stored.Content, []byte(body)) {
			t.Fatalf("document content was lost: %s, %v", stored.Content, err)
		}
		summaries, err := index.FindNotes(ctx, ports.NoteFilter{Type: "habit", Limit: 100})
		if err != nil || len(summaries) != 1 || summaries[0].ID != entity.ID() || !reflect.DeepEqual(summaries[0].Tags, []string{"updated"}) {
			t.Fatalf("index did not update: %v, %v", summaries, err)
		}
		if _, err := application.Queries.GetHabit.Handle(ctx, habit.GetQuery{ID: "bad-id"}); err == nil {
			t.Fatal("invalid lookup ID accepted")
		}
		if _, err := application.Commands.UpdateHabit.Handle(ctx, habit.UpdateCommand{ID: uuid.New().String(), Content: body, Title: "New habit", Amount: 2.5, Unit: "sessions", Period: "month", Tags: []string{"updated"}}); err == nil {
			t.Fatal("update created a missing note")
		}
		if _, err := application.Commands.UpdateHabit.Handle(ctx, habit.UpdateCommand{ID: entity.ID().String(), Tags: []string{"!!!"}}); err == nil {
			t.Fatal("invalid update accepted")
		}
		unchanged, err := application.Queries.GetHabit.Handle(ctx, habit.GetQuery{ID: entity.ID().String()})
		if err != nil || !reflect.DeepEqual(unchanged.Habit, entity) {
			t.Fatal("invalid update changed stored note")
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := application.Commands.UpdateHabit.Handle(cancelled, habit.UpdateCommand{ID: entity.ID().String(), Content: body, Title: "New habit", Amount: 2.5, Unit: "sessions", Period: "month", Tags: []string{"updated"}}); !errors.Is(err, context.Canceled) {
			t.Fatal("cancelled update was not rejected")
		}
		cleared, err := application.Commands.UpdateHabit.Handle(ctx, habit.UpdateCommand{ID: entity.ID().String(), Content: "", Title: "New habit", Amount: 2.5, Unit: "sessions", Period: "month", Tags: []string{"updated"}})
		if err != nil || cleared.Habit.Content() != "" {
			t.Fatalf("empty content did not clear body: %v", err)
		}
		reloadedEmpty, err := application.Queries.GetHabit.Handle(ctx, habit.GetQuery{ID: entity.ID().String()})
		if err != nil || reloadedEmpty.Habit.Content() != "" {
			t.Fatal("cleared content reappeared on reload")
		}

	})

	t.Run("person", func(t *testing.T) {
		created, err := application.Commands.CreatePerson.Handle(ctx, person.CreateCommand{Content: "Initial content\nSecond line\n", Name: "Old person", Groups: []string{"work"}, Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		original := created.Person
		if original.Content() != "Initial content\nSecond line\n" {
			t.Fatal("creation dropped content")
		}
		note, err := store.Get(ctx, original.ID())
		if err != nil {
			t.Fatal(err)
		}
		note.Content = append(note.Content, []byte(body)...)
		if _, err := store.Put(ctx, note); err != nil {
			t.Fatal(err)
		}
		loadedBefore, err := application.Queries.GetPerson.Handle(ctx, person.GetQuery{ID: original.ID().String()})
		if err != nil || !bytes.HasSuffix([]byte(loadedBefore.Person.Content()), []byte(body)) {
			t.Fatal("existing document body was not loaded")
		}
		updated, err := application.Commands.UpdatePerson.Handle(ctx, person.UpdateCommand{ID: original.ID().String(), Content: body, Name: "New person", Groups: []string{"friends"}, Tags: []string{"updated"}})
		if err != nil {
			t.Fatal(err)
		}
		entity := updated.Person
		if entity.ID() != original.ID() || !entity.Metadata().CreatedAt().Equal(original.Metadata().CreatedAt()) || !entity.Metadata().UpdatedAt().After(original.Metadata().UpdatedAt()) || !reflect.DeepEqual(entity.Tags().Strings(), []string{"updated"}) {
			t.Fatal("update changed identity/creation time or failed to update metadata/tags")
		}
		if entity.Content() != body || entity.Name().String() != "New person" || !reflect.DeepEqual(entity.Groups().Strings(), []string{"friends"}) {
			t.Fatal("updated fields did not persist")
		}
		reloaded, err := application.Queries.GetPerson.Handle(ctx, person.GetQuery{ID: original.ID().String()})
		if err != nil || !reflect.DeepEqual(reloaded.Person, entity) {
			t.Fatalf("round trip changed entity: %v", err)
		}
		stored, err := store.Get(ctx, original.ID())
		if err != nil || !bytes.HasSuffix(stored.Content, []byte(body)) {
			t.Fatalf("document content was lost: %s, %v", stored.Content, err)
		}
		summaries, err := index.FindNotes(ctx, ports.NoteFilter{Type: "person", Limit: 100})
		if err != nil || len(summaries) != 1 || summaries[0].ID != entity.ID() || !reflect.DeepEqual(summaries[0].Tags, []string{"updated"}) {
			t.Fatalf("index did not update: %v, %v", summaries, err)
		}
		if _, err := application.Queries.GetPerson.Handle(ctx, person.GetQuery{ID: "bad-id"}); err == nil {
			t.Fatal("invalid lookup ID accepted")
		}
		if _, err := application.Commands.UpdatePerson.Handle(ctx, person.UpdateCommand{ID: uuid.New().String(), Content: body, Name: "New person", Groups: []string{"friends"}, Tags: []string{"updated"}}); err == nil {
			t.Fatal("update created a missing note")
		}
		if _, err := application.Commands.UpdatePerson.Handle(ctx, person.UpdateCommand{ID: entity.ID().String(), Tags: []string{"!!!"}}); err == nil {
			t.Fatal("invalid update accepted")
		}
		unchanged, err := application.Queries.GetPerson.Handle(ctx, person.GetQuery{ID: entity.ID().String()})
		if err != nil || !reflect.DeepEqual(unchanged.Person, entity) {
			t.Fatal("invalid update changed stored note")
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := application.Commands.UpdatePerson.Handle(cancelled, person.UpdateCommand{ID: entity.ID().String(), Content: body, Name: "New person", Groups: []string{"friends"}, Tags: []string{"updated"}}); !errors.Is(err, context.Canceled) {
			t.Fatal("cancelled update was not rejected")
		}
		cleared, err := application.Commands.UpdatePerson.Handle(ctx, person.UpdateCommand{ID: entity.ID().String(), Content: "", Name: "New person", Groups: []string{"friends"}, Tags: []string{"updated"}})
		if err != nil || cleared.Person.Content() != "" {
			t.Fatalf("empty content did not clear body: %v", err)
		}
		reloadedEmpty, err := application.Queries.GetPerson.Handle(ctx, person.GetQuery{ID: entity.ID().String()})
		if err != nil || reloadedEmpty.Person.Content() != "" {
			t.Fatal("cleared content reappeared on reload")
		}

	})

	t.Run("bookmark", func(t *testing.T) {
		created, err := application.Commands.CreateBookmark.Handle(ctx, bookmark.CreateCommand{Content: "Initial content\nSecond line\n", URL: "https://example.com/old", Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		original := created.Bookmark
		if original.Content() != "Initial content\nSecond line\n" {
			t.Fatal("creation dropped content")
		}
		note, err := store.Get(ctx, original.ID())
		if err != nil {
			t.Fatal(err)
		}
		note.Content = append(note.Content, []byte(body)...)
		if _, err := store.Put(ctx, note); err != nil {
			t.Fatal(err)
		}
		loadedBefore, err := application.Queries.GetBookmark.Handle(ctx, bookmark.GetQuery{ID: original.ID().String()})
		if err != nil || !bytes.HasSuffix([]byte(loadedBefore.Bookmark.Content()), []byte(body)) {
			t.Fatal("existing document body was not loaded")
		}
		updated, err := application.Commands.UpdateBookmark.Handle(ctx, bookmark.UpdateCommand{ID: original.ID().String(), Content: body, URL: "https://example.com/new", Tags: []string{"updated"}})
		if err != nil {
			t.Fatal(err)
		}
		entity := updated.Bookmark
		if entity.ID() != original.ID() || !entity.Metadata().CreatedAt().Equal(original.Metadata().CreatedAt()) || !entity.Metadata().UpdatedAt().After(original.Metadata().UpdatedAt()) || !reflect.DeepEqual(entity.Tags().Strings(), []string{"updated"}) {
			t.Fatal("update changed identity/creation time or failed to update metadata/tags")
		}
		if entity.Content() != body || entity.URL().String() != "https://example.com/new" {
			t.Fatal("updated fields did not persist")
		}
		reloaded, err := application.Queries.GetBookmark.Handle(ctx, bookmark.GetQuery{ID: original.ID().String()})
		if err != nil || !reflect.DeepEqual(reloaded.Bookmark, entity) {
			t.Fatalf("round trip changed entity: %v", err)
		}
		stored, err := store.Get(ctx, original.ID())
		if err != nil || !bytes.HasSuffix(stored.Content, []byte(body)) {
			t.Fatalf("document content was lost: %s, %v", stored.Content, err)
		}
		summaries, err := index.FindNotes(ctx, ports.NoteFilter{Type: "bookmark", Limit: 100})
		if err != nil || len(summaries) != 1 || summaries[0].ID != entity.ID() || !reflect.DeepEqual(summaries[0].Tags, []string{"updated"}) {
			t.Fatalf("index did not update: %v, %v", summaries, err)
		}
		if _, err := application.Queries.GetBookmark.Handle(ctx, bookmark.GetQuery{ID: "bad-id"}); err == nil {
			t.Fatal("invalid lookup ID accepted")
		}
		if _, err := application.Commands.UpdateBookmark.Handle(ctx, bookmark.UpdateCommand{ID: uuid.New().String(), Content: body, URL: "https://example.com/new", Tags: []string{"updated"}}); err == nil {
			t.Fatal("update created a missing note")
		}
		if _, err := application.Commands.UpdateBookmark.Handle(ctx, bookmark.UpdateCommand{ID: entity.ID().String(), Tags: []string{"!!!"}}); err == nil {
			t.Fatal("invalid update accepted")
		}
		unchanged, err := application.Queries.GetBookmark.Handle(ctx, bookmark.GetQuery{ID: entity.ID().String()})
		if err != nil || !reflect.DeepEqual(unchanged.Bookmark, entity) {
			t.Fatal("invalid update changed stored note")
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := application.Commands.UpdateBookmark.Handle(cancelled, bookmark.UpdateCommand{ID: entity.ID().String(), Content: body, URL: "https://example.com/new", Tags: []string{"updated"}}); !errors.Is(err, context.Canceled) {
			t.Fatal("cancelled update was not rejected")
		}
		cleared, err := application.Commands.UpdateBookmark.Handle(ctx, bookmark.UpdateCommand{ID: entity.ID().String(), Content: "", URL: "https://example.com/new", Tags: []string{"updated"}})
		if err != nil || cleared.Bookmark.Content() != "" {
			t.Fatalf("empty content did not clear body: %v", err)
		}
		reloadedEmpty, err := application.Queries.GetBookmark.Handle(ctx, bookmark.GetQuery{ID: entity.ID().String()})
		if err != nil || reloadedEmpty.Bookmark.Content() != "" {
			t.Fatal("cleared content reappeared on reload")
		}

	})

	t.Run("inbox", func(t *testing.T) {
		created, err := application.Commands.CreateInbox.Handle(ctx, inbox.CreateCommand{Content: "Old content", Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		original := created.Inbox
		updated, err := application.Commands.UpdateInbox.Handle(ctx, inbox.UpdateCommand{ID: original.ID().String(), Content: body, Tags: []string{"updated"}})
		if err != nil {
			t.Fatal(err)
		}
		entity := updated.Inbox
		if entity.ID() != original.ID() || !entity.Metadata().CreatedAt().Equal(original.Metadata().CreatedAt()) || !entity.Metadata().UpdatedAt().After(original.Metadata().UpdatedAt()) || !reflect.DeepEqual(entity.Tags().Strings(), []string{"updated"}) {
			t.Fatal("update changed identity/creation time or failed to update metadata/tags")
		}
		if entity.Content() != body {
			t.Fatal("updated fields did not persist")
		}
		reloaded, err := application.Queries.GetInbox.Handle(ctx, inbox.GetQuery{ID: original.ID().String()})
		if err != nil || !reflect.DeepEqual(reloaded.Inbox, entity) {
			t.Fatalf("round trip changed entity: %v", err)
		}
		stored, err := store.Get(ctx, original.ID())
		if err != nil || !bytes.HasSuffix(stored.Content, []byte(body)) {
			t.Fatalf("document content was lost: %s, %v", stored.Content, err)
		}
		summaries, err := index.FindNotes(ctx, ports.NoteFilter{Type: "inbox", Limit: 100})
		if err != nil || len(summaries) != 1 || summaries[0].ID != entity.ID() || !reflect.DeepEqual(summaries[0].Tags, []string{"updated"}) {
			t.Fatalf("index did not update: %v, %v", summaries, err)
		}
		if _, err := application.Queries.GetInbox.Handle(ctx, inbox.GetQuery{ID: "bad-id"}); err == nil {
			t.Fatal("invalid lookup ID accepted")
		}
		if _, err := application.Commands.UpdateInbox.Handle(ctx, inbox.UpdateCommand{ID: uuid.New().String(), Content: body, Tags: []string{"updated"}}); err == nil {
			t.Fatal("update created a missing note")
		}
		if _, err := application.Commands.UpdateInbox.Handle(ctx, inbox.UpdateCommand{ID: entity.ID().String(), Tags: []string{"!!!"}}); err == nil {
			t.Fatal("invalid update accepted")
		}
		unchanged, err := application.Queries.GetInbox.Handle(ctx, inbox.GetQuery{ID: entity.ID().String()})
		if err != nil || !reflect.DeepEqual(unchanged.Inbox, entity) {
			t.Fatal("invalid update changed stored note")
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := application.Commands.UpdateInbox.Handle(cancelled, inbox.UpdateCommand{ID: entity.ID().String(), Content: body, Tags: []string{"updated"}}); !errors.Is(err, context.Canceled) {
			t.Fatal("cancelled update was not rejected")
		}
		cleared, err := application.Commands.UpdateInbox.Handle(ctx, inbox.UpdateCommand{ID: entity.ID().String(), Content: "", Tags: []string{"updated"}})
		if err != nil || cleared.Inbox.Content() != "" {
			t.Fatalf("empty content did not clear body: %v", err)
		}
		reloadedEmpty, err := application.Queries.GetInbox.Handle(ctx, inbox.GetQuery{ID: entity.ID().String()})
		if err != nil || reloadedEmpty.Inbox.Content() != "" {
			t.Fatal("cleared content reappeared on reload")
		}

	})

	t.Run("page", func(t *testing.T) {
		created, err := application.Commands.CreatePage.Handle(ctx, page.CreateCommand{Title: "Old page", Area: "work", Content: "Old content\n", Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		original := created.Page
		updated, err := application.Commands.UpdatePage.Handle(ctx, page.UpdateCommand{ID: original.ID().String(), Title: "New page", Area: "personal", Content: body, Tags: []string{"updated"}})
		if err != nil {
			t.Fatal(err)
		}
		entity := updated.Page
		if entity.ID() != original.ID() || !entity.Metadata().CreatedAt().Equal(original.Metadata().CreatedAt()) || !entity.Metadata().UpdatedAt().After(original.Metadata().UpdatedAt()) || !reflect.DeepEqual(entity.Tags().Strings(), []string{"updated"}) {
			t.Fatal("update changed identity/creation time or failed to update metadata/tags")
		}
		if entity.Content() != body || entity.Title().String() != "New page" || entity.Area().String() != "personal" || entity.Content() != body {
			t.Fatal("updated fields did not persist")
		}
		reloaded, err := application.Queries.GetPage.Handle(ctx, page.GetQuery{ID: original.ID().String()})
		if err != nil || !reflect.DeepEqual(reloaded.Page, entity) {
			t.Fatalf("round trip changed entity: %v", err)
		}
		stored, err := store.Get(ctx, original.ID())
		if err != nil || !bytes.HasSuffix(stored.Content, []byte(body)) {
			t.Fatalf("document content was lost: %s, %v", stored.Content, err)
		}
		summaries, err := index.FindNotes(ctx, ports.NoteFilter{Type: "page", Limit: 100})
		if err != nil || len(summaries) != 1 || summaries[0].ID != entity.ID() || !reflect.DeepEqual(summaries[0].Tags, []string{"updated"}) {
			t.Fatalf("index did not update: %v, %v", summaries, err)
		}
		if _, err := application.Queries.GetPage.Handle(ctx, page.GetQuery{ID: "bad-id"}); err == nil {
			t.Fatal("invalid lookup ID accepted")
		}
		if _, err := application.Commands.UpdatePage.Handle(ctx, page.UpdateCommand{ID: uuid.New().String(), Title: "New page", Area: "personal", Content: body, Tags: []string{"updated"}}); err == nil {
			t.Fatal("update created a missing note")
		}
		if _, err := application.Commands.UpdatePage.Handle(ctx, page.UpdateCommand{ID: entity.ID().String(), Tags: []string{"!!!"}}); err == nil {
			t.Fatal("invalid update accepted")
		}
		unchanged, err := application.Queries.GetPage.Handle(ctx, page.GetQuery{ID: entity.ID().String()})
		if err != nil || !reflect.DeepEqual(unchanged.Page, entity) {
			t.Fatal("invalid update changed stored note")
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := application.Commands.UpdatePage.Handle(cancelled, page.UpdateCommand{ID: entity.ID().String(), Title: "New page", Area: "personal", Content: body, Tags: []string{"updated"}}); !errors.Is(err, context.Canceled) {
			t.Fatal("cancelled update was not rejected")
		}
		cleared, err := application.Commands.UpdatePage.Handle(ctx, page.UpdateCommand{ID: entity.ID().String(), Title: "New page", Area: "personal", Content: "", Tags: []string{"updated"}})
		if err != nil || cleared.Page.Content() != "" {
			t.Fatalf("empty content did not clear body: %v", err)
		}
		reloadedEmpty, err := application.Queries.GetPage.Handle(ctx, page.GetQuery{ID: entity.ID().String()})
		if err != nil || reloadedEmpty.Page.Content() != "" {
			t.Fatal("cleared content reappeared on reload")
		}

	})

	t.Run("journal", func(t *testing.T) {
		created, err := application.Commands.CreateJournal.Handle(ctx, journal.CreateCommand{Date: "2020-01-02", Content: "Old day", Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		original := created.Journal
		updated, err := application.Commands.UpdateJournal.Handle(ctx, journal.UpdateCommand{ID: original.ID().String(), Content: body, Tags: []string{"updated"}})
		if err != nil {
			t.Fatal(err)
		}
		entity := updated.Journal
		if entity.ID() != original.ID() || !entity.Metadata().CreatedAt().Equal(original.Metadata().CreatedAt()) || !entity.Metadata().UpdatedAt().After(original.Metadata().UpdatedAt()) || !reflect.DeepEqual(entity.Tags().Strings(), []string{"updated"}) {
			t.Fatal("update changed identity/creation time or failed to update metadata/tags")
		}
		if entity.Date().String() != "2020-01-02" || entity.Content() != body {
			t.Fatal("updated fields did not persist")
		}
		reloaded, err := application.Queries.GetJournal.Handle(ctx, journal.GetQuery{ID: original.ID().String()})
		if err != nil || !reflect.DeepEqual(reloaded.Journal, entity) {
			t.Fatalf("round trip changed entity: %v", err)
		}
		stored, err := store.Get(ctx, original.ID())
		if err != nil || !bytes.HasSuffix(stored.Content, []byte(body)) {
			t.Fatalf("document content was lost: %s, %v", stored.Content, err)
		}
		summaries, err := index.FindNotes(ctx, ports.NoteFilter{Type: "journal", Limit: 100})
		if err != nil || len(summaries) != 1 || summaries[0].ID != entity.ID() || !reflect.DeepEqual(summaries[0].Tags, []string{"updated"}) {
			t.Fatalf("index did not update: %v, %v", summaries, err)
		}
		if _, err := application.Queries.GetJournal.Handle(ctx, journal.GetQuery{ID: "bad-id"}); err == nil {
			t.Fatal("invalid lookup ID accepted")
		}
		if _, err := application.Commands.UpdateJournal.Handle(ctx, journal.UpdateCommand{ID: uuid.New().String(), Content: body, Tags: []string{"updated"}}); err == nil {
			t.Fatal("update created a missing note")
		}
		if _, err := application.Commands.UpdateJournal.Handle(ctx, journal.UpdateCommand{ID: entity.ID().String(), Tags: []string{"!!!"}}); err == nil {
			t.Fatal("invalid update accepted")
		}
		unchanged, err := application.Queries.GetJournal.Handle(ctx, journal.GetQuery{ID: entity.ID().String()})
		if err != nil || !reflect.DeepEqual(unchanged.Journal, entity) {
			t.Fatal("invalid update changed stored note")
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := application.Commands.UpdateJournal.Handle(cancelled, journal.UpdateCommand{ID: entity.ID().String(), Content: body, Tags: []string{"updated"}}); !errors.Is(err, context.Canceled) {
			t.Fatal("cancelled update was not rejected")
		}
		cleared, err := application.Commands.UpdateJournal.Handle(ctx, journal.UpdateCommand{ID: entity.ID().String(), Content: "", Tags: []string{"updated"}})
		if err != nil || cleared.Journal.Content() != "" {
			t.Fatalf("empty content did not clear body: %v", err)
		}
		reloadedEmpty, err := application.Queries.GetJournal.Handle(ctx, journal.GetQuery{ID: entity.ID().String()})
		if err != nil || reloadedEmpty.Journal.Content() != "" {
			t.Fatal("cleared content reappeared on reload")
		}

	})
}
