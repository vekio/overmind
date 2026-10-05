package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/notes"
)

// ReindexFunc rebuilds searchable projections from the vault documents.
type ReindexFunc func(context.Context, notes.ReindexCommand) (notes.ReindexResult, error)
type reindexFinished struct {
	result notes.ReindexResult
	err    error
}

func (m model) runReindex() tea.Cmd {
	return func() tea.Msg {
		if m.reindex == nil {
			return reindexFinished{err: fmt.Errorf("reindex is not configured")}
		}
		result, err := m.reindex(m.ctx, notes.ReindexCommand{})
		return reindexFinished{result: result, err: err}
	}
}
