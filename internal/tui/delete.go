package tui

import (
	"context"
	"fmt"
	"strings"
	"uuid"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app"
)

type DeleteNoteFunc func(context.Context, app.DeleteNoteCommand) (app.DeleteNoteResult, error)
type deleteFinished struct {
	id  uuid.UUID
	err error
}

func (m model) confirmDelete() (tea.Model, tea.Cmd) {
	if m.loading || m.deleting || m.reindexing {
		return m, m.notify("Wait for the current operation to finish", notificationInfo)
	}
	cursor := m.notes.Cursor()
	if cursor < 0 || cursor >= len(m.rows) || m.rows[cursor].id == uuid.Nil() {
		return m, m.notify("Select a note to delete", notificationInfo)
	}
	note := m.rows[cursor]
	m.deleteID = note.id
	m.askReturn = screenNotes
	label := strings.Join(strings.Fields(note.name), " ")
	m.ask = newAsk(fmt.Sprintf("Delete %s: %s?\nID: %s", note.kind, label, note.id), cancelOption,
		askOption{id: deleteOption, label: "Delete", icon: "✕", shortcut: "y"},
		askOption{id: cancelOption, label: "Cancel", shortcut: "n"})
	m.screen = screenAsk
	return m, nil
}

func (m model) runDelete() tea.Cmd {
	id := m.deleteID
	return func() tea.Msg {
		if m.deleteNote == nil {
			return deleteFinished{id: id, err: fmt.Errorf("note deletion is not configured")}
		}
		_, err := m.deleteNote(m.ctx, app.DeleteNoteCommand{ID: id.String()})
		return deleteFinished{id: id, err: err}
	}
}
