package tui

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/vekio/overmind/internal/app/person"
)

func TestGroupsSearchSelectCreateAndRemove(t *testing.T) {
	f := newPersonForm()
	f.Resize(80, 24)
	f.fields[0].input.SetValue("Person")
	f.fields[1].groups.options = []string{"friends", "work"}
	f.focus = 1
	f.Focus()
	update := func(message tea.Msg) { f, _ = f.Update(message) }
	update(tea.PasteMsg{Content: "Wor"})
	if choices := f.fields[1].groupChoices(); len(choices) != 2 || choices[0].name != "work" {
		t.Fatalf("unexpected filtered choices: %v", choices)
	}
	update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if f.fields[1].Value() != "work" || f.fields[1].input.Value() != "" {
		t.Fatal("selection did not commit and clear search")
	}
	update(tea.PasteMsg{Content: "New Group"})
	view := ansi.Strip(f.View())
	if !strings.Contains(view, "+ Add new-group") || strings.Index(view, "Groups >") > strings.Index(view, "Tags >") {
		t.Fatalf("missing add choice or incorrect field order: %s", view)
	}
	update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !reflect.DeepEqual(f.fields[1].groups.chosen, []string{"work", "new-group"}) {
		t.Fatal("new group was not normalized and added")
	}
	update(tea.PasteMsg{Content: "WORK"})
	if choices := f.fields[1].groupChoices(); len(choices) != 1 || choices[0].create {
		t.Fatal("normalized existing group offered duplicate creation")
	}
	update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if f.fields[1].Value() != "new-group" {
		t.Fatal("selected group was not removed")
	}
	update(tea.PasteMsg{Content: "draft"})
	_, cmd := f.submit()
	if cmd != nil {
		if _, ok := cmd().(formSubmitted); ok {
			t.Fatal("unconfirmed search silently submitted")
		}
	}
	if f.fields[1].problem == "" {
		t.Fatal("unconfirmed search did not receive an explanation")
	}
	update(tea.KeyPressMsg{Code: tea.KeyEnter})
	_, cmd = f.submit()
	if cmd == nil || cmd().(formSubmitted).values["groups"] != "new-group, draft" {
		t.Fatal("selected groups not included in submission")
	}
}

func TestGroupsLoadingRetryAndStaleResults(t *testing.T) {
	m := newModel()
	ctx := context.WithValue(context.Background(), struct{}{}, "groups")
	m.ctx = ctx
	calls := 0
	m.listGroups = func(gotCtx context.Context, _ person.ListGroupsQuery) (person.ListGroupsResult, error) {
		calls++
		if gotCtx != ctx {
			t.Fatal("lookup lost context")
		}
		if calls == 1 {
			return person.ListGroupsResult{}, errors.New("lookup failed")
		}
		return person.ListGroupsResult{Groups: []string{"friends", "work"}}, nil
	}
	next, _ := m.begin(actionPerson)
	m = next.(model)
	// Obtain a single lookup command independently from the initial Batch.
	load := m.requestGroups()
	next, _ = m.Update(load())
	m = next.(model)
	if m.form.fields[1].groups.loading || !strings.Contains(m.form.fields[1].groups.problem, "lookup failed") {
		t.Fatal("lookup failure was not displayed")
	}
	m.form.fields[0].input.SetValue("Draft")
	m.form.fields[1].groups.chosen = []string{"new-group"}
	m.form.focus = 1
	m.form.Focus()
	next, retry := m.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	m = next.(model)
	if retry == nil || !m.form.fields[1].groups.loading {
		t.Fatal("retry did not start")
	}
	next, _ = m.Update(retry())
	m = next.(model)
	if !reflect.DeepEqual(m.form.fields[1].groups.options, []string{"friends", "work"}) || m.form.fields[1].Value() != "new-group" || m.form.fields[0].Value() != "Draft" {
		t.Fatal("lookup lost draft or failed to load choices")
	}
	oldRevision := m.groupsRevision
	next, _ = m.begin(actionPerson)
	m = next.(model)
	next, _ = m.Update(groupsLoaded{revision: oldRevision, groups: []string{"stale"}})
	m = next.(model)
	if len(m.form.fields[1].groups.options) != 0 {
		t.Fatal("stale lookup affected reopened form")
	}
	next, cancel := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	next, _ = m.Update(cancel())
	m = next.(model)
	next, _ = m.Update(groupsLoaded{revision: m.groupsRevision, groups: []string{"late"}})
	if next.(model).screen != screenMenu || len(next.(model).form.fields) != 0 {
		t.Fatal("late lookup reopened cancelled form")
	}
}

func TestGroupsChoicesScrollAndSearchKeepsVimLetters(t *testing.T) {
	f := newPersonForm()
	f.focus = 1
	f.Focus()
	f.fields[1].groups.options = []string{"a", "b", "c", "d", "e", "f"}
	for range 5 {
		f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	if view := ansi.Strip(f.View()); !strings.Contains(view, "> [ ] f") || !strings.Contains(view, "3–6 of 6") {
		t.Fatalf("choices did not scroll: %s", view)
	}
	f, _ = f.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if f.fields[1].input.Value() != "j" {
		t.Fatal("Vim letters must remain available for searching group names")
	}
}
