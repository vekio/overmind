package tui

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	appjournal "github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/domain/journals"
	"github.com/vekio/overmind/internal/editor"
)

type screen uint8

const (
	screenMenu screen = iota
	screenAsk
	screenNotes
	screenForm
	screenJournal
	screenEdit
	screenRawEdit
)
const (
	reindexOption askOptionID = "reindex"
	cancelOption  askOptionID = "cancel"
	deleteOption  askOptionID = "delete"
)

// The model runs application operations asynchronously.
type model struct {
	newRawEditClient         RawEditClientFactory
	raw                      *rawEditSession
	rawDrafts                map[string]*editor.Draft
	ctx                      context.Context
	newEditClient            EditClientFactory
	editID                   uuid.UUID
	editRevision             uint64
	editCursor               int
	restoreID                uuid.UUID
	restoreCursor            int
	listNotes                ListNotesFunc
	reindex                  ReindexFunc
	reindexing               bool
	deleteNote               DeleteNoteFunc
	createInbox              CreateInboxFunc
	createPage               CreatePageFunc
	createBookmark           CreateBookmarkFunc
	createHabit              CreateHabitFunc
	createPerson             CreatePersonFunc
	listGroups               ListGroupsFunc
	groupsRevision           uint64
	createJournal            CreateJournalFunc
	getJournal               GetJournalFunc
	updateJournal            UpdateJournalFunc
	journalDate              string
	journal                  *journals.Journal
	journalLoading           bool
	journalRevision          uint64
	form                     form
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
	return model{rawDrafts: make(map[string]*editor.Draft), ctx: context.Background(), filters: newFilterInputs(), notes: newNotesTable(80, 24), width: 80, height: 24}
}
func newApplicationModel(ctx context.Context, listNotes ListNotesFunc) model {
	m := newModel()
	m.ctx, m.listNotes = ctx, listNotes
	return m
}

// Init starts with no commands; queries are scheduled when their screen is opened.
func (m model) Init() tea.Cmd { return nil }

