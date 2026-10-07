package tui

import (
	"strings"
	"testing"
	"time"
)

func TestNotesViewKeepsControlsAboveNotification(t *testing.T) {
	m := newModel()
	m.screen = screenNotes
	m.setNotes([]noteRow{{kind: "inbox", name: "Una idea", updatedAt: time.Date(2026, 9, 28, 12, 30, 0, 0, time.Local)}})
	m.width, m.height = 100, 24
	m.resizeNotesTable()
	m.notify("Note deleted", notificationSuccess)
	content := m.View().Content
	footer := strings.LastIndex(content, "refresh")
	notification := strings.LastIndex(content, "Note deleted")
	if footer < 0 || notification < footer || !strings.Contains(content, "Una idea") {
		t.Fatalf("Notes view omitted controls, notification, or note label: %q", content)
	}
	if !strings.Contains(content[footer:notification], "\n") {
		t.Fatal("notification was not placed below controls")
	}
}

func TestMenuViewShowsSelectionControlsAndNotification(t *testing.T) {
	m := newModel()
	m.selected = 2
	m.notify("Habit created", notificationSuccess)
	content := m.View().Content
	if !strings.Contains(content, "> Habit") || !strings.Contains(content, "move") || !strings.Contains(content, "select") || !strings.Contains(content, "quit") {
		t.Fatalf("selector or footer missing: %q", content)
	}
	controls := strings.LastIndex(content, "quit")
	notification := strings.LastIndex(content, "Habit created")
	if notification < controls || !strings.Contains(content[controls:notification], "\n") {
		t.Fatal("notification must appear below the controls")
	}
}

func TestNotificationKeepsControlsInPlace(t *testing.T) {
	for _, screen := range []screen{screenMenu, screenNotes} {
		for _, height := range []int{16, 24, 40} {
			m := newModel()
			m.screen = screen
			m.width, m.height = 100, height
			m.setNotes([]noteRow{{kind: "inbox", name: "Una idea"}})
			before := m.View().Content
			m.notify("Note created", notificationSuccess)
			during := m.View().Content
			next, _ := m.Update(dismissNotification{id: m.notification.id})
			after := next.(model).View().Content
			controlRow := func(content string) int {
				for row, line := range strings.Split(content, "\n") {
					if strings.Contains(line, "move") {
						return row
					}
				}
				return -1
			}
			if row := controlRow(before); row < 0 || controlRow(during) != row || controlRow(after) != row {
				t.Fatalf("screen %d, height %d: controls moved when notification appeared or disappeared", screen, height)
			}
			if controlRow(before) != height-2 || len(strings.Split(before, "\n")) != height {
				t.Fatalf("screen %d, height %d: expected exactly one notification row below controls", screen, height)
			}
			if before != after || strings.Count(before, "\n") != strings.Count(during, "\n") {
				t.Fatalf("screen %d, height %d: notification changed the layout height or did not restore the view", screen, height)
			}
		}
	}
}

func TestFilterModalIsCenteredOverNotesAndKeepsFooter(t *testing.T) {
	m := newModel()
	m.screen = screenNotes
	m.filterOpen = true
	m.width, m.height = 100, 30
	m.setNotes([]noteRow{{kind: "habit", name: "Agua"}})
	content := m.View().Content
	if !strings.Contains(content, "Overmind / Notes") || !strings.Contains(content, "Filter notes") || !strings.Contains(content, "refresh") {
		t.Fatalf("modal obscured heading or footer: %q", content)
	}
	lines := strings.Split(content, "\n")
	found := false
	for row, line := range lines {
		if strings.Contains(line, "╭") {
			if row == 0 || strings.HasPrefix(line, "╭") {
				t.Fatal("modal must be centered, not drawn at origin")
			}
			found = true
		}
	}
	if !found {
		t.Fatal("modal border missing")
	}
}
