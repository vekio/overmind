package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestStartScreenCanQuitWithoutApplicationDependencies(t *testing.T) {
	m := newModel()
	if m.Init() != nil {
		t.Fatal("start screen should not run application operations")
	}
	for _, key := range []tea.KeyPressMsg{{Code: 'q'}, {Code: tea.KeyEscape}} {
		_, cmd := m.Update(key)
		if cmd == nil {
			t.Fatal("quit key ignored")
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatal("quit key did not close the TUI")
		}
	}
}

func TestMenuNavigationAndDummySelection(t *testing.T) {
	m := newModel()
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	m = next.(model)
	if m.selected != len(menuItems)-1 {
		t.Fatal("menu did not wrap upwards")
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = next.(model)
	if m.selected != 0 {
		t.Fatal("menu did not wrap downwards")
	}
	for _, item := range menuItems {
		if item.action == actionList || item.action == actionReindex {
			continue
		}
		next, cmd := m.begin(item.action)
		m = next.(model)
		if cmd == nil || m.screen != screenMenu || m.notification.text != "Demo: "+item.title+" selected" {
			t.Fatalf("dummy action %s did not notify", item.title)
		}
	}
}
