// Package tui provides the default Overmind terminal screen.
package tui

import (
	"context"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

// Run opens the TUI with application callbacks and restores the terminal on exit.
// Unfinished raw-edit drafts are retained and their paths are reported afterward.
func Run(
	ctx context.Context,
	listNotes ListNotesFunc,
	reindex ReindexFunc,
	deleteNote DeleteNoteFunc,
	createInbox CreateInboxFunc,
	createPage CreatePageFunc,
	createBookmark CreateBookmarkFunc,
	createJournal CreateJournalFunc,
	getJournal GetJournalFunc,
	updateJournal UpdateJournalFunc,
	createHabit CreateHabitFunc,
	createPerson CreatePersonFunc,
	listGroups ListGroupsFunc,
	newEditClient EditClientFactory,
	newRawEditClient RawEditClientFactory,
) error {
	model := newApplicationModel(ctx, listNotes)
	model.reindex = reindex
	model.deleteNote = deleteNote
	model.createInbox = createInbox
	model.createPage = createPage
	model.createBookmark = createBookmark
	model.createJournal = createJournal
	model.getJournal = getJournal
	model.updateJournal = updateJournal
	model.createHabit = createHabit
	model.createPerson = createPerson
	model.listGroups = listGroups
	model.newEditClient = newEditClient
	model.newRawEditClient = newRawEditClient
	defer func() {
		for _, draft := range model.rawDrafts {
			path, err := draft.Finish()
			if path != "" {
				fmt.Fprintf(os.Stderr, "Raw edit draft retained at %s\n", path)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}()
	_, err := tea.NewProgram(model, tea.WithContext(ctx)).Run()
	return err
}
