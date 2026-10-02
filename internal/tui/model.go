package tui

import (
	"context"
	"fmt"

	"uuid"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app"
)

type screen uint8

const (
	rebuildOption askOptionID = "rebuild"
	cancelOption  askOptionID = "cancel"
	deleteOption  askOptionID = "delete"
)

const (
	screenMenu screen = iota
	screenForm
	screenAsk
	screenBusy
	screenNotes
)

type model struct {
	ctx          context.Context
	client       Client
	formInputs   []textinput.Model
	formLabels   []string
	formFocus    int
	notes        table.Model
	ask          ask
	notification notification
	noteEditor   noteEditorState
	listedNotes  []app.ListedNote
	deleteNoteID uuid.UUID

	screen    screen
	selected  int
	action    action
	problem   string
	width     int
	height    int
	noteCount int
}

func newModel(ctx context.Context, client Client) model {
	return model{ctx: ctx, client: client, notes: newNotesTable(80, 24), width: 80, height: 24}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		for index := range m.formInputs {
			m.formInputs[index].SetWidth(max(20, msg.Width-4))
		}
		m.resizeNotesTable()
		return m, nil
	case notesResult:
		if msg.err != nil {
			m.screen = screenMenu
			m.problem = msg.err.Error()
			return m, nil
		}
		m.setNotes(msg.notes)
		m.screen = screenNotes
		return m, nil
	case noteDeleteResult:
		m.deleteNoteID = uuid.Nil()
		if msg.err != nil {
			m.screen = screenNotes
			return m, m.notify(msg.err.Error(), notificationError)
		}
		next, cmd := m.loadNotes()
		m = next.(model)
		return m, tea.Batch(cmd, m.notify("Note deleted", notificationSuccess))
	case noteOpenResult:
		return m.openNoteEditor(msg)
	case noteEditorFinished:
		return m.finishNoteEdit(msg.err)
	case noteSaveResult:
		return m.finishNoteSave(msg)
	case operationResult:
		m.problem = ""
		if msg.err != nil {
			m.screen = msg.returnScreen
			return m, m.notify(msg.err.Error(), notificationError)
		}
		m.screen = screenMenu
		return m, m.notify(msg.message, notificationSuccess)
	case dismissNotification:
		if msg.id == m.notification.id {
			m.notification.text = ""
		}
		return m, nil
	case askAnswer:
		if m.deleteNoteID != uuid.Nil() {
			id := m.deleteNoteID
			m.deleteNoteID = uuid.Nil()
			if msg.option != deleteOption {
				m.screen = screenNotes
				return m, nil
			}
			m.screen = screenBusy
			return m, func() tea.Msg {
				return noteDeleteResult{err: m.client.DeleteNote(m.ctx, id)}
			}
		}
		switch m.action {
		case actionRebuild:
			if msg.option != rebuildOption {
				m.screen = screenMenu
				return m, nil
			}
			m.screen = screenBusy
			return m, func() tea.Msg {
				count, err := m.client.RebuildIndex(m.ctx)
				return operationResult{message: fmt.Sprintf("Indexed %d note(s)", count), err: err}
			}
		}
		return m, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.screen {
		case screenMenu:
			if msg.String() == "q" || msg.String() == "esc" {
				return m, tea.Quit
			}
			switch msg.String() {
			case "up", "k":
				m.selected = (m.selected - 1 + len(menuItems)) % len(menuItems)
				return m, nil
			case "down", "j":
				m.selected = (m.selected + 1) % len(menuItems)
				return m, nil
			case "enter":
				return m.begin(menuItems[m.selected].action)
			}
		case screenForm:
			if msg.String() == "esc" {
				m.screen = screenMenu
				m.problem = ""
				return m, nil
			}
			switch msg.String() {
			case "tab", "down":
				return m.focusFormField((m.formFocus + 1) % len(m.formInputs))
			case "shift+tab", "up":
				return m.focusFormField((m.formFocus - 1 + len(m.formInputs)) % len(m.formInputs))
			case "enter":
				return m.submitForm(false)
			case "ctrl+e":
				return m.submitForm(true)
			}
		case screenAsk:
			if msg.String() == "esc" || msg.String() == "q" {
				if m.deleteNoteID != uuid.Nil() {
					m.deleteNoteID = uuid.Nil()
					m.screen = screenNotes
				} else {
					m.screen = screenMenu
				}
				return m, nil
			}
		case screenBusy:
			return m, nil
		case screenNotes:
			if msg.String() == "esc" || msg.String() == "q" {
				m.screen = screenMenu
				return m, nil
			}
			switch msg.String() {
			case "e", "enter":
				return m.editSelectedNote()
			case "r":
				return m.loadNotes()
			case "d":
				return m.confirmDeleteSelectedNote()
			}
		}
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenForm:
		m.formInputs[m.formFocus], cmd = m.formInputs[m.formFocus].Update(message)
	case screenNotes:
		m.notes, cmd = m.notes.Update(message)
	case screenAsk:
		m.ask, cmd = m.ask.Update(message)
	}
	return m, cmd
}

func (m model) begin(selected action) (tea.Model, tea.Cmd) {
	m.action = selected
	m.problem = ""
	switch selected {
	case actionCapture:
		return m.startCapture()
	case actionPage:
		return m.startForm([]formField{{"Title", "Page title"}, {"Area (optional)", "project/subarea"}, {"Tags (optional, comma-separated)", "tag-one, tag-two"}})
	case actionPerson:
		return m.startForm([]formField{{"Name", "Full name"}, {"Groups (optional, comma-separated)", "work, university"}, {"Tags (optional, comma-separated)", "tag-one, tag-two"}})
	case actionBookmark:
		return m.startForm([]formField{{"URL", "https://example.com"}, {"Tags (optional, comma-separated)", "tag-one, tag-two"}})
	case actionJournal:
		return m.startJournal()
	case actionRebuild:
		m.ask = newAsk(
			"Rebuild the index from notes in the vault?",
			cancelOption,
			askOption{id: rebuildOption, label: "Yes", icon: "✓", shortcut: "y"},
			askOption{id: cancelOption, label: "No", icon: "✕", shortcut: "n"},
		)
		m.screen = screenAsk
		return m, nil
	case actionList:
		return m.loadNotes()
	}
	return m, nil
}

type operationResult struct {
	message      string
	returnScreen screen
	err          error
}
