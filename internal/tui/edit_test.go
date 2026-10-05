package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"uuid"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/app/bookmark"
	"github.com/vekio/overmind/internal/app/habit"
	"github.com/vekio/overmind/internal/app/inbox"
	"github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/app/notes"
	"github.com/vekio/overmind/internal/app/page"
	"github.com/vekio/overmind/internal/app/person"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/infra/codecs"
	"github.com/vekio/overmind/internal/infra/idgenerator"
	noteindex "github.com/vekio/overmind/internal/infra/index"
	"github.com/vekio/overmind/internal/infra/notestore"
	"github.com/vekio/overmind/internal/infra/repositories"
	"github.com/vekio/overmind/internal/ports"
)

type editAppClient struct {
	application      *app.Application
	loadErr, saveErr error
	loads, saves     int
}

func (client *editAppClient) GetHabit(ctx context.Context, query habit.GetQuery) (habit.GetResult, error) {
	client.loads++
	if client.loadErr != nil {
		return habit.GetResult{}, client.loadErr
	}
	return client.application.Queries.GetHabit.Handle(ctx, query)
}
func (client *editAppClient) UpdateHabit(ctx context.Context, command habit.UpdateCommand) (habit.UpdateResult, error) {
	client.saves++
	if client.saveErr != nil {
		return habit.UpdateResult{}, client.saveErr
	}
	return client.application.Commands.UpdateHabit.Handle(ctx, command)
}
func (client *editAppClient) GetPerson(ctx context.Context, query person.GetQuery) (person.GetResult, error) {
	client.loads++
	if client.loadErr != nil {
		return person.GetResult{}, client.loadErr
	}
	return client.application.Queries.GetPerson.Handle(ctx, query)
}
func (client *editAppClient) UpdatePerson(ctx context.Context, command person.UpdateCommand) (person.UpdateResult, error) {
	client.saves++
	if client.saveErr != nil {
		return person.UpdateResult{}, client.saveErr
	}
	return client.application.Commands.UpdatePerson.Handle(ctx, command)
}
func (client *editAppClient) GetBookmark(ctx context.Context, query bookmark.GetQuery) (bookmark.GetResult, error) {
	client.loads++
	if client.loadErr != nil {
		return bookmark.GetResult{}, client.loadErr
	}
	return client.application.Queries.GetBookmark.Handle(ctx, query)
}
func (client *editAppClient) UpdateBookmark(ctx context.Context, command bookmark.UpdateCommand) (bookmark.UpdateResult, error) {
	client.saves++
	if client.saveErr != nil {
		return bookmark.UpdateResult{}, client.saveErr
	}
	return client.application.Commands.UpdateBookmark.Handle(ctx, command)
}
func (client *editAppClient) GetInbox(ctx context.Context, query inbox.GetQuery) (inbox.GetResult, error) {
	client.loads++
	if client.loadErr != nil {
		return inbox.GetResult{}, client.loadErr
	}
	return client.application.Queries.GetInbox.Handle(ctx, query)
}
func (client *editAppClient) UpdateInbox(ctx context.Context, command inbox.UpdateCommand) (inbox.UpdateResult, error) {
	client.saves++
	if client.saveErr != nil {
		return inbox.UpdateResult{}, client.saveErr
	}
	return client.application.Commands.UpdateInbox.Handle(ctx, command)
}
func (client *editAppClient) GetPage(ctx context.Context, query page.GetQuery) (page.GetResult, error) {
	client.loads++
	if client.loadErr != nil {
		return page.GetResult{}, client.loadErr
	}
	return client.application.Queries.GetPage.Handle(ctx, query)
}
func (client *editAppClient) UpdatePage(ctx context.Context, command page.UpdateCommand) (page.UpdateResult, error) {
	client.saves++
	if client.saveErr != nil {
		return page.UpdateResult{}, client.saveErr
	}
	return client.application.Commands.UpdatePage.Handle(ctx, command)
}
func (client *editAppClient) GetJournal(ctx context.Context, query journal.GetQuery) (journal.GetResult, error) {
	client.loads++
	if client.loadErr != nil {
		return journal.GetResult{}, client.loadErr
	}
	return client.application.Queries.GetJournal.Handle(ctx, query)
}
func (client *editAppClient) UpdateJournal(ctx context.Context, command journal.UpdateCommand) (journal.UpdateResult, error) {
	client.saves++
	if client.saveErr != nil {
		return journal.UpdateResult{}, client.saveErr
	}
	return client.application.Commands.UpdateJournal.Handle(ctx, command)
}

