package tui

import tea "charm.land/bubbletea/v2"

func (m model) startCapture() (tea.Model, tea.Cmd) {
	m.screen = screenBusy
	return m, func() tea.Msg {
		result, err := m.client.Capture(m.ctx, "")
		if err != nil {
			return noteOpenResult{returnScreen: screenMenu, err: err}
		}
		id := result.Inbox.Metadata().ID()
		source, err := m.client.OpenNote(m.ctx, id)
		return noteOpenResult{
			id: id, source: source, createdPath: result.Path, returnScreen: screenMenu,
			unchangedMessage: "Note captured", err: err,
		}
	}
}
