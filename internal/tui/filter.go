package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/vekio/overmind/internal/app"
)

// ListNotesFunc lets the TUI query notes through a local or remote client.
type ListNotesFunc func(context.Context, app.ListNotesQuery) (app.ListNotesResult, error)

const notesPageSize = 100

type loadNotes struct {
	revision uint64
	query    app.ListNotesQuery
}
type notesLoaded struct {
	revision uint64
	result   app.ListNotesResult
	err      error
}

func newFilterInputs() [2]textinput.Model {
	var inputs [2]textinput.Model
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].SetVirtualCursor(true)
		inputs[i].SetWidth(32)
		inputs[i].CharLimit = 80
	}
	inputs[0].Prompt = "Type: "
	inputs[0].Placeholder = "habit, person, bookmark…"
	inputs[1].Prompt = "Tag:  "
	inputs[1].Placeholder = "salud"
	return inputs
}

func (m *model) requestNotes(delay time.Duration) tea.Cmd {
	m.revision++
	m.loading = true
	m.problem = ""
	message := loadNotes{revision: m.revision, query: app.ListNotesQuery{
		Type: m.filters[0].Value(), Tag: m.filters[1].Value(), Limit: notesPageSize + 1, Offset: m.offset,
	}}
	if delay > 0 {
		return tea.Tick(delay, func(time.Time) tea.Msg { return message })
	}
	return func() tea.Msg { return message }
}

func (m model) fetchNotes(msg loadNotes) tea.Cmd {
	return func() tea.Msg {
		if m.listNotes == nil {
			return notesLoaded{revision: msg.revision, err: fmt.Errorf("note listing is not configured")}
		}
		result, err := m.listNotes(m.ctx, msg.query)
		return notesLoaded{revision: msg.revision, result: result, err: err}
	}
}

func (m model) updateFilters(message tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := message.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "enter", "esc":
			m.filterOpen = false
			for i := range m.filters {
				m.filters[i].Blur()
			}
			return m, nil
		case "tab", "shift+tab", "up", "down":
			m.filters[m.filterFocus].Blur()
			m.filterFocus = 1 - m.filterFocus
			return m, m.filters[m.filterFocus].Focus()
		}
	}
	before := m.filters[m.filterFocus].Value()
	var cmd tea.Cmd
	m.filters[m.filterFocus], cmd = m.filters[m.filterFocus].Update(message)
	if before != m.filters[m.filterFocus].Value() {
		m.offset = 0
		return m, tea.Batch(cmd, m.requestNotes(180*time.Millisecond))
	}
	return m, cmd
}

func (m model) filterView() string {
	width := max(12, min(54, m.width-6))
	inputs := m.filters
	for i := range inputs {
		inputs[i].SetWidth(max(1, width-8))
	}
	status := fmt.Sprintf("%d matching notes on this page", m.noteCount)
	if m.loading {
		status = "Searching…"
	}
	if m.problem != "" {
		status = m.problem
	}
	content := "Filter notes\n\n" + inputs[0].View() + "\n" + inputs[1].View() +
		"\n\nTypes: habit, person, bookmark,\ninbox, page, journal\n\n" + status +
		"\n\nTab: next field · Enter/Esc: close\nEmpty field: all · filters combine"
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).
		Width(width).Background(lipgloss.Color("235")).Foreground(lipgloss.Color("252")).Render(content)
}

func (m model) filterSummary() string {
	var filters []string
	if value := strings.TrimSpace(m.filters[0].Value()); value != "" {
		filters = append(filters, "type="+value)
	}
	if value := strings.TrimSpace(m.filters[1].Value()); value != "" {
		filters = append(filters, "tag="+value)
	}
	if len(filters) == 0 {
		return "all notes"
	}
	return strings.Join(filters, " · ")
}
