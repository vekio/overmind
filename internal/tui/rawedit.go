package tui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"uuid"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/rawedit"
	"github.com/vekio/overmind/internal/editor"
	"github.com/vekio/overmind/internal/ports"
)

// RawEditClient supplies source reads and validated writes; the TUI owns only
// the external-editor process and its private draft files.
type RawEditClient interface {
	// GetRawNote loads managed source without requiring a valid body.
	GetRawNote(context.Context, rawedit.GetQuery) (rawedit.GetResult, error)
	// UpdateRawNote validates and saves source, reporting partial index failures.
	UpdateRawNote(context.Context, rawedit.UpdateCommand) (rawedit.UpdateResult, error)
}

// RawEditClientFactory resolves the client when a raw-edit operation starts.
type RawEditClientFactory func(context.Context) (RawEditClient, error)

// rawEditSession is also the identity token for asynchronous replies. Replies
// belonging to a discarded or replaced session must not change the active draft.
type rawEditSession struct {
	id     uuid.UUID
	kind   ports.NoteKind
	cursor int
	// original is the last known vault source, not the current editor buffer.
	original []byte
	draft    *editor.Draft
	busy     bool
	problem  string
	// indexPending means the source was saved but its projection still needs repair.
	indexPending bool
	conflict     bool
}

type rawLoaded struct {
	session *rawEditSession
	note    ports.Note
	kind    ports.NoteKind
	err     error
}

type rawEditorClosed struct {
	session *rawEditSession
	err     error
}

type rawSaved struct {
	session      *rawEditSession
	changed      bool
	note         ports.Note
	indexPending bool
	err          error
}

func (m model) beginRawEdit() (tea.Model, tea.Cmd) {
	if m.loading || m.deleting || m.reindexing {
		return m, m.notify("Wait for the current operation to finish", notificationInfo)
	}
	cursor := m.notes.Cursor()
	if cursor < 0 || cursor >= len(m.rows) || m.rows[cursor].id == uuid.Nil() {
		return m, m.notify("Select a note to edit", notificationInfo)
	}
	if m.newRawEditClient == nil {
		return m, m.notify("Raw editing is not configured", notificationError)
	}
	row := m.rows[cursor]
	session := &rawEditSession{id: row.id, kind: ports.NoteKind(row.kind), cursor: cursor, busy: true}
	m.raw = session
	m.screen = screenRawEdit
	m.problem = ""
	return m, func() tea.Msg {
		msg := rawLoaded{session: session}
		client, err := m.newRawEditClient(m.ctx)
		if err != nil {
			msg.err = err
			return msg
		}
		if client == nil {
			msg.err = fmt.Errorf("raw editing is not configured")
			return msg
		}
		result, err := client.GetRawNote(m.ctx, rawedit.GetQuery{ID: session.id.String()})
		msg.note, msg.kind, msg.err = result.Note, result.Kind, err
		return msg
	}
}

func (m model) rawLoaded(msg rawLoaded) (tea.Model, tea.Cmd) {
	if m.raw != msg.session || m.screen != screenRawEdit {
		return m, nil
	}
	if msg.err != nil {
		m.closeRawEdit()
		return m, m.notify(msg.err.Error(), notificationError)
	}
	draft, err := editor.New(msg.note.Content)
	if err != nil {
		m.closeRawEdit()
		return m, m.notify(err.Error(), notificationError)
	}
	m.rawDrafts[draft.Path] = draft
	m.raw.kind = msg.kind
	// Keep the loaded source separate from the file the editor will modify.
	m.raw.original = append([]byte(nil), msg.note.Content...)
	m.raw.draft = draft
	return m, m.openRawEditor()
}

// openRawEditor suspends Bubble Tea so the editor receives the real terminal.
func (m model) openRawEditor() tea.Cmd {
	session := m.raw
	session.busy = true
	session.problem = ""
	return tea.ExecProcess(session.draft.Command(m.ctx), func(err error) tea.Msg {
		return rawEditorClosed{session: session, err: err}
	})
}

func (m model) rawEditorClosed(msg rawEditorClosed) (tea.Model, tea.Cmd) {
	if m.raw != msg.session || m.screen != screenRawEdit {
		return m, nil
	}
	if msg.err != nil {
		m.raw.busy = false
		m.raw.draft.Keep()
		m.raw.problem = fmt.Sprintf("Editor failed: %v. Draft retained; reopen or discard.", msg.err)
		return m, nil
	}
	return m, m.saveRawEdit()
}

