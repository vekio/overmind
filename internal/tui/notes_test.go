package tui

import (
	"context"
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/notes"
	"github.com/vekio/overmind/internal/ports"
)

func TestNotesQueryAndDummyActionsPreserveRows(t *testing.T) {
	m := newApplicationModel(context.Background(), func(context.Context, notes.ListQuery) (notes.ListResult, error) {
		return notes.ListResult{Notes: []ports.NoteSummary{{Type: "habit", Label: "Beber agua"}, {Type: "page", Label: "Plan"}}}, nil
	})
	next, cmd := m.begin(actionList)
	m = next.(model)
	next, cmd = m.Update(cmd())
	m = next.(model)
	next, _ = m.Update(cmd())
	m = next.(model)
	if m.screen != screenNotes || m.noteCount != 2 {
		t.Fatal("notes table did not show indexed rows")
	}
	before := m.notes.Rows()
	for _, key := range []rune{'e'} {
		next, cmd := m.Update(tea.KeyPressMsg{Code: key})
		m = next.(model)
		if cmd == nil || m.notification.text == "" || !reflect.DeepEqual(m.notes.Rows(), before) {
			t.Fatal("dummy table action altered rows or omitted notification")
		}
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = next.(model)
	if m.notes.Cursor() != 1 {
		t.Fatal("table did not navigate downwards")
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	if m.screen != screenMenu {
		t.Fatal("escape did not return to selector")
	}
}
