package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m model) submitCapture() (tea.Model, tea.Cmd) {
	content := m.editor.Value()
	if strings.TrimSpace(content) == "" {
		m.problem = "Note text is required"
		return m, nil
	}
	m.screen = screenBusy
	return m, func() tea.Msg {
		path, err := m.client.Capture(m.ctx, content)
		return operationResult{message: "Captured " + path, err: err}
	}
}
