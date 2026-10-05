package tui

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"uuid"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/rawedit"
	"github.com/vekio/overmind/internal/ports"
)

type rawTestClient struct {
	original  ports.Note
	saveErr   error
	changed   bool
	saves     int
	source    []byte
	result    *rawedit.UpdateResult
	indexOnly bool
}

func (client *rawTestClient) GetRawNote(context.Context, rawedit.GetQuery) (rawedit.GetResult, error) {
	return rawedit.GetResult{Note: client.original, Kind: ports.NoteKindPage}, nil
}

func (client *rawTestClient) UpdateRawNote(_ context.Context, command rawedit.UpdateCommand) (rawedit.UpdateResult, error) {
	client.saves++
	client.source = command.Source
	client.indexOnly = command.IndexOnly
	if client.result != nil {
		return *client.result, client.saveErr
	}
	return rawedit.UpdateResult{Changed: client.changed}, client.saveErr
}

func rawFixture(t *testing.T) (model, *rawTestClient) {
	t.Helper()
	m := newModel()
	client := &rawTestClient{original: ports.Note{ID: uuid.New(), Content: []byte("original source")}, changed: true}
	m.newRawEditClient = func(context.Context) (RawEditClient, error) { return client, nil }
	m.screen = screenNotes
	m.setNotes([]noteRow{{id: uuid.New(), kind: "page"}, {id: client.original.ID, kind: "page"}})
	m.notes.SetCursor(1)
	m.filters[0].SetValue("page")
	m.offset = notesPageSize
	t.Cleanup(func() {
		for _, draft := range m.rawDrafts {
			draft.Close()
		}
	})
	return m, client
}

func TestRawEditShortcut(t *testing.T) {
	m, _ := rawFixture(t)
	next, cmd := m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	m = next.(model)
	if m.screen != screenRawEdit || cmd == nil {
		t.Fatal("Ctrl+E did not open raw editor")
	}
}

func TestRawEditRetrySaveAndRestore(t *testing.T) {
	m, client := rawFixture(t)
	next, load := m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	m = next.(model)
	next, open := m.Update(load())
	m = next.(model)
	if open == nil || m.raw.draft == nil {
		t.Fatal("editor draft not prepared")
	}
	path := m.raw.draft.Path
	if err := os.WriteFile(path, []byte("edited source"), 0600); err != nil {
		t.Fatal(err)
	}
	client.saveErr = errors.New("invalid note")
	next, save := m.Update(rawEditorClosed{session: m.raw})
	m = next.(model)
	next, _ = m.Update(save())
	m = next.(model)
	if m.screen != screenRawEdit || !strings.Contains(m.rawStatus(), "invalid note") || m.raw.busy {
		t.Fatal("validation failure lost draft")
	}
	if source, err := os.ReadFile(path); err != nil || string(source) != "edited source" {
		t.Fatal("draft lost")
	}
	next, reopen := m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	m = next.(model)
	if reopen == nil || m.raw.draft.Path != path {
		t.Fatal("retry replaced draft")
	}
	client.saveErr = nil
	next, save = m.Update(rawEditorClosed{session: m.raw})
	m = next.(model)
	next, refresh := m.Update(save())
	m = next.(model)
	if m.screen != screenNotes || m.raw != nil || refresh == nil || m.restoreID != client.original.ID || m.restoreCursor != 1 {
		t.Fatal("save did not return to selected note")
	}
	if m.offset != notesPageSize || m.filters[0].Value() != "page" || string(client.source) != "edited source" {
		t.Fatal("state or content lost")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("successful draft not cleaned")
	}
}

