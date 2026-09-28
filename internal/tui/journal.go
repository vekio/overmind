package tui

import (
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
)

func (m model) startJournal() (tea.Model, tea.Cmd) {
	m.screen = screenBusy
	return m, func() tea.Msg {
		date := domain.DateFromTime(time.Now())
		id, exists, err := m.client.FindJournal(m.ctx, date)
		if err != nil {
			return noteOpenResult{returnScreen: screenMenu, err: err}
		}
		created := false
		createdPath := ""
		if !exists {
			result, createErr := m.client.CreateJournal(m.ctx, domain.Tags{})
			if errors.Is(createErr, app.ErrJournalAlreadyExists) {
				id, exists, err = m.client.FindJournal(m.ctx, date)
				if err != nil {
					return noteOpenResult{returnScreen: screenMenu, err: err}
				}
				if !exists {
					return noteOpenResult{returnScreen: screenMenu, err: createErr}
				}
			} else if createErr != nil {
				return noteOpenResult{returnScreen: screenMenu, err: createErr}
			} else {
				id = result.Journal.Metadata().ID()
				created = true
				createdPath = result.Path
			}
		}
		source, err := m.client.OpenNote(m.ctx, id)
		unchangedMessage := "Journal unchanged"
		if created {
			unchangedMessage = "Today's journal created"
		}
		return noteOpenResult{
			id: id, source: source, createdPath: createdPath, returnScreen: screenMenu,
			unchangedMessage: unchangedMessage, err: err,
		}
	}
}
