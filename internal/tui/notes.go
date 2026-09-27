package tui

import (
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
)

type notesResult struct {
	notes []app.ListedNote
	err   error
}

func (m model) loadNotes() (tea.Model, tea.Cmd) {
	m.screen = screenBusy
	return m, func() tea.Msg {
		notes, err := m.client.ListNotes(m.ctx)
		return notesResult{notes: notes, err: err}
	}
}

func newNotesTable(width, height int) table.Model {
	return table.New(
		table.WithColumns(noteColumns(width)),
		table.WithWidth(max(20, width-2)),
		table.WithHeight(max(4, height-5)),
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
	m.notes.SetHeight(max(4, m.height-5))
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
	m.noteCount = len(notes)
}

func noteName(note app.ListedNote) string {
	switch note.Kind {
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
		return "Inbox " + note.ID.String()[:8]
	default:
		return note.ID.String()
	}
}
