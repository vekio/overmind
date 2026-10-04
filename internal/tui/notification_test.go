package tui

import (
	"testing"
)

func TestOlderNotificationTimerDoesNotDismissNewerMessage(t *testing.T) {
	m := newModel()
	m.notify("first", notificationInfo)
	oldID := m.notification.id
	m.notify("second", notificationSuccess)
	next, _ := m.Update(dismissNotification{id: oldID})
	m = next.(model)
	if m.notification.text != "second" {
		t.Fatalf("old timer dismissed newer notification: %q", m.notification.text)
	}
	next, _ = m.Update(dismissNotification{id: m.notification.id})
	m = next.(model)
	if m.notification.text != "" {
		t.Fatalf("current timer did not dismiss notification: %q", m.notification.text)
	}
}
