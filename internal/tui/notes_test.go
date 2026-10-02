package tui

import (
	"context"
	"testing"
	"time"
	"uuid"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/domain"
)

type deletingClient struct {
	Client
	deleted []uuid.UUID
}

func (client *deletingClient) DeleteNote(_ context.Context, id uuid.UUID) error {
	client.deleted = append(client.deleted, id)
	return nil
}

func (client *deletingClient) ListNotes(context.Context) ([]app.ListedNote, error) {
	return nil, nil
}

func TestNotesDeleteRequiresConfirmation(t *testing.T) {
	client := &deletingClient{}
	m := newModel(context.Background(), client)
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	m.setNotes([]app.ListedNote{{ID: id, Kind: domain.NoteKindInbox, CreatedAt: time.Now()}})
	m.screen = screenNotes

	next, _ := m.Update(tea.KeyPressMsg{Code: 'd'})
	m = next.(model)
	if m.screen != screenAsk || m.deleteNoteID != id || m.ask.options[m.ask.selected].id != cancelOption {
		t.Fatal("delete confirmation did not default to cancel")
	}
	next, _ = m.Update(askAnswer{option: cancelOption})
	m = next.(model)
	if m.screen != screenNotes || len(client.deleted) != 0 {
		t.Fatal("cancelled delete changed the note")
	}

	next, _ = m.Update(tea.KeyPressMsg{Code: 'd'})
	m = next.(model)
	next, command := m.Update(askAnswer{option: deleteOption})
	m = next.(model)
	if m.screen != screenBusy || command == nil {
		t.Fatal("confirmed delete did not start")
	}
	if result, ok := command().(noteDeleteResult); !ok || result.err != nil {
		t.Fatalf("delete result = %#v", result)
	}
	if len(client.deleted) != 1 || client.deleted[0] != id {
		t.Fatalf("deleted IDs = %v", client.deleted)
	}
	next, _ = m.Update(noteDeleteResult{})
	m = next.(model)
	if m.screen != screenBusy || m.notification.text != "Note deleted" {
		t.Fatalf("delete did not refresh notes: screen=%d notification=%q", m.screen, m.notification.text)
	}
	next, _ = m.Update(notesResult{})
	m = next.(model)
	if m.screen != screenNotes || m.noteCount != 0 {
		t.Fatalf("refreshed notes = screen %d count %d", m.screen, m.noteCount)
	}
}
