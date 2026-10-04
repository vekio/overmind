package tui

import (
	"context"
	"fmt"
	"uuid"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type screen uint8

const (
	screenMenu screen = iota
	screenAsk
	screenNotes
)
const (
	reindexOption askOptionID = "reindex"
	cancelOption  askOptionID = "cancel"
	deleteOption  askOptionID = "delete"
)

// The model queries notes asynchronously; creation actions remain placeholders.
type model struct {
	ctx                      context.Context
	listNotes                ListNotesFunc
	reindex                  ReindexFunc
	reindexing               bool
	deleteNote               DeleteNoteFunc
	deleteID                 uuid.UUID
	deleting                 bool
	rows                     []noteRow
	askReturn                screen
	filters                  [2]textinput.Model
	filterOpen               bool
	filterFocus              int
	revision                 uint64
	loading, hasMore         bool
	offset                   int
	notes                    table.Model
	ask                      ask
	notification             notification
	screen                   screen
	selected                 int
	action                   action
	problem                  string
	width, height, noteCount int
}

func newModel() model {
	return model{ctx: context.Background(), filters: newFilterInputs(), notes: newNotesTable(80, 24), width: 80, height: 24}
}
func newApplicationModel(ctx context.Context, listNotes ListNotesFunc) model {
	m := newModel()
	m.ctx, m.listNotes = ctx, listNotes
	return m
}
func (m model) Init() tea.Cmd { return nil }
func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case loadNotes:
		if msg.revision != m.revision {
			return m, nil
		}
		return m, m.fetchNotes(msg)
	case notesLoaded:
		if msg.revision != m.revision {
			return m, nil
		}
		m.loading = false
		if msg.err != nil {
			m.problem = msg.err.Error()
			m.setNotes(nil)
			m.hasMore = false
			if m.filterOpen {
				return m, nil
			}
			return m, m.notify(m.problem, notificationError)
		}
		m.problem = ""
		notes := msg.result.Notes
		m.hasMore = len(notes) > notesPageSize
		if m.hasMore {
			notes = notes[:notesPageSize]
		}
		rows := make([]noteRow, 0, len(notes))
		for _, note := range notes {
			rows = append(rows, noteRow{id: note.ID, kind: note.Type, name: note.Label, tags: note.Tags, updatedAt: note.UpdatedAt})
		}
		m.setNotes(rows)
		m.notes.SetCursor(0)
		return m, nil
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resizeNotesTable()
		return m, nil
	case dismissNotification:
		if msg.id == m.notification.id {
			m.notification.text = ""
		}
		return m, nil
	case deleteFinished:
		if !m.deleting || msg.id != m.deleteID {
			return m, nil
		}
		m.deleting = false
		m.deleteID = uuid.Nil()
		if msg.err != nil {
			return m, m.notify(msg.err.Error(), notificationError)
		}
		if m.noteCount <= 1 && m.offset > 0 {
			m.offset = max(0, m.offset-notesPageSize)
		}
		notification := m.notify("Note deleted", notificationSuccess)
		return m, tea.Batch(notification, m.requestNotes(0))
	case reindexFinished:
		m.reindexing = false
		if msg.err != nil {
			return m, m.notify(msg.err.Error(), notificationError)
		}
		notification := m.notify(fmt.Sprintf("Indexed %d notes", msg.result.Indexed), notificationSuccess)
		if m.screen == screenNotes {
			return m, tea.Batch(notification, m.requestNotes(0))
		}
		return m, notification
	case askAnswer:
		m.screen = m.askReturn
		if msg.option == deleteOption && !m.deleting && m.deleteID != uuid.Nil() {
			m.deleting = true
			return m, m.runDelete()
		}
		if msg.option == reindexOption && !m.reindexing {
			m.reindexing = true
			return m, m.runReindex()
		}
		return m, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.filterOpen {
			return m.updateFilters(message)
		}
		switch m.screen {
		case screenMenu:
			switch msg.String() {
			case "q", "esc":
				return m, tea.Quit
			case "up", "k":
				m.selected = (m.selected - 1 + len(menuItems)) % len(menuItems)
				return m, nil
			case "down", "j":
				m.selected = (m.selected + 1) % len(menuItems)
				return m, nil
			case "enter":
				return m.begin(menuItems[m.selected].action)
			}
		case screenAsk:
			if msg.String() == "q" || msg.String() == "esc" {
				m.screen = m.askReturn
				return m, nil
			}
		case screenNotes:
			switch msg.String() {
			case "q", "esc":
				m.screen = screenMenu
				return m, nil
			case "e", "enter":
				return m, m.notify("Demo: edit selected", notificationInfo)
			case "d":
				return m.confirmDelete()
			case "r":
				return m, m.requestNotes(0)
			case "f", "/":
				m.filterOpen = true
				return m, m.filters[m.filterFocus].Focus()
			case "n":
				if m.hasMore && !m.loading {
					m.offset += notesPageSize
					return m, m.requestNotes(0)
				}
				return m, nil
			case "p":
				if m.offset > 0 && !m.loading {
					m.offset = max(0, m.offset-notesPageSize)
					return m, m.requestNotes(0)
				}
				return m, nil
			}
		}
	}
	if m.filterOpen {
		return m.updateFilters(message)
	}
	var cmd tea.Cmd
	switch m.screen {
	case screenAsk:
		m.ask, cmd = m.ask.Update(message)
	case screenNotes:
		m.notes, cmd = m.notes.Update(message)
	}
	return m, cmd
}
func (m model) begin(selected action) (tea.Model, tea.Cmd) {
	m.action = selected
	m.problem = ""
	switch selected {
	case actionList:
		m.screen = screenNotes
		return m, m.requestNotes(0)
	case actionReindex:
		m.askReturn = screenMenu
		if m.reindexing {
			return m, m.notify("Reindexing is already running", notificationInfo)
		}
		m.ask = newAsk("Rebuild the index from note files?", cancelOption,
			askOption{id: reindexOption, label: "Yes", icon: "✓", shortcut: "y"},
			askOption{id: cancelOption, label: "No", icon: "✕", shortcut: "n"})
		m.screen = screenAsk
		return m, nil
	default:
		return m, m.notify(fmt.Sprintf("Demo: %s selected", selected), notificationInfo)
	}
}
