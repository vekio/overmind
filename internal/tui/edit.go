package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"uuid"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/bookmark"
	"github.com/vekio/overmind/internal/app/habit"
	"github.com/vekio/overmind/internal/app/inbox"
	"github.com/vekio/overmind/internal/app/journal"
	"github.com/vekio/overmind/internal/app/page"
	"github.com/vekio/overmind/internal/app/person"
	"github.com/vekio/overmind/internal/domain/shared"
)

// EditClient supplies the typed operations used by the shared edit forms.
type EditClient interface {
	// GetHabit loads and verifies the requested habit document.
	GetHabit(context.Context, habit.GetQuery) (habit.GetResult, error)
	// UpdateHabit replaces editable habit values while preserving identity and creation time.
	UpdateHabit(context.Context, habit.UpdateCommand) (habit.UpdateResult, error)
	// GetPerson loads and verifies the requested person document.
	GetPerson(context.Context, person.GetQuery) (person.GetResult, error)
	// UpdatePerson replaces editable person values while preserving identity and creation time.
	UpdatePerson(context.Context, person.UpdateCommand) (person.UpdateResult, error)
	// GetBookmark loads and verifies the requested bookmark document.
	GetBookmark(context.Context, bookmark.GetQuery) (bookmark.GetResult, error)
	// UpdateBookmark replaces editable bookmark values while preserving identity and creation time.
	UpdateBookmark(context.Context, bookmark.UpdateCommand) (bookmark.UpdateResult, error)
	// GetInbox loads and verifies the requested inbox document.
	GetInbox(context.Context, inbox.GetQuery) (inbox.GetResult, error)
	// UpdateInbox replaces editable inbox values while preserving identity and creation time.
	UpdateInbox(context.Context, inbox.UpdateCommand) (inbox.UpdateResult, error)
	// GetPage loads and verifies the requested page document.
	GetPage(context.Context, page.GetQuery) (page.GetResult, error)
	// UpdatePage replaces editable page values while preserving identity and creation time.
	UpdatePage(context.Context, page.UpdateCommand) (page.UpdateResult, error)
	// GetJournal loads and verifies the requested journal document.
	GetJournal(context.Context, journal.GetQuery) (journal.GetResult, error)
	// UpdateJournal replaces editable journal values while preserving identity and creation time.
	UpdateJournal(context.Context, journal.UpdateCommand) (journal.UpdateResult, error)
}

// EditClientFactory resolves the typed client for loading and saving note forms.
type EditClientFactory func(context.Context) (EditClient, error)

type editLoaded struct {
	revision uint64
	id       uuid.UUID
	form     form
	err      error
}

type editFinished struct {
	revision uint64
	id       uuid.UUID
	err      error
}

type editableNote interface {
	ID() uuid.UUID
	Metadata() shared.EntityMetadata
	Tags() shared.Tags
}

func setFormValue(f *form, id, value string) {
	for i := range f.fields {
		field := &f.fields[i]
		if field.spec.id != id {
			continue
		}
		switch field.spec.kind {
		case fieldStatic:
			field.spec.value = value
		case fieldMultiline:
			field.textarea.SetValue(value)
		case fieldGroups:
			field.groups.chosen = formListValues(value)
		case fieldSelect:
			for j, option := range field.spec.options {
				if option == value {
					field.selected = j
				}
			}
		default:
			field.input.SetValue(value)
		}
		return
	}
}

func (m model) beginEdit() (tea.Model, tea.Cmd) {
	if m.loading || m.deleting || m.reindexing {
		return m, m.notify("Wait for the current operation to finish", notificationInfo)
	}
	cursor := m.notes.Cursor()
	if cursor < 0 || cursor >= len(m.rows) || m.rows[cursor].id == uuid.Nil() {
		return m, m.notify("Select a note to edit", notificationInfo)
	}
	row := m.rows[cursor]
	switch row.kind {
	case "habit":
		m.action = actionHabit
	case "person":
		m.action = actionPerson
	case "bookmark":
		m.action = actionBookmark
	case "inbox":
		m.action = actionInbox
	case "page":
		m.action = actionPage
	case "journal":
		m.action = actionJournal
	default:
		return m, m.notify("Unsupported note type", notificationError)
	}
	m.editID = row.id
	m.editCursor = cursor
	m.editRevision++
	m.screen = screenEdit
	m.problem = ""
	cmd := m.loadEdit(row.kind)
	return m, cmd
}

