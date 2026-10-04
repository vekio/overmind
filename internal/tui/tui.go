// Package tui provides the default Overmind terminal screen.
package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

// Run opens the interactive menu and loads notes through the supplied query.
func Run(ctx context.Context, listNotes ListNotesFunc, reindex ReindexFunc, deleteNote DeleteNoteFunc) error {
	model := newApplicationModel(ctx, listNotes)
	model.reindex = reindex
	model.deleteNote = deleteNote
	_, err := tea.NewProgram(model, tea.WithContext(ctx)).Run()
	return err
}