// saveRawEdit retries only indexing when the saved source is unchanged. Any
// additional editor changes follow the normal validation and persistence path.
func (m model) saveRawEdit() tea.Cmd {
	session := m.raw
	session.busy = true
	session.problem = ""
	indexPending := session.indexPending
	return func() tea.Msg {
		msg := rawSaved{session: session}
		source, err := session.draft.Read()
		if err != nil {
			msg.err = err
			return msg
		}
		client, err := m.newRawEditClient(m.ctx)
		if err != nil {
			msg.err = err
			return msg
		}
		if client == nil {
			msg.err = fmt.Errorf("raw editing is not configured")
			return msg
		}
		result, err := client.UpdateRawNote(m.ctx, rawedit.UpdateCommand{
			ID: session.id.String(), Kind: session.kind, Original: session.original, Source: source,
			IndexOnly: indexPending && bytes.Equal(source, session.original),
		})
		msg.changed, msg.note, msg.indexPending, msg.err = result.Changed, result.Note, result.IndexPending, err
		return msg
	}
}

func (m model) rawSaved(msg rawSaved) (tea.Model, tea.Cmd) {
	if m.raw != msg.session || m.screen != screenRawEdit {
		return m, nil
	}
	m.raw.busy = false
	if msg.err != nil {
		m.raw.draft.Keep()
		m.raw.conflict = errors.Is(msg.err, rawedit.ErrConflict)
		if msg.indexPending {
			// Adopt the persisted bytes so an index retry cannot rewrite the document
			// or advance its managed timestamp a second time.
			m.raw.original = append([]byte(nil), msg.note.Content...)
			m.raw.indexPending = true
			if err := m.raw.draft.Write(msg.note.Content); err != nil {
				m.raw.problem = fmt.Sprintf("%v; could not refresh draft: %v", msg.err, err)
				return m, nil
			}
		}
		m.raw.problem = msg.err.Error()
		return m, nil
	}
	m.restoreID, m.restoreCursor = m.raw.id, m.raw.cursor
	m.closeRawEdit()
	if !msg.changed {
		return m, m.notify("No changes", notificationInfo)
	}
	return m, tea.Batch(m.requestNotes(0), m.notify("Note updated", notificationSuccess))
}

func (m model) updateRawEdit(key tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.raw == nil || m.raw.busy {
		return m, nil
	}
	switch key.String() {
	case "ctrl+e":
		return m, m.openRawEditor()
	case "ctrl+s":
		return m, m.saveRawEdit()
	case "ctrl+k":
		path := m.raw.draft.Path
		m.raw.draft.Keep()
		m.raw = nil
		m.screen = screenNotes
		return m, m.notify("Draft retained at "+path, notificationInfo)
	case "ctrl+r":
		path := m.raw.draft.Path
		m.raw.draft.Keep()
		m.raw = nil
		m.screen = screenNotes
		next, load := m.beginRawEdit()
		fresh := next.(model)
		return fresh, tea.Batch(load, fresh.notify("Previous draft retained at "+path, notificationInfo))
	case "esc":
		m.closeRawEdit()
	}
	return m, nil
}

func (m *model) closeRawEdit() {
	if m.raw != nil && m.raw.draft != nil {
		m.raw.draft.Close()
		delete(m.rawDrafts, m.raw.draft.Path)
	}
	m.raw = nil
	m.screen = screenNotes
}

func (m model) rawStatus() string {
	if m.raw == nil {
		return "Loading…"
	}
	if m.raw.busy {
		return "Editing / validating…"
	}
	detail := "Draft retained at " + m.raw.draft.Path
	if m.raw.indexPending {
		detail += "\nThe note is saved. Ctrl+S retries indexing without saving again."
	} else if m.raw.conflict {
		detail += "\nThe vault note was not overwritten. Ctrl+R opens the latest note and keeps this draft."
	} else {
		detail += "\nReopen the editor to correct it, keep the draft, or discard."
	}
	return "Error: " + m.raw.problem + "\n\n" + detail
}

func (m model) rawControls() []controlHint {
	if m.raw == nil || m.raw.busy {
		return []controlHint{{"ctrl+c", "quit"}}
	}
	save := "retry save"
	if m.raw.indexPending {
		save = "retry index"
	}
	return []controlHint{{"ctrl+e", "editor"}, {"ctrl+s", save}, {"ctrl+r", "latest note"}, {"ctrl+k", "keep draft"}, {"esc", "discard"}}
}
