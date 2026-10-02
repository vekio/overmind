package tui

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app"
)

type editorClient struct {
	Client
	id            uuid.UUID
	updatedID     uuid.UUID
	updatedOld    []byte
	updatedSource []byte
	updateErr     error
}

func (client *editorClient) UpdateNote(_ context.Context, id uuid.UUID, original, source []byte) (app.UpdateNoteResult, error) {
	client.updatedID, client.updatedOld, client.updatedSource = id, bytes.Clone(original), bytes.Clone(source)
	return app.UpdateNoteResult{}, client.updateErr
}

func TestEditorKeepsFailedDraftAndRemovesSavedDraft(t *testing.T) {
	client := &editorClient{id: uuid.MustParse("11111111-1111-4111-8111-111111111111")}
	original := []byte("= Page\n\nOriginal\n")
	draft, err := createNoteDraft(original)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { removeDraft(draft) })
	if err := os.WriteFile(draft, []byte("= Page\n\nEdited\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m := newModel(context.Background(), client)
	m.noteEditor = noteEditorState{id: client.id, original: original, draftPath: draft, returnScreen: screenNotes}
	next, command := m.finishNoteEdit(nil)
	m = next.(model)
	if m.screen != screenBusy || command == nil {
		t.Fatal("edited draft did not start save")
	}
	result, ok := command().(noteSaveResult)
	if !ok || result.err != nil || client.updatedID != client.id || !bytes.Equal(client.updatedOld, original) || !strings.Contains(string(client.updatedSource), "Edited") {
		t.Fatalf("save command = %+v, source %q", result, client.updatedSource)
	}
	client.updateErr = errors.New("disk full")
	next, _ = m.finishNoteSave(noteSaveResult{err: client.updateErr})
	m = next.(model)
	if m.screen != screenNotes || !strings.Contains(m.problem, draft) {
		t.Fatalf("failed save state = %d, %q", m.screen, m.problem)
	}
	if _, err := os.Stat(draft); err != nil {
		t.Fatalf("failed draft was removed: %v", err)
	}
	next, _ = m.finishNoteSave(noteSaveResult{})
	m = next.(model)
	if _, err := os.Stat(draft); !os.IsNotExist(err) || m.noteEditor.id != uuid.Nil() {
		t.Fatalf("saved draft remains: %v, editor %+v", err, m.noteEditor)
	}
}

func TestUnchangedEditorDoesNotSaveNote(t *testing.T) {
	client := &editorClient{id: uuid.MustParse("11111111-1111-4111-8111-111111111111")}
	draft, err := createNoteDraft([]byte("unchanged"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { removeDraft(draft) })
	m := newModel(context.Background(), client)
	m.noteEditor = noteEditorState{id: client.id, original: []byte("unchanged"), draftPath: draft, returnScreen: screenNotes, unchangedMessage: "Note unchanged"}
	next, command := m.finishNoteEdit(nil)
	m = next.(model)
	if command == nil || m.screen != screenNotes || m.notification.text != "Note unchanged" || client.updatedID != uuid.Nil() {
		t.Fatal("unchanged note was saved or not reported")
	}
	if _, err := os.Stat(draft); !os.IsNotExist(err) {
		t.Fatalf("unchanged draft remains: %v", err)
	}
}
