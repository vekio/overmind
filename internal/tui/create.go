package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/domain/journals"
)

type noteCreated struct {
	action  action
	err     error
	journal *journals.Journal
	updated bool
}

func (m model) saveForm(values map[string]string) tea.Cmd {
	switch m.action {
	case actionInbox:
		return m.saveInbox(values)
	case actionPage:
		return m.savePage(values)
	case actionBookmark:
		return m.saveBookmark(values)
	case actionJournal:
		return m.saveJournal(values)
	case actionHabit:
		return m.saveHabit(values)
	case actionPerson:
		return m.savePerson(values)
	default:
		return func() tea.Msg {
			return noteCreated{action: m.action, err: fmt.Errorf("creation is not configured for %s", m.action)}
		}
	}
}