func (m model) loadEdit(kind string) tea.Cmd {
	return func() tea.Msg {
		msg := editLoaded{revision: m.editRevision, id: m.editID}
		if m.newEditClient == nil {
			msg.err = fmt.Errorf("note editing is not configured")
			return msg
		}
		client, err := m.newEditClient(m.ctx)
		if err != nil {
			msg.err = err
			return msg
		}
		if client == nil {
			msg.err = fmt.Errorf("note editing is not configured")
			return msg
		}
		var entity editableNote
		switch kind {
		case "habit":
			result, err := client.GetHabit(m.ctx, habit.GetQuery{ID: m.editID.String()})
			if err != nil {
				msg.err = err
				return msg
			}
			if result.Habit == nil {
				msg.err = fmt.Errorf("loaded habit is missing")
				return msg
			}
			entity = result.Habit
			msg.form = newHabitForm()
			setFormValue(&msg.form, "content", result.Habit.Content())
			setFormValue(&msg.form, "title", result.Habit.Title().String())
			setFormValue(&msg.form, "amount", strconv.FormatFloat(result.Habit.Goal().Amount(), 'f', -1, 64))
			setFormValue(&msg.form, "unit", result.Habit.Goal().Unit().String())
			setFormValue(&msg.form, "period", result.Habit.Goal().Period().String())
		case "person":
			result, err := client.GetPerson(m.ctx, person.GetQuery{ID: m.editID.String()})
			if err != nil {
				msg.err = err
				return msg
			}
			if result.Person == nil {
				msg.err = fmt.Errorf("loaded person is missing")
				return msg
			}
			entity = result.Person
			msg.form = newPersonForm()
			setFormValue(&msg.form, "content", result.Person.Content())
			setFormValue(&msg.form, "name", result.Person.Name().String())
			setFormValue(&msg.form, "groups", strings.Join(result.Person.Groups().Strings(), ", "))
		case "bookmark":
			result, err := client.GetBookmark(m.ctx, bookmark.GetQuery{ID: m.editID.String()})
			if err != nil {
				msg.err = err
				return msg
			}
			if result.Bookmark == nil {
				msg.err = fmt.Errorf("loaded bookmark is missing")
				return msg
			}
			entity = result.Bookmark
			msg.form = newBookmarkForm()
			setFormValue(&msg.form, "content", result.Bookmark.Content())
			setFormValue(&msg.form, "url", result.Bookmark.URL().String())
		case "inbox":
			result, err := client.GetInbox(m.ctx, inbox.GetQuery{ID: m.editID.String()})
			if err != nil {
				msg.err = err
				return msg
			}
			if result.Inbox == nil {
				msg.err = fmt.Errorf("loaded inbox is missing")
				return msg
			}
			entity = result.Inbox
			msg.form = newInboxForm()
			setFormValue(&msg.form, "content", result.Inbox.Content())
		case "page":
			result, err := client.GetPage(m.ctx, page.GetQuery{ID: m.editID.String()})
			if err != nil {
				msg.err = err
				return msg
			}
			if result.Page == nil {
				msg.err = fmt.Errorf("loaded page is missing")
				return msg
			}
			entity = result.Page
			msg.form = newPageForm()
			setFormValue(&msg.form, "title", result.Page.Title().String())
			setFormValue(&msg.form, "area", result.Page.Area().String())
			setFormValue(&msg.form, "content", result.Page.Content())
		case "journal":
			result, err := client.GetJournal(m.ctx, journal.GetQuery{ID: m.editID.String()})
			if err != nil {
				msg.err = err
				return msg
			}
			if result.Journal == nil {
				msg.err = fmt.Errorf("loaded journal is missing")
				return msg
			}
			entity = result.Journal
			msg.form = newJournalForm(result.Journal.Date().String())
			setFormValue(&msg.form, "date", result.Journal.Date().String())
			setFormValue(&msg.form, "content", result.Journal.Content())
		default:
			msg.err = fmt.Errorf("unsupported note type")
			return msg
		}
		if entity.ID() != m.editID {
			msg.err = fmt.Errorf("loaded note identity does not match selected note")
			return msg
		}
		setFormValue(&msg.form, "tags", strings.Join(entity.Tags().Strings(), ", "))
		metadata := entity.Metadata()
		msg.form.detail = fmt.Sprintf("Created %s · Updated %s", metadata.CreatedAt().Local().Format("2006-01-02 15:04"), metadata.UpdatedAt().Local().Format("2006-01-02 15:04"))
		return msg
	}
}

func (m model) updateNote(values map[string]string) tea.Cmd {
	return func() tea.Msg {
		msg := editFinished{revision: m.editRevision, id: m.editID}
		if m.newEditClient == nil {
			msg.err = fmt.Errorf("note editing is not configured")
			return msg
		}
		client, err := m.newEditClient(m.ctx)
		if err != nil {
			msg.err = err
			return msg
		}
		if client == nil {
			msg.err = fmt.Errorf("note editing is not configured")
			return msg
		}
		tags := formListValues(values["tags"])
		switch m.action {
		case actionHabit:
			amount, err := habitAmount(values["amount"])
			if err != nil {
				msg.err = err
				return msg
			}
			_, msg.err = client.UpdateHabit(m.ctx, habit.UpdateCommand{ID: m.editID.String(), Content: values["content"], Title: values["title"], Amount: amount, Unit: values["unit"], Period: values["period"], Tags: tags})
		case actionPerson:
			_, msg.err = client.UpdatePerson(m.ctx, person.UpdateCommand{ID: m.editID.String(), Content: values["content"], Name: values["name"], Groups: formListValues(values["groups"]), Tags: tags})
		case actionBookmark:
			_, msg.err = client.UpdateBookmark(m.ctx, bookmark.UpdateCommand{ID: m.editID.String(), Content: values["content"], URL: values["url"], Tags: tags})
		case actionInbox:
			_, msg.err = client.UpdateInbox(m.ctx, inbox.UpdateCommand{ID: m.editID.String(), Content: values["content"], Tags: tags})
		case actionPage:
			_, msg.err = client.UpdatePage(m.ctx, page.UpdateCommand{ID: m.editID.String(), Title: values["title"], Area: strings.TrimSpace(values["area"]), Content: values["content"], Tags: tags})
		case actionJournal:
			_, msg.err = client.UpdateJournal(m.ctx, journal.UpdateCommand{ID: m.editID.String(), Content: values["content"], Tags: tags})
		default:
			msg.err = fmt.Errorf("unsupported note type")
		}
		return msg
	}
}

// leaveEdit returns to Notes without resetting its filters or current page.
func (m *model) leaveEdit() {
	m.screen = screenNotes
	m.form = form{}
	m.editID = uuid.Nil()
	m.problem = ""
}
