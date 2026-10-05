package tui

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
	"uuid"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/app/notes"
	"github.com/vekio/overmind/internal/domain/calendar"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/domain/shared"
	"github.com/vekio/overmind/internal/ports"
)

func journalEntity(t *testing.T, date string) *journals.Journal {
	t.Helper()
	day, err := calendar.NewDate(date)
	if err != nil {
		t.Fatal(err)
	}
	tag, _ := shared.NewTag("daily")
	tags, _ := shared.NewTags(tag)
	metadata, _ := shared.NewEntityMetadata(time.Now(), time.Now())
	entity, err := journals.NewJournal(uuid.New(), day, "", tags, metadata)
	if err != nil {
		t.Fatal(err)
	}
	return entity
}

func TestJournalTodayIsFixedAndCreationOpensSavedJournal(t *testing.T) {
	m := newModel()
	m.listNotes = func(context.Context, notes.ListQuery) (notes.ListResult, error) {
		return notes.ListResult{}, nil
	}
	before := time.Now().Format(time.DateOnly)
	next, lookup := m.begin(actionJournal)
	m = next.(model)
	if m.journalDate != before && m.journalDate != time.Now().Format(time.DateOnly) {
		t.Fatal("Journal did not use today's local date")
	}
	next, _ = m.Update(lookup())
	m = next.(model)
	if m.screen != screenForm || m.form.focus != 1 || !m.form.fields[1].textarea.Focused() {
		t.Fatal("new Journal should focus content and skip the fixed date")
	}
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "Date > "+m.journalDate) || strings.Contains(view, "YYYY-MM-DD") {
		t.Fatal("date should display today's value, not an editable placeholder")
	}
	for i, key := range []tea.KeyPressMsg{{Code: tea.KeyTab}, {Code: tea.KeyTab, Mod: tea.ModShift}} {
		next, _ = m.Update(key)
		m = next.(model)
		if m.form.focus != []int{2, 1}[i] || m.form.fields[0].Value() != m.journalDate {
			t.Fatal("keyboard navigation focused or changed the fixed date")
		}
	}
	next, _ = m.Update(tea.PasteMsg{Content: "My day\nA second line"})
	m = next.(model)
	m.form.fields[2].input.SetValue("daily")
	ctx := context.WithValue(context.Background(), struct{}{}, "journal")
	m.ctx = ctx
	entity := journalEntity(t, m.journalDate)
	entity.Rewrite("My day\nA second line", time.Now())
	calls := 0
	m.createJournal = func(gotCtx context.Context, command journal.CreateCommand) (journal.CreateResult, error) {
		calls++
		if gotCtx != ctx || command.Date != m.journalDate || command.Content != "My day\nA second line" || !reflect.DeepEqual(command.Tags, []string{"daily"}) {
			t.Fatalf("unexpected creation command/context: %+v", command)
		}
		return journal.CreateResult{Journal: entity}, nil
	}
	next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = next.(model)
	message := submit().(formSubmitted)
	message.values["date"] = "2000-01-01"
	next, save := m.Update(message)
	m = next.(model)
	if calls != 0 || !m.form.saving {
		t.Fatal("creation should run asynchronously")
	}
	next, _ = m.Update(save())
	m = next.(model)
	if calls != 1 || m.screen != screenForm || m.journal.ID() != entity.ID() || m.notification.text != "Journal created" || strings.Contains(m.View().Content, entity.ID().String()) || m.form.fields[1].Value() != entity.Content() {
		t.Fatal("created Journal was not shown")
	}
}

