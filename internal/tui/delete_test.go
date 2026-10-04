package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"uuid"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app"
)

func TestDeleteConfirmationUsesSelectedIdentityAndKeepsFilters(t *testing.T) {
	first, selected := uuid.New(), uuid.New()
	calls := 0
	m := newModel()
	m.screen = screenNotes
	m.filters[0].SetValue("habit")
	m.filters[1].SetValue("salud")
	m.setNotes([]noteRow{{id: first, kind: "habit", name: "First"}, {id: selected, kind: "habit", name: "Selected"}})
	m.notes.SetCursor(1)
	m.deleteNote = func(_ context.Context, command app.DeleteNoteCommand) (app.DeleteNoteResult, error) {
		calls++
		if command.ID != selected.String() {
			t.Fatalf("deleted wrong note: %s", command.ID)
		}
		return app.DeleteNoteResult{ID: selected}, nil
	}
	next, _ := m.Update(tea.KeyPressMsg{Code: 'd'})
	m = next.(model)
	if m.screen != screenAsk || m.ask.options[m.ask.selected].id != cancelOption || !strings.Contains(m.View().Content, selected.String()) {
		t.Fatal("confirmation did not identify selected note with cancel default")
	}
	next, cmd := m.Update(askAnswer{option: cancelOption})
	m = next.(model)
	if cmd != nil || m.screen != screenNotes || calls != 0 || m.noteCount != 2 {
		t.Fatal("cancel changed notes")
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: 'd'})
	m = next.(model)
	next, cmd = m.Update(askAnswer{option: deleteOption})
	m = next.(model)
	if !m.deleting || cmd == nil || m.screen != screenNotes {
		t.Fatal("confirmed delete was not asynchronous")
	}
	next, duplicate := m.Update(askAnswer{option: deleteOption})
	m = next.(model)
	if duplicate != nil {
		t.Fatal("duplicate confirmation started another deletion")
	}
	next, refresh := m.Update(cmd())
	m = next.(model)
	if calls != 1 || m.deleting || !m.loading || refresh == nil || m.notification.text != "Note deleted" || m.filters[0].Value() != "habit" || m.filters[1].Value() != "salud" {
		t.Fatal("delete did not refresh using existing filters")
	}
}

func TestDeleteEscapeAndFailureKeepTable(t *testing.T) {
	id := uuid.New()
	m := newModel()
	m.screen = screenNotes
	m.setNotes([]noteRow{{id: id, name: "Keep me"}})
	m.deleteNote = func(context.Context, app.DeleteNoteCommand) (app.DeleteNoteResult, error) {
		return app.DeleteNoteResult{}, errors.New("cannot delete")
	}
	next, _ := m.Update(tea.KeyPressMsg{Code: 'd'})
	m = next.(model)
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	if m.screen != screenNotes || m.deleting {
		t.Fatal("escape must return to table without deleting")
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: 'd'})
	m = next.(model)
	next, cmd := m.Update(askAnswer{option: deleteOption})
	m = next.(model)
	next, _ = m.Update(cmd())
	m = next.(model)
	if m.deleting || m.notification.kind != notificationError || m.noteCount != 1 {
		t.Fatal("failure did not preserve table and notify")
	}
}

func TestDeleteLastNoteOnPageReturnsToPreviousPage(t *testing.T) {
	id := uuid.New()
	m := newModel()
	m.screen = screenNotes
	m.offset = notesPageSize
	m.deleting = true
	m.deleteID = id
	m.setNotes([]noteRow{{id: id}})
	next, cmd := m.Update(deleteFinished{id: id})
	m = next.(model)
	if m.offset != 0 || cmd == nil || !m.loading {
		t.Fatal("last note deletion did not reload previous page")
	}
}

func TestDeleteEmptySelectionDoesNotAskOrExecute(t *testing.T) {
	m := newModel()
	m.screen = screenNotes
	next, cmd := m.Update(tea.KeyPressMsg{Code: 'd'})
	m = next.(model)
	if m.screen != screenNotes || m.deleting || cmd == nil || m.notification.text != "Select a note to delete" {
		t.Fatal("empty selection was not handled")
	}
}
