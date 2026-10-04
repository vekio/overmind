package tui

import (
	"context"
	"errors"
	"testing"

	"github.com/vekio/overmind/internal/app"
)

func TestReindexConfirmationRunsUseCaseAndReportsResult(t *testing.T) {
	m := newModel()
	calls := 0
	m.reindex = func(context.Context, app.ReindexCommand) (app.ReindexResult, error) {
		calls++
		return app.ReindexResult{Indexed: 6}, nil
	}
	next, _ := m.begin(actionReindex)
	m = next.(model)
	if m.screen != screenAsk || m.ask.options[m.ask.selected].id != cancelOption {
		t.Fatal("confirmation missing or incorrect default")
	}
	next, cmd := m.Update(askAnswer{option: cancelOption})
	m = next.(model)
	if cmd != nil || calls != 0 || m.reindexing {
		t.Fatal("cancel started reindex")
	}
	next, _ = m.begin(actionReindex)
	m = next.(model)
	next, cmd = m.Update(askAnswer{option: reindexOption})
	m = next.(model)
	if m.screen != screenMenu || cmd == nil || !m.reindexing {
		t.Fatal("reindex did not start asynchronously")
	}
	// Duplicate confirmation messages must not run a second reconstruction.
	next, duplicate := m.Update(askAnswer{option: reindexOption})
	m = next.(model)
	if duplicate != nil {
		t.Fatal("duplicate reindex started")
	}
	next, notification := m.Update(cmd())
	m = next.(model)
	if calls != 1 || m.reindexing || notification == nil || m.notification.kind != notificationSuccess || m.notification.text != "Indexed 6 notes" {
		t.Fatal("reindex result missing")
	}
}

func TestReindexErrorAppearsAsNotification(t *testing.T) {
	m := newModel()
	m.reindex = func(context.Context, app.ReindexCommand) (app.ReindexResult, error) {
		return app.ReindexResult{}, errors.New("invalid note")
	}
	next, cmd := m.Update(askAnswer{option: reindexOption})
	m = next.(model)
	next, _ = m.Update(cmd())
	m = next.(model)
	if m.reindexing || m.notification.kind != notificationError || m.notification.text != "invalid note" {
		t.Fatal("reindex error missing")
	}
}
