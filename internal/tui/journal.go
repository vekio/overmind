package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	appjournal "github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/app/notes"
	"github.com/vekio/overmind/internal/domain/journals"
)

// CreateJournalFunc connects form submissions to the journal creation use case.
type CreateJournalFunc func(context.Context, appjournal.CreateCommand) (appjournal.CreateResult, error)

// GetJournalFunc loads the document selected for today's journal.
type GetJournalFunc func(context.Context, appjournal.GetQuery) (appjournal.GetResult, error)

// UpdateJournalFunc persists journal content and tags without changing its date.
type UpdateJournalFunc func(context.Context, appjournal.UpdateCommand) (appjournal.UpdateResult, error)

type journalLoaded struct {
	revision uint64
	journal  *journals.Journal
	expected bool
	err      error
}

func newJournalForm(date string) form {
	f := newForm("Journal",
		fieldSpec{id: "date", label: "Date", kind: fieldStatic, value: date},
		largeContentField("Write your day…"),
		tagsField(),
	)
	f.focus = 1
	return f
}

// showJournalForm preserves typed content when another writer creates today's
// journal first; the next save updates that existing entry.
func (m *model) showJournalForm(preserveDraft bool) tea.Cmd {
	m.screen = screenForm
	if !preserveDraft {
		m.form = newJournalForm(m.journalDate)
		if m.journal != nil {
			m.form.fields[2].input.SetValue(strings.Join(m.journal.Tags().Strings(), ", "))
			m.form.fields[1].textarea.SetValue(m.journal.Content())
		}
	} else {
		m.form.problem = "Journal already exists. Your draft is preserved; Ctrl+S updates today's entry."
	}
	if m.journal != nil {
		metadata := m.journal.Metadata()
		m.form.detail = fmt.Sprintf(
			"Created %s · Updated %s",
			metadata.CreatedAt().Local().Format("2006-01-02 15:04"),
			metadata.UpdatedAt().Local().Format("2006-01-02 15:04"),
		)
	}
	m.form.Resize(m.width, m.height)
	return m.form.Focus()
}

// requestJournal locates today's entry across index pages and verifies the
// loaded source identity and date before offering it for editing.
func (m *model) requestJournal(expected bool) tea.Cmd {
	m.screen = screenJournal
	m.journalLoading = true
	m.journal = nil
	m.problem = ""
	m.journalRevision++
	revision, date := m.journalRevision, m.journalDate
	ctx, listNotes, getJournal := m.ctx, m.listNotes, m.getJournal
	return func() tea.Msg {
		result := journalLoaded{revision: revision, expected: expected}
		if listNotes == nil {
			result.err = fmt.Errorf("journal lookup is not configured")
			return result
		}
		for offset := 0; ; offset += notesPageSize {
			if result.err = ctx.Err(); result.err != nil {
				return result
			}
			page, err := listNotes(ctx, notes.ListQuery{Type: "journal", Limit: notesPageSize, Offset: offset})
			if err != nil {
				result.err = err
				return result
			}
			for _, note := range page.Notes {
				if note.Type == "journal" && note.Label == date {
					if getJournal == nil {
						result.err = fmt.Errorf("journal reader is not configured")
						return result
					}
					loaded, err := getJournal(ctx, appjournal.GetQuery{ID: note.ID.String()})
					result.err = err
					if err == nil {
						if loaded.Journal == nil || loaded.Journal.ID() != note.ID || loaded.Journal.Date().String() != date {
							result.err = fmt.Errorf("loaded journal does not match today's entry")
						} else {
							result.journal = loaded.Journal
						}
					}
					return result
				}
			}
			if len(page.Notes) < notesPageSize {
				return result
			}
		}
	}
}

func (m model) saveJournal(values map[string]string) tea.Cmd {
	content, tags := values["content"], formListValues(values["tags"])
	return func() tea.Msg {
		message := noteCreated{action: actionJournal, updated: m.journal != nil}
		if m.journal == nil {
			if m.createJournal == nil {
				message.err = fmt.Errorf("journal creation is not configured")
				return message
			}
			result, err := m.createJournal(m.ctx, appjournal.CreateCommand{Date: m.journalDate, Content: content, Tags: tags})
			message.journal, message.err = result.Journal, err
		} else {
			if m.updateJournal == nil {
				message.err = fmt.Errorf("journal editing is not configured")
				return message
			}
			result, err := m.updateJournal(m.ctx, appjournal.UpdateCommand{ID: m.journal.ID().String(), Content: content, Tags: tags})
			message.journal, message.err = result.Journal, err
		}
		if message.err == nil && (message.journal == nil || message.journal.Date().String() != m.journalDate || m.journal != nil && message.journal.ID() != m.journal.ID()) {
			message.err = fmt.Errorf("saved journal does not match today's entry")
		}
		return message
	}
}

func (m model) journalView() string {
	content := "Overmind / Journal\n\nDate > " + m.journalDate
	if m.journalLoading {
		return content + "\n\nLoading…"
	}
	return content
}
