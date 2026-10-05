package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/notes"
	"github.com/vekio/overmind/internal/ports"
)

func TestFilterModalSearchesAsYouTypeAndPreservesFiltersOnClose(t *testing.T) {
	var received notes.ListQuery
	m := newApplicationModel(context.Background(), func(_ context.Context, query notes.ListQuery) (notes.ListResult, error) {
		received = query
		return notes.ListResult{Notes: []ports.NoteSummary{{Type: "habit", Label: "Beber agua"}}}, nil
	})
	m.screen = screenNotes
	m.offset = notesPageSize
	next, _ := m.Update(tea.KeyPressMsg{Code: 'f'})
	m = next.(model)
	if !m.filterOpen || !m.filters[0].Focused() || !strings.Contains(m.View().Content, "Filter notes") {
		t.Fatal("modal did not open or focus type")
	}
	// Set the prefix, then type the last character through the input update loop.
	m.filters[0].SetValue("habi")
	next, cmd := m.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	m = next.(model)
	if cmd == nil || !m.loading || m.offset != 0 || m.filters[0].Value() != "habit" {
		t.Fatalf("input did not schedule query: %+v", m)
	}
	// Consume the pending query directly, without sleeping for the debounce timer.
	next, cmd = m.Update(loadNotes{revision: m.revision, query: notes.ListQuery{Type: m.filters[0].Value(), Limit: notesPageSize + 1}})
	m = next.(model)
	next, _ = m.Update(cmd())
	m = next.(model)
	if received.Type != "habit" || m.noteCount != 1 || m.loading {
		t.Fatal("live result not applied")
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = next.(model)
	if m.filterFocus != 1 || !m.filters[1].Focused() || m.filters[0].Focused() {
		t.Fatal("tab did not focus tag")
	}
	m.filters[1].SetValue("salu")
	next, _ = m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	m = next.(model)
	next, cmd = m.Update(loadNotes{revision: m.revision, query: notes.ListQuery{Type: m.filters[0].Value(), Tag: m.filters[1].Value(), Limit: notesPageSize + 1}})
	m = next.(model)
	next, _ = m.Update(cmd())
	m = next.(model)
	if received.Type != "habit" || received.Tag != "salud" {
		t.Fatalf("query=%+v", received)
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	if m.filterOpen || m.screen != screenNotes || m.filters[1].Value() != "salud" {
		t.Fatal("escape must keep filters and return to table")
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: 'f'})
	m = next.(model)
	m.filters[1].SetValue("x")
	next, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	m = next.(model)
	if cmd == nil || m.filters[1].Value() != "" {
		t.Fatal("clearing tag did not schedule unfiltered query")
	}
}

func TestNoteQueriesIgnoreStaleResponsesAndTimers(t *testing.T) {
	m := newModel()
	m.screen = screenNotes
	m.requestNotes(0)
	first := m.revision
	m.requestNotes(0)
	next, cmd := m.Update(loadNotes{revision: first})
	m = next.(model)
	if cmd != nil {
		t.Fatal("stale timer started a query")
	}
	next, _ = m.Update(notesLoaded{revision: m.revision, result: notes.ListResult{Notes: []ports.NoteSummary{{Label: "Latest"}}}})
	m = next.(model)
	for _, stale := range []notesLoaded{
		{revision: first, result: notes.ListResult{Notes: []ports.NoteSummary{{Label: "Old"}}}},
		{revision: first, err: errors.New("old failure")},
	} {
		next, _ = m.Update(stale)
		m = next.(model)
		if m.noteCount != 1 || m.notes.Rows()[0][1] != "Latest" || m.problem != "" {
			t.Fatal("stale response replaced current results")
		}
	}
}

func TestNotesPaginationAndQueryFailure(t *testing.T) {
	m := newModel()
	m.screen = screenNotes
	m.revision = 1
	result := notes.ListResult{Notes: make([]ports.NoteSummary, notesPageSize+1)}
	next, _ := m.Update(notesLoaded{revision: 1, result: result})
	m = next.(model)
	if !m.hasMore || m.noteCount != notesPageSize {
		t.Fatal("page boundary not detected")
	}
	next, cmd := m.Update(tea.KeyPressMsg{Code: 'n'})
	m = next.(model)
	if cmd == nil || m.offset != notesPageSize {
		t.Fatal("next page not requested")
	}
	next, _ = m.Update(notesLoaded{revision: m.revision, result: notes.ListResult{}})
	m = next.(model)
	if m.hasMore || m.noteCount != 0 {
		t.Fatal("empty page not handled")
	}
	next, cmd = m.Update(tea.KeyPressMsg{Code: 'p'})
	m = next.(model)
	if cmd == nil || m.offset != 0 {
		t.Fatal("previous page not requested")
	}
	next, cmd = m.Update(notesLoaded{revision: m.revision, err: errors.New("index unavailable")})
	m = next.(model)
	if cmd == nil || m.notification.kind != notificationError || m.loading || m.problem == "" {
		t.Fatal("query error not reported")
	}
}
