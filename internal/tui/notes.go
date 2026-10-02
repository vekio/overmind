package tui

import (
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/domain"
)

type notesResult struct {
	notes []app.ListedNote
	err   error
}

type noteDeleteResult struct {
	err error
}

func (m model) loadNotes() (tea.Model, tea.Cmd) {
	m.screen = screenBusy
	return m, func() tea.Msg {
		notes, err := m.client.ListNotes(m.ctx)
		return notesResult{notes: notes, err: err}
	}
}

func (m model) editSelectedNote() (tea.Model, tea.Cmd) {
	if len(m.listedNotes) == 0 {
		return m, nil
	}
	index := m.notes.Cursor()
	if index < 0 || index >= len(m.listedNotes) {
		m.problem = "selected note is no longer available"
		return m, nil
	}
	id := m.listedNotes[index].ID
	m.screen = screenBusy
	return m, func() tea.Msg {
		source, err := m.client.OpenNote(m.ctx, id)
		return noteOpenResult{
			id: id, source: source, returnScreen: screenNotes,
			unchangedMessage: "Note unchanged", err: err,
		}
	}
}

func (m model) confirmDeleteSelectedNote() (tea.Model, tea.Cmd) {
	if len(m.listedNotes) == 0 {
		return m, nil
	}
	index := m.notes.Cursor()
	if index < 0 || index >= len(m.listedNotes) {
		m.problem = "selected note is no longer available"
		return m, nil
	}
	note := m.listedNotes[index]
	m.deleteNoteID = note.ID
	m.ask = newAsk(
		"Delete "+noteName(note)+" from the vault?",
		cancelOption,
		askOption{id: deleteOption, label: "Yes", icon: "✓", shortcut: "y"},
		askOption{id: cancelOption, label: "No", icon: "✕", shortcut: "n"},
	)
	m.screen = screenAsk
	return m, nil
}

func newNotesTable(width, height int) table.Model {
	return table.New(
		table.WithColumns(noteColumns(width)),
		table.WithWidth(max(20, width-2)),
		table.WithHeight(notesTableHeight(height, 0)),
		table.WithFocused(true),
	)
}

func noteColumns(width int) []table.Column {
	return []table.Column{
		{Title: "Kind", Width: 9},
		{Title: "Note", Width: max(20, width-59)},
		{Title: "Tags", Width: 16},
		{Title: "Updated", Width: 16},
	}
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

func (m *model) setNotes(notes []app.ListedNote) {
	rows := make([]table.Row, 0, len(notes))
	for _, note := range notes {
		rows = append(rows, table.Row{
			note.Kind.String(), noteName(note), strings.Join(note.Tags, ", "),
			note.UpdatedAt.Local().Format("2006-01-02 15:04"),
		})
	}
	m.notes.SetRows(rows)
	m.notes.Focus()
	m.listedNotes = notes
	m.noteCount = len(notes)
	m.resizeNotesTable()
}

func noteName(note app.ListedNote) string {
	switch note.Kind {
	case domain.NoteKindPerson:
		if groups := note.Attributes["groups"]; groups != "" {
			return note.Attributes["name"] + " [" + groups + "]"
		}
		return note.Attributes["name"]
	case domain.NoteKindPage:
		if area := note.Attributes["area"]; area != "" {
			return note.Attributes["title"] + " [" + area + "]"
		}
		return note.Attributes["title"]
	case domain.NoteKindBookmark:
		return note.Attributes["url"]
	case domain.NoteKindJournal:
		return note.Attributes["date"]
	case domain.NoteKindInbox:
		return "Capture " + note.CreatedAt.Local().Format("2006-01-02 15:04")
	default:
		return note.ID.String()
	}
}
