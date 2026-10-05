package tui

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/inbox"
)

func openInbox(t *testing.T) model {
	t.Helper()
	next, _ := newModel().begin(actionInbox)
	m := next.(model)
	if m.screen != screenForm || !m.form.fields[0].textarea.Focused() {
		t.Fatal("Inbox did not open with content focused")
	}
	return m
}

func TestInboxCreationPreservesMultilineTextAndPassesTags(t *testing.T) {
	m := openInbox(t)
	calls := 0
	ctx := context.WithValue(context.Background(), struct{}{}, "inbox")
	m.ctx = ctx
	m.createInbox = func(gotCtx context.Context, command inbox.CreateCommand) (inbox.CreateResult, error) {
		calls++
		if gotCtx != ctx || command.Content != "q\nUna idea\notra línea" || !reflect.DeepEqual(command.Tags, []string{"ideas", "Trabajo personal"}) {
			t.Fatalf("unexpected command/context: %+v", command)
		}
		return inbox.CreateResult{}, nil
	}
	for _, message := range []tea.Msg{
		tea.KeyPressMsg{Code: 'q', Text: "q"},
		tea.KeyPressMsg{Code: tea.KeyEnter},
		tea.PasteMsg{Content: "Una idea\notra línea"},
		tea.KeyPressMsg{Code: tea.KeyTab},
		tea.PasteMsg{Content: " ideas, Trabajo personal "},
	} {
		next, _ := m.Update(message)
		m = next.(model)
	}
	if calls != 0 || m.screen != screenForm {
		t.Fatal("typing or Enter submitted the form")
	}
	next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = next.(model)
	if submit == nil {
		t.Fatal("Ctrl+S did not submit")
	}
	message := submit()
	next, save := m.Update(message)
	m = next.(model)
	if save == nil || !m.form.saving || calls != 0 {
		t.Fatal("creation should run asynchronously")
	}
	for _, blocked := range []tea.Msg{message, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}, tea.KeyPressMsg{Code: tea.KeyEscape}, formCancelled{}} {
		next, cmd := m.Update(blocked)
		m = next.(model)
		if cmd != nil || m.screen != screenForm || !m.form.saving {
			t.Fatal("saving must block duplicate submits and cancellation")
		}
	}
	if !strings.Contains(m.View().Content, "Saving…") {
		t.Fatal("missing saving status")
	}
	next, _ = m.Update(save())
	m = next.(model)
	if calls != 1 || m.screen != screenMenu || len(m.form.fields) != 0 || m.notification.text != "Inbox created" {
		t.Fatal("successful creation did not return to menu with notification")
	}
}

func TestInboxFailureKeepsValuesAndAllowsRetry(t *testing.T) {
	m := openInbox(t)
	m.form.fields[0].textarea.SetValue("Keep\nthis")
	m.form.fields[1].input.SetValue("ideas")
	calls := 0
	m.createInbox = func(_ context.Context, command inbox.CreateCommand) (inbox.CreateResult, error) {
		calls++
		if command.Content != "Keep\nthis" || !reflect.DeepEqual(command.Tags, []string{"ideas"}) {
			t.Fatal("retry lost values")
		}
		if calls == 1 {
			return inbox.CreateResult{}, errors.New("cannot save")
		}
		return inbox.CreateResult{}, nil
	}
	for attempt := range 2 {
		next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
		m = next.(model)
		next, save := m.Update(submit())
		m = next.(model)
		next, _ = m.Update(save())
		m = next.(model)
		if attempt == 0 && (m.screen != screenForm || m.form.saving || !strings.Contains(m.View().Content, "cannot save") || m.form.fields[0].Value() != "Keep\nthis") {
			t.Fatal("error did not preserve editable form")
		}
	}
	if calls != 2 || m.screen != screenMenu {
		t.Fatal("retry did not complete")
	}
}

func TestInboxCancelDiscardsWithoutSaving(t *testing.T) {
	m := openInbox(t)
	m.form.fields[0].textarea.SetValue("Discard")
	m.createInbox = func(context.Context, inbox.CreateCommand) (inbox.CreateResult, error) {
		t.Fatal("cancel called creation")
		return inbox.CreateResult{}, nil
	}
	next, cancel := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	if cancel == nil {
		t.Fatal("cancel action ignored")
	}
	next, _ = m.Update(cancel())
	m = next.(model)
	if m.screen != screenMenu || len(m.form.fields) != 0 {
		t.Fatal("cancel did not discard form")
	}
	next, _ = m.begin(actionInbox)
	m = next.(model)
	if m.form.fields[0].Value() != "" {
		t.Fatal("reopened form retained discarded content")
	}
}

func TestInboxAllowsEmptyContentAndTagsAndHandlesMissingDependency(t *testing.T) {
	m := openInbox(t)
	next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = next.(model)
	if submit == nil {
		t.Fatal("empty content and tags should be valid")
	}
	next, save := m.Update(submit())
	m = next.(model)
	next, _ = m.Update(save())
	m = next.(model)
	if m.screen != screenForm || m.form.saving || !strings.Contains(m.form.problem, "not configured") {
		t.Fatal("missing dependency was not reported")
	}
}