// Update routes terminal events and asynchronous results to the active screen.
// Revision and session checks prevent stale replies from replacing current state.
func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case rawLoaded:
		return m.rawLoaded(msg)
	case rawEditorClosed:
		return m.rawEditorClosed(msg)
	case rawSaved:
		return m.rawSaved(msg)
	case editLoaded:
		if m.screen != screenEdit || msg.revision != m.editRevision || msg.id != m.editID {
			return m, nil
		}
		if msg.err != nil {
			m.leaveEdit()
			return m, m.notify(msg.err.Error(), notificationError)
		}
		m.form = msg.form
		m.screen = screenForm
		m.form.Resize(m.width, m.height)
		focus := m.form.Focus()
		if m.action == actionPerson {
			load := m.requestGroups()
			return m, tea.Batch(focus, load)
		}
		return m, focus
	case editFinished:
		if m.screen != screenForm || !m.form.saving || msg.revision != m.editRevision || msg.id != m.editID {
			return m, nil
		}
		m.form.saving = false
		if msg.err != nil {
			m.form.problem = msg.err.Error()
			return m, m.form.Focus()
		}
		m.restoreID, m.restoreCursor = m.editID, m.editCursor
		m.leaveEdit()
		refresh := m.requestNotes(0)
		return m, tea.Batch(refresh, m.notify("Note updated", notificationSuccess))
	case groupsLoaded:
		if m.screen != screenForm || m.action != actionPerson || msg.revision != m.groupsRevision {
			return m, nil
		}
		picker := &m.form.fields[1].groups
		picker.loading = false
		if msg.err != nil {
			picker.problem = msg.err.Error()
		} else {
			picker.options = msg.groups
			picker.problem = ""
			picker.cursor = 0
		}
		m.form.Resize(m.width, m.height)
		return m, nil
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
		if m.restoreID != uuid.Nil() {
			cursor := min(m.restoreCursor, max(0, len(m.rows)-1))
			for i, row := range m.rows {
				if row.id == m.restoreID {
					cursor = i
					break
				}
			}
			m.notes.SetCursor(cursor)
			m.restoreID = uuid.Nil()
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resizeNotesTable()
		m.form.Resize(m.width, m.height)
		return m, nil
	case formSubmitted:
		if m.screen != screenForm || m.form.saving {
			return m, nil
		}
		m.form.saving = true
		m.form.problem = ""
		if m.editID != uuid.Nil() {
			return m, m.updateNote(msg.values)
		}
		return m, m.saveForm(msg.values)
	case journalLoaded:
		if m.screen != screenJournal || msg.revision != m.journalRevision {
			return m, nil
		}
		m.journalLoading = false
		if msg.err != nil {
			m.problem = msg.err.Error()
			return m, nil
		}
		m.journal = msg.journal
		if msg.journal != nil {
			return m, m.showJournalForm(msg.expected)
		}
		if msg.expected {
			m.problem = "Today's journal exists but could not be found in the index. Reindex notes and try again."
			return m, nil
		}
		return m, m.showJournalForm(false)
	case formCancelled:
		if m.screen == screenForm && !m.form.saving {
			if m.editID != uuid.Nil() {
				m.leaveEdit()
			} else {
				m.screen = screenMenu
				m.form = form{}
			}
		}
		return m, nil
	case noteCreated:
		if m.screen != screenForm || !m.form.saving || msg.action != m.action {
			return m, nil
		}
		m.form.saving = false
		if msg.action == actionJournal && errors.Is(msg.err, appjournal.ErrAlreadyExists) {
			return m, m.requestJournal(true)
		}
		if msg.err != nil {
			m.form.problem = msg.err.Error()
			return m, m.form.Focus()
		}
		if msg.action == actionJournal {
			m.journal = msg.journal
			focus := m.showJournalForm(false)
			verb := "created"
			if msg.updated {
				verb = "updated"
			}
			return m, tea.Batch(focus, m.notify("Journal "+verb, notificationSuccess))
		}
		m.screen = screenMenu
		m.form = form{}
		return m, m.notify(fmt.Sprintf("%s created", msg.action), notificationSuccess)
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
		case screenRawEdit:
			return m.updateRawEdit(msg)
		case screenEdit:
			if msg.String() == "esc" || msg.String() == "q" {
				m.leaveEdit()
			}
			return m, nil
		case screenForm:
			if msg.String() == "ctrl+r" && m.action == actionPerson && !m.form.saving && m.form.focus == 1 && !m.form.fields[1].groups.loading {
				cmd := m.requestGroups()
				return m, cmd
			}
		case screenJournal:
			switch msg.String() {
			case "q", "esc":
				m.screen = screenMenu
				return m, nil
			case "r":
				if !m.journalLoading {
					return m, m.requestJournal(false)
				}
				return m, nil
			}
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
			case "ctrl+e":
				return m.beginRawEdit()
			case "e", "enter":
				return m.beginEdit()
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
	case screenForm:
		m.form, cmd = m.form.Update(message)
	}
	return m, cmd
}
func (m model) begin(selected action) (tea.Model, tea.Cmd) {
	m.action = selected
	m.editID = uuid.Nil()
	m.problem = ""
	switch selected {
	case actionJournal:
		m.journalDate = time.Now().Format(time.DateOnly)
		return m, m.requestJournal(false)
	case actionInbox, actionPage, actionBookmark, actionHabit, actionPerson:
		m.screen = screenForm
		switch selected {
		case actionInbox:
			m.form = newInboxForm()
		case actionPage:
			m.form = newPageForm()
		case actionBookmark:
			m.form = newBookmarkForm()
		case actionHabit:
			m.form = newHabitForm()
		case actionPerson:
			m.form = newPersonForm()
		}
		m.form.Resize(m.width, m.height)
		focus := m.form.Focus()
		if selected == actionPerson {
			load := m.requestGroups()
			return m, tea.Batch(focus, load)
		}
		return m, focus
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
		return m, m.notify("Unknown action", notificationError)
	}
}