func editFixture(t *testing.T, kind string) (model, *editAppClient) {
	t.Helper()
	ctx := context.Background()
	index, err := noteindex.New(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = index.Close() })
	store := notestore.New(t.TempDir())
	application := app.New(app.Dependencies{
		IDGenerator: idgenerator.New(), Index: index,
		Habits:    repositories.NewHabitRepository(store, index, codecs.HabitCodec{}),
		Persons:   repositories.NewPersonRepository(store, index, codecs.PersonCodec{}),
		Bookmarks: repositories.NewBookmarkRepository(store, index, codecs.BookmarkCodec{}),
		Inbox:     repositories.NewInboxRepository(store, index, codecs.InboxCodec{}),
		Pages:     repositories.NewPageRepository(store, index, codecs.PageCodec{}),
		Journals:  repositories.NewJournalRepository(store, index, codecs.JournalCodec{}),
	})
	var entity interface {
		ID() uuid.UUID
		Metadata() shared.EntityMetadata
	}
	switch kind {
	case "habit":
		result, err := application.Commands.CreateHabit.Handle(ctx, habit.CreateCommand{Title: "Old habit", Amount: 1, Unit: "hours", Period: "week", Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		entity = result.Habit
	case "person":
		result, err := application.Commands.CreatePerson.Handle(ctx, person.CreateCommand{Name: "Old person", Groups: []string{"work"}, Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		entity = result.Person
	case "bookmark":
		result, err := application.Commands.CreateBookmark.Handle(ctx, bookmark.CreateCommand{URL: "https://example.com/old", Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		entity = result.Bookmark
	case "inbox":
		result, err := application.Commands.CreateInbox.Handle(ctx, inbox.CreateCommand{Content: "Old content", Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		entity = result.Inbox
	case "page":
		result, err := application.Commands.CreatePage.Handle(ctx, page.CreateCommand{Title: "Old page", Area: "work", Content: "Old content\n", Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		entity = result.Page
	case "journal":
		result, err := application.Commands.CreateJournal.Handle(ctx, journal.CreateCommand{Date: "2020-01-02", Content: "Old day", Tags: []string{"keep"}})
		if err != nil {
			t.Fatal(err)
		}
		entity = result.Journal
	}
	client := &editAppClient{application: application}
	m := newModel()
	m.newEditClient = func(context.Context) (EditClient, error) { return client, nil }
	m.listGroups = application.Queries.ListGroups.Handle
	m.screen = screenNotes
	m.setNotes([]noteRow{{id: uuid.New(), kind: kind, name: "Other"}, {id: entity.ID(), kind: kind, name: "Selected"}})
	m.notes.SetCursor(1)
	m.offset = notesPageSize
	m.filters[0].SetValue(kind)
	m.filters[1].SetValue("keep")
	return m, client
}

func TestEditAllTypesFromNotesAndRestoreSelection(t *testing.T) {
	for _, kind := range []string{"habit", "person", "bookmark", "inbox", "page", "journal"} {
		t.Run(kind, func(t *testing.T) {
			m, client := editFixture(t, kind)
			selectedID := m.rows[1].id
			calls := 0
			m.listNotes = func(_ context.Context, query notes.ListQuery) (notes.ListResult, error) {
				calls++
				if query.Type != kind || query.Tag != "keep" || query.Offset != notesPageSize {
					t.Fatalf("refresh changed filters or page: %+v", query)
				}
				return notes.ListResult{Notes: []ports.NoteSummary{{ID: selectedID, Type: kind, Label: "Updated", Tags: []string{"keep", "updated"}}}}, nil
			}
			next, load := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			m = next.(model)
			if load == nil || m.screen != screenEdit || client.loads != 0 {
				t.Fatal("edit did not load asynchronously")
			}
			next, _ = m.Update(load())
			m = next.(model)
			if m.screen != screenForm || m.editID != selectedID || client.loads != 1 || !strings.Contains(m.form.detail, "Created") || strings.Contains(m.View().Content, selectedID.String()) {
				t.Fatal("edit did not populate form or expose subtle metadata")
			}
			for _, field := range m.form.fields {
				if field.spec.id == "tags" && field.Value() != "keep" {
					t.Fatal("existing tags were lost")
				}
			}
			if kind == "journal" && (m.form.fields[0].Value() != "2020-01-02" || m.form.fields[0].spec.kind != fieldStatic) {
				t.Fatal("Journal edit must use the selected fixed date")
			}

			setFormValue(&m.form, "content", "Changed content\nSecond line\n")
			setFormValue(&m.form, "tags", "keep, updated")
			next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			m = next.(model)
			next, save := m.Update(submit())
			m = next.(model)
			if save == nil || !m.form.saving || client.saves != 0 {
				t.Fatal("update was not asynchronous")
			}
			next, blocked := m.Update(formCancelled{})
			m = next.(model)
			if blocked != nil || !m.form.saving {
				t.Fatal("cancel interrupted saving")
			}
			next, refresh := m.Update(save())
			m = next.(model)
			if m.screen != screenNotes || m.editID != uuid.Nil() || !m.loading || client.saves != 1 || refresh == nil {
				t.Fatal("successful update did not return to Notes and refresh")
			}
			batch := refresh().(tea.BatchMsg)
			next, fetch := m.Update(batch[0]())
			m = next.(model)
			next, _ = m.Update(fetch())
			m = next.(model)
			if calls != 1 || m.notes.Cursor() != 0 || m.rows[0].id != selectedID || m.offset != notesPageSize {
				t.Fatal("refresh did not preserve selected identity or page")
			}
			m.editID = selectedID
			m.editRevision++
			loaded := m.loadEdit(kind)
			result := loaded().(editLoaded)
			if result.err != nil {
				t.Fatal(result.err)
			}
			for _, field := range result.form.fields {
				if field.spec.id == "tags" && field.Value() != "keep, updated" {
					t.Fatal("update did not persist tags")
				}
				if field.spec.id == "content" && !strings.HasPrefix(field.Value(), "Changed") {
					t.Fatal("updated content was lost")
				}
			}
		})
	}
}

func TestEditErrorsRetryCancelAndStaleLoads(t *testing.T) {
	m, client := editFixture(t, "page")
	client.loadErr = errors.New("cannot load")
	next, load := m.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
	m = next.(model)
	next, _ = m.Update(load())
	m = next.(model)
	if m.screen != screenNotes || !strings.Contains(m.notification.text, "cannot load") || m.notes.Cursor() != 1 {
		t.Fatal("load failure lost Notes state")
	}
	client.loadErr = nil
	next, load = m.beginEdit()
	m = next.(model)
	oldLoad := load()
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	next, _ = m.Update(oldLoad)
	m = next.(model)
	if m.screen != screenNotes {
		t.Fatal("cancelled load reopened editor")
	}
	next, load = m.beginEdit()
	m = next.(model)
	next, _ = m.Update(oldLoad)
	m = next.(model)
	if m.screen != screenEdit {
		t.Fatal("stale load changed reopened editor")
	}
	next, _ = m.Update(load())
	m = next.(model)
	setFormValue(&m.form, "content", "Draft")
	client.saveErr = errors.New("cannot save")
	next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = next.(model)
	next, save := m.Update(submit())
	m = next.(model)
	next, _ = m.Update(save())
	m = next.(model)
	if m.screen != screenForm || m.form.saving || !strings.Contains(m.form.problem, "cannot save") || m.form.fields[2].Value() != "Draft" {
		t.Fatal("update failure lost editable draft")
	}
	next, cancel := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	next, _ = m.Update(cancel())
	m = next.(model)
	if m.screen != screenNotes || m.notes.Cursor() != 1 || len(m.form.fields) != 0 {
		t.Fatal("cancel did not return to unchanged Notes selection")
	}
	next, load = m.beginEdit()
	m = next.(model)
	next, _ = m.Update(load())
	m = next.(model)
	if m.form.fields[2].Value() != "Old content\n" {
		t.Fatal("cancel saved discarded draft")
	}
	client.saveErr = nil
	next, submit = m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = next.(model)
	next, save = m.Update(submit())
	m = next.(model)
	next, _ = m.Update(save())
	if next.(model).screen != screenNotes {
		t.Fatal("retry did not complete")
	}
}

func TestPageAndJournalShareLargerContentSize(t *testing.T) {
	for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 24}, {Width: 100, Height: 40}, {Width: 45, Height: 22}} {
		pageForm, journalForm := newPageForm(), newJournalForm("2020-01-02")
		pageForm.focus = 2
		journalForm.focus = 1
		pageForm.detail, journalForm.detail = "Created yesterday · Updated today", "Created yesterday · Updated today"
		pageForm.Resize(size.Width, size.Height)
		journalForm.Resize(size.Width, size.Height)
		contentHeight := pageForm.fields[2].textarea.Height()
		pageForm.focus = 0
		pageForm.Resize(size.Width, size.Height)
		if pageForm.fields[2].textarea.Height() != contentHeight {
			t.Fatal("content height depended on focused field")
		}
		pageForm.focus = 2
		if pageForm.fields[2].textarea.Height() != journalForm.fields[1].textarea.Height() || pageForm.fields[2].textarea.Height() < 1 {
			t.Fatal("Page/Journal content sizes differ or were not increased")
		}
		for _, f := range []form{pageForm, journalForm} {
			f.Focus()
			view := ansi.Strip(f.View())
			if strings.LastIndex(view, "Tags >") < strings.Index(view, "Content") || lipgloss.Height(view)+lipgloss.Height(ansi.Wrap(formatControls(f.Controls()), size.Width, "")) > size.Height {
				t.Fatalf("expanded form omitted tags ordering or exceeded terminal height: %s", view)
			}
		}
	}
}
