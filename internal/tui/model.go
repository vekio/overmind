package tui

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type screen uint8

const (
	screenMenu screen = iota
	screenForm
	screenCapture
	screenConfirm
	screenBusy
	screenNotes
)

type model struct {
	ctx    context.Context
	client Client
	input  textinput.Model
	editor textarea.Model
	notes  table.Model

	screen     screen
	selected   int
	action     action
	fields     []field
	fieldIndex int
	status     string
	problem    string
	width      int
	height     int
	noteCount  int
}

func newModel(ctx context.Context, client Client) model {
	input := textinput.New()
	input.Prompt = "> "
	input.SetWidth(72)
	editor := textarea.New()
	editor.Placeholder = "Write your note here..."
	editor.SetWidth(72)
	editor.SetHeight(12)

	return model{ctx: ctx, client: client, input: input, editor: editor, notes: newNotesTable(80, 24), width: 80, height: 24}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.SetWidth(max(20, msg.Width-4))
		m.editor.SetWidth(max(20, msg.Width-4))
		m.editor.SetHeight(max(4, msg.Height-7))
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
	case operationResult:
		m.screen = screenMenu
		m.problem = ""
		if msg.err != nil {
			m.problem = msg.err.Error()
		} else {
			m.status = msg.message
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
			if msg.String() == "enter" {
				return m.advance()
			}
		case screenCapture:
			if msg.String() == "esc" {
				m.screen = screenMenu
				m.problem = ""
				return m, nil
			}
			if msg.String() == "ctrl+s" {
				return m.submitCapture()
			}
		case screenConfirm:
			switch msg.String() {
			case "y", "Y":
				m.screen = screenBusy
				return m, func() tea.Msg {
					count, err := m.client.RebuildIndex(m.ctx)
					return operationResult{message: fmt.Sprintf("Indexed %d note(s)", count), err: err}
				}
			case "n", "N", "esc":
				m.screen = screenMenu
				return m, nil
			}
		case screenBusy:
			return m, nil
		case screenNotes:
			if msg.String() == "esc" || msg.String() == "q" {
				m.screen = screenMenu
				return m, nil
			}
			if msg.String() == "r" {
				return m.loadNotes()
			}
		}
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenForm:
		m.input, cmd = m.input.Update(message)
	case screenCapture:
		m.editor, cmd = m.editor.Update(message)
	case screenNotes:
		m.notes, cmd = m.notes.Update(message)
	}
	return m, cmd
}

func (m model) begin(selected action) (tea.Model, tea.Cmd) {
	m.action = selected
	m.problem = ""
	m.status = ""
	m.fieldIndex = 0
	m.input.Reset()
	switch selected {
	case actionCapture:
		m.screen = screenCapture
		m.editor.Reset()
		return m, m.editor.Focus()
	case actionPage:
		m.fields = []field{{label: "Title", placeholder: "Page title"}, {label: "Area (optional)", placeholder: "project/subarea"}, {label: "Tags (optional, comma-separated)", placeholder: "tag-one, tag-two"}}
	case actionBookmark:
		m.fields = []field{{label: "URL", placeholder: "https://example.com"}, {label: "Tags (optional, comma-separated)", placeholder: "tag-one, tag-two"}}
	case actionJournal:
		m.fields = []field{{label: "Tags (optional, comma-separated)", placeholder: "tag-one, tag-two"}}
	case actionRebuild:
		m.screen = screenConfirm
		return m, nil
	case actionList:
		return m.loadNotes()
	}
	m.screen = screenForm
	m.input.Placeholder = m.fields[0].placeholder
	return m, m.input.Focus()
}

type operationResult struct {
	message string
	err     error
}
