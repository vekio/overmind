package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/inbox"
)

// CreateInboxFunc connects form submissions to the inbox creation use case.
type CreateInboxFunc func(context.Context, inbox.CreateCommand) (inbox.CreateResult, error)

func newInboxForm() form {
	return newForm("Inbox",
		contentField("Write your note…"),
		tagsField(),
	)
}

func (m model) saveInbox(values map[string]string) tea.Cmd {
	command := inbox.CreateCommand{Content: values["content"], Tags: formListValues(values["tags"])}
	return func() tea.Msg {
		if m.createInbox == nil {
			return noteCreated{action: actionInbox, err: fmt.Errorf("inbox creation is not configured")}
		}
		_, err := m.createInbox(m.ctx, command)
		return noteCreated{action: actionInbox, err: err}
	}
}
