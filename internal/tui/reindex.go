package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app"
)

type ReindexFunc func(context.Context, app.ReindexCommand) (app.ReindexResult, error)
type reindexFinished struct {
	result app.ReindexResult
	err    error
}

func (m model) runReindex() tea.Cmd {
	return func() tea.Msg {
		if m.reindex == nil {
			return reindexFinished{err: fmt.Errorf("reindex is not configured")}
		}
		result, err := m.reindex(m.ctx, app.ReindexCommand{})
		return reindexFinished{result: result, err: err}
	}
}