func TestExistingJournalIsShownWithoutCreationAcrossPages(t *testing.T) {
	m := newModel()
	calls := 0
	today := time.Now().Format(time.DateOnly)
	entity := journalEntity(t, today)
	entity.Rewrite("Existing\nentry", time.Now())
	note := ports.NoteSummary{ID: entity.ID(), Type: "journal", Label: today, Tags: entity.Tags().Strings(), UpdatedAt: time.Now()}
	m.getJournal = func(_ context.Context, query journal.GetQuery) (journal.GetResult, error) {
		if query.ID != entity.ID().String() {
			t.Fatal("loaded the wrong journal")
		}
		return journal.GetResult{Journal: entity}, nil
	}
	m.listNotes = func(_ context.Context, query notes.ListQuery) (notes.ListResult, error) {
		calls++
		if query.Type != "journal" || query.Tag != "" || query.Limit != notesPageSize {
			t.Fatalf("incorrect lookup filters: %+v", query)
		}
		if query.Offset == 0 {
			page := make([]ports.NoteSummary, notesPageSize)
			for i := range page {
				page[i] = ports.NoteSummary{Type: "journal", Label: "2000-01-01"}
			}
			return notes.ListResult{Notes: page}, nil
		}
		if query.Offset != notesPageSize {
			t.Fatal("incorrect pagination")
		}
		return notes.ListResult{Notes: []ports.NoteSummary{note}}, nil
	}
	m.createJournal = func(context.Context, journal.CreateCommand) (journal.CreateResult, error) {
		t.Fatal("existing Journal must not be recreated")
		return journal.CreateResult{}, nil
	}
	m.filters[1].SetValue("unrelated")
	next, lookup := m.begin(actionJournal)
	m = next.(model)
	next, _ = m.Update(lookup())
	m = next.(model)
	if calls != 2 || m.screen != screenForm || m.journal.ID() != note.ID || m.form.fields[1].Value() != entity.Content() {
		t.Fatal("existing Journal was not shown with its original tags")
	}
	view := ansi.Strip(m.View().Content)
	if strings.Contains(view, entity.ID().String()) || strings.Contains(view, "ID >") || strings.Contains(view, "Updated >") || !strings.Contains(view, "Created ") || !strings.Contains(view, "Updated ") {
		t.Fatal("timestamps should appear as secondary metadata without the ID")
	}
	m.updateJournal = func(_ context.Context, command journal.UpdateCommand) (journal.UpdateResult, error) {
		if command.ID != entity.ID().String() || command.Content != "Existing\nentry\nMore writing" {
			t.Fatalf("unexpected update command: %+v", command)
		}
		metadata, _ := entity.Metadata().Updated(time.Now())
		updated, _ := journals.NewJournal(entity.ID(), entity.Date(), command.Content, entity.Tags(), metadata)
		return journal.UpdateResult{Journal: updated}, nil
	}
	next, _ = m.Update(tea.PasteMsg{Content: "\nMore writing"})
	m = next.(model)
	next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = next.(model)
	next, save := m.Update(submit())
	m = next.(model)
	next, _ = m.Update(save())
	m = next.(model)
	if m.journal.ID() != entity.ID() || m.journal.Content() != "Existing\nentry\nMore writing" || m.notification.text != "Journal updated" {
		t.Fatal("editing did not update the existing Journal")
	}
}

func TestJournalConcurrentCreationOpensExistingJournal(t *testing.T) {
	m := newModel()
	entity := journalEntity(t, time.Now().Format(time.DateOnly))
	note := ports.NoteSummary{ID: entity.ID(), Type: "journal", Label: entity.Date().String(), Tags: entity.Tags().Strings(), UpdatedAt: time.Now()}
	m.getJournal = func(context.Context, journal.GetQuery) (journal.GetResult, error) {
		return journal.GetResult{Journal: entity}, nil
	}
	lookups := 0
	m.listNotes = func(context.Context, notes.ListQuery) (notes.ListResult, error) {
		lookups++
		if lookups == 1 {
			return notes.ListResult{}, nil
		}
		return notes.ListResult{Notes: []ports.NoteSummary{note}}, nil
	}
	m.createJournal = func(context.Context, journal.CreateCommand) (journal.CreateResult, error) {
		return journal.CreateResult{}, fmt.Errorf("save journal: %w", journal.ErrAlreadyExists)
	}
	next, lookup := m.begin(actionJournal)
	m = next.(model)
	next, _ = m.Update(lookup())
	m = next.(model)
	m.form.fields[2].input.SetValue("new tags")
	m.form.fields[1].textarea.SetValue("Preserve this draft")
	next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = next.(model)
	next, save := m.Update(submit())
	m = next.(model)
	next, lookup = m.Update(save())
	m = next.(model)
	if lookup == nil || !m.journalLoading {
		t.Fatal("duplicate must reload today's existing Journal")
	}
	next, _ = m.Update(lookup())
	m = next.(model)
	if m.screen != screenForm || m.journal.ID() != note.ID || m.journal.Content() != "" || m.form.fields[1].Value() != "Preserve this draft" || !strings.Contains(m.form.problem, "draft is preserved") {
		t.Fatal("concurrent creation did not show the existing Journal without replacing tags")
	}
}

func TestJournalLookupFailureRetryAndStaleResult(t *testing.T) {
	m := newModel()
	m.listNotes = func(context.Context, notes.ListQuery) (notes.ListResult, error) {
		return notes.ListResult{}, errors.New("cannot load journal")
	}
	next, lookup := m.begin(actionJournal)
	m = next.(model)
	next, _ = m.Update(lookup())
	m = next.(model)
	if m.screen != screenJournal || m.journalLoading || !strings.Contains(m.View().Content, "cannot load journal") {
		t.Fatal("lookup failure must be shown without opening creation")
	}
	m.listNotes = func(context.Context, notes.ListQuery) (notes.ListResult, error) {
		return notes.ListResult{}, nil
	}
	next, retry := m.Update(tea.KeyPressMsg{Code: 'r'})
	m = next.(model)
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	next, newer := m.begin(actionJournal)
	m = next.(model)
	next, cmd := m.Update(retry())
	m = next.(model)
	if cmd != nil || m.screen != screenJournal || !m.journalLoading {
		t.Fatal("stale lookup changed the newly opened Journal")
	}
	next, _ = m.Update(newer())
	m = next.(model)
	if m.screen != screenForm {
		t.Fatal("latest lookup did not open today's creation form")
	}
}