func TestRawEditorFailureAndDiscard(t *testing.T) {
	m, client := rawFixture(t)
	next, load := m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	m = next.(model)
	next, _ = m.Update(load())
	m = next.(model)
	path := m.raw.draft.Path
	next, save := m.Update(rawEditorClosed{session: m.raw, err: errors.New("editor exited")})
	m = next.(model)
	if save != nil || client.saves != 0 || m.raw.busy {
		t.Fatal("failed editor saved note")
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	if m.screen != screenNotes || m.raw != nil {
		t.Fatal("discard did not leave editor")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("discarded draft not cleaned")
	}
}

func TestRawEditIndexRetryUsesSavedSnapshot(t *testing.T) {
	m, client := rawFixture(t)
	next, load := m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	m = next.(model)
	next, _ = m.Update(load())
	m = next.(model)
	if err := m.raw.draft.Write([]byte("edited source")); err != nil {
		t.Fatal(err)
	}
	saved := ports.Note{ID: client.original.ID, Content: []byte("saved source with managed timestamp")}
	client.result = &rawedit.UpdateResult{Changed: true, Note: saved, IndexPending: true}
	client.saveErr = rawedit.ErrIndexUpdate
	next, save := m.Update(rawEditorClosed{session: m.raw})
	m = next.(model)
	next, _ = m.Update(save())
	m = next.(model)
	if !m.raw.indexPending || !bytes.Equal(m.raw.original, saved.Content) || !strings.Contains(m.rawStatus(), "note is saved") {
		t.Fatal("saved document was treated as unsaved")
	}
	actual, err := m.raw.draft.Read()
	if err != nil || !bytes.Equal(actual, saved.Content) {
		t.Fatal("retry draft does not match saved document")
	}
	client.result = &rawedit.UpdateResult{Changed: true, Note: saved}
	client.saveErr = nil
	next, retry := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = next.(model)
	next, _ = m.Update(retry())
	m = next.(model)
	if !client.indexOnly || m.screen != screenNotes {
		t.Fatal("retry rewrote the note")
	}
}

func TestRawEditRetainsCopyAndReloadsLatestAfterConflict(t *testing.T) {
	m, client := rawFixture(t)
	next, load := m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	m = next.(model)
	next, _ = m.Update(load())
	m = next.(model)
	oldPath := m.raw.draft.Path
	if err := m.raw.draft.Write([]byte("my unsaved changes")); err != nil {
		t.Fatal(err)
	}
	client.saveErr = rawedit.ErrConflict
	next, save := m.Update(rawEditorClosed{session: m.raw})
	m = next.(model)
	next, _ = m.Update(save())
	m = next.(model)
	if !m.raw.conflict || !strings.Contains(m.rawStatus(), "not overwritten") {
		t.Fatal("conflict not explained")
	}
	client.original.Content = []byte("latest vault source")
	next, load = m.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	m = next.(model)
	loaded := rawLoaded{session: m.raw, note: client.original, kind: ports.NoteKindPage}
	next, _ = m.Update(loaded)
	m = next.(model)
	if m.raw.draft.Path == oldPath || string(m.raw.original) != "latest vault source" {
		t.Fatal("latest note was not loaded into a separate draft")
	}
	old, err := os.ReadFile(oldPath)
	if err != nil || string(old) != "my unsaved changes" {
		t.Fatal("conflict draft was lost")
	}
	retained, err := m.rawDrafts[oldPath].Finish()
	if err != nil || retained != oldPath {
		t.Fatal("conflict draft not retained on exit")
	}
}

func TestRawEditKeepDraftReturnsToNotes(t *testing.T) {
	m, _ := rawFixture(t)
	next, load := m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	m = next.(model)
	next, _ = m.Update(load())
	m = next.(model)
	// Simulate returning from the editor with an error and explicitly keeping it.
	next, _ = m.Update(rawEditorClosed{session: m.raw, err: errors.New("editor error")})
	m = next.(model)
	path := m.raw.draft.Path
	next, _ = m.Update(tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl})
	m = next.(model)
	if m.screen != screenNotes || m.raw != nil {
		t.Fatal("keep did not return to Notes")
	}
	retained, err := m.rawDrafts[path].Finish()
	if err != nil || retained != path {
		t.Fatal("kept draft removed at exit")
	}
}

func TestRawEditIgnoresRepliesFromReplacedSession(t *testing.T) {
	m, client := rawFixture(t)
	next, load := m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	m = next.(model)
	next, _ = m.Update(load())
	m = next.(model)
	previous := m.raw
	next, _ = m.Update(rawEditorClosed{session: previous, err: errors.New("editor stopped")})
	m = next.(model)
	next, _ = m.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	m = next.(model)
	current := m.raw
	next, _ = m.Update(rawLoaded{session: current, note: client.original, kind: ports.NoteKindPage})
	m = next.(model)
	path := current.draft.Path
	for _, reply := range []tea.Msg{
		rawLoaded{session: previous, err: errors.New("late load failure")},
		rawEditorClosed{session: previous, err: errors.New("late editor failure")},
		rawSaved{session: previous, changed: true},
		rawSaved{session: previous, err: rawedit.ErrConflict},
	} {
		next, cmd := m.Update(reply)
		m = next.(model)
		if cmd != nil || m.raw != current || m.screen != screenRawEdit || !m.raw.busy || m.raw.problem != "" {
			t.Fatalf("late reply changed current session: %T", reply)
		}
	}
	source, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(source, client.original.Content) || client.saves != 0 {
		t.Fatalf("late reply changed draft or triggered a save: %v", err)
	}
}

func TestRawEditBlocksDraftActionsWhileOperationIsPending(t *testing.T) {
	m, _ := rawFixture(t)
	next, load := m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	m = next.(model)
	next, _ = m.Update(load())
	m = next.(model)
	session, path := m.raw, m.raw.draft.Path
	for _, key := range []tea.KeyPressMsg{
		{Code: 'e', Mod: tea.ModCtrl}, {Code: 's', Mod: tea.ModCtrl},
		{Code: 'r', Mod: tea.ModCtrl}, {Code: 'k', Mod: tea.ModCtrl}, {Code: tea.KeyEscape},
	} {
		next, cmd := m.Update(key)
		m = next.(model)
		if cmd != nil || m.raw != session || m.screen != screenRawEdit || !m.raw.busy {
			t.Fatalf("pending operation accepted %s", key.String())
		}
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("pending draft removed: %v", err)
	}
}
