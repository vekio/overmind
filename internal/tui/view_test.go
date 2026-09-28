package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
)

func TestNotesViewKeepsControlsAboveNotification(t *testing.T) {
	m := newModel(context.Background(), &deletingClient{})
	m.screen = screenNotes
	m.setNotes([]app.ListedNote{{Kind: domain.NoteKindInbox, CreatedAt: time.Date(2026, 9, 28, 12, 30, 0, 0, time.Local)}})
	m.width, m.height = 100, 24
	m.resizeNotesTable()
	m.notify("Note deleted", notificationSuccess)
	content := m.View().Content
	footer := strings.LastIndex(content, "refresh")
	notification := strings.LastIndex(content, "Note deleted")
	if footer < 0 || notification < footer || !strings.Contains(content, "Capture 2026-09-28 12:30") {
		t.Fatalf("Notes view omitted controls, notification, or capture label: %q", content)
	}
	if !strings.Contains(content[footer:notification], "\n") {
		t.Fatal("notification was not placed below controls")
	}
}
