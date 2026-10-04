package tui

import (
	"strings"
	"time"
	"uuid"

	"charm.land/bubbles/v2/table"
)

// noteRow contains display data only, independent of the application's entities.
type noteRow struct {
	id         uuid.UUID
	kind, name string
	tags       []string
	updatedAt  time.Time
}

func newNotesTable(width, height int) table.Model {
	return table.New(table.WithColumns(noteColumns(width)), table.WithWidth(max(20, width-2)), table.WithHeight(notesTableHeight(height, 0)), table.WithFocused(true))
}
func noteColumns(width int) []table.Column {
	return []table.Column{{Title: "Kind", Width: 9}, {Title: "Note", Width: max(20, width-59)}, {Title: "Tags", Width: 16}, {Title: "Updated", Width: 16}}
}
func (m *model) resizeNotesTable() {
	m.notes.SetColumns(noteColumns(m.width))
	m.notes.SetWidth(max(20, m.width-2))
	m.notes.SetHeight(notesTableHeight(m.height, m.noteCount))
}
func notesTableHeight(height, count int) int {
	reserved := 5
	if count == 0 {
		reserved = 7
	}
	return max(4, height-reserved)
}
func (m *model) setNotes(notes []noteRow) {
	m.rows = append([]noteRow(nil), notes...)
	rows := make([]table.Row, 0, len(notes))
	for _, note := range notes {
		rows = append(rows, table.Row{note.kind, note.name, strings.Join(note.tags, ", "), note.updatedAt.Local().Format("2006-01-02 15:04")})
	}
	m.notes.SetRows(rows)
	m.notes.Focus()
	m.noteCount = len(notes)
	m.resizeNotesTable()
}
