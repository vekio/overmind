package tui

import (
	"fmt"
	"strings"

	"uuid"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/domain"
)

type formField struct {
	label       string
	placeholder string
}

func (m model) startForm(fields []formField) (tea.Model, tea.Cmd) {
	m.formInputs = make([]textinput.Model, len(fields))
	m.formLabels = make([]string, len(fields))
	m.formFocus = 0
	for index, field := range fields {
		input := textinput.New()
		input.Prompt = "> "
		input.Placeholder = field.placeholder
		input.SetWidth(max(20, m.width-4))
		m.formInputs[index] = input
		m.formLabels[index] = field.label
	}
	m.screen = screenForm
	return m, m.formInputs[0].Focus()
}

func (m model) focusFormField(index int) (tea.Model, tea.Cmd) {
	m.formInputs[m.formFocus].Blur()
	m.formFocus = index
	return m, m.formInputs[index].Focus()
}

func (m model) formView() string {
	var view strings.Builder
	fmt.Fprintf(&view, "Overmind / %s\n", m.action)
	for index, input := range m.formInputs {
		marker := "  "
		if index == m.formFocus {
			marker = "› "
		}
		fmt.Fprintf(&view, "\n%s%s\n%s", marker, m.formLabels[index], input.View())
	}
	return view.String()
}

func (m model) formError(index int, err error) (tea.Model, tea.Cmd) {
	m.problem = err.Error()
	return m.focusFormField(index)
}

func (m model) submitForm(edit bool) (tea.Model, tea.Cmd) {
	var create func() (uuid.UUID, string, error)
	var createdMessage string
	switch m.action {
	case actionPerson:
		name, err := domain.NewTitle(m.formInputs[0].Value())
		if err != nil {
			return m.formError(0, err)
		}
		var values []domain.Group
		if value := m.formInputs[1].Value(); strings.TrimSpace(value) != "" {
			for _, part := range strings.Split(value, ",") {
				group, err := domain.NewGroup(part)
				if err != nil {
					return m.formError(1, err)
				}
				values = append(values, group)
			}
		}
		groups, err := domain.NewGroups(values...)
		if err != nil {
			return m.formError(1, err)
		}
		tags, err := parseTags(m.formInputs[2].Value())
		if err != nil {
			return m.formError(2, err)
		}
		create = func() (uuid.UUID, string, error) {
			result, err := m.client.CreatePerson(m.ctx, name, groups, tags)
			return result.Person.Metadata().ID(), result.Path, err
		}
		createdMessage = "Person created"
	case actionPage:
		title, err := domain.NewTitle(strings.TrimSpace(m.formInputs[0].Value()))
		if err != nil {
			return m.formError(0, err)
		}
		var area domain.Area
		if value := strings.TrimSpace(m.formInputs[1].Value()); value != "" {
			area, err = domain.NewArea(value)
			if err != nil {
				return m.formError(1, err)
			}
		}
		tags, err := parseTags(m.formInputs[2].Value())
		if err != nil {
			return m.formError(2, err)
		}
		create = func() (uuid.UUID, string, error) {
			result, err := m.client.CreatePage(m.ctx, title, area, tags)
			return result.Page.Metadata().ID(), result.Path, err
		}
		createdMessage = "Page created"
	case actionBookmark:
		url, err := domain.NewURL(strings.TrimSpace(m.formInputs[0].Value()))
		if err != nil {
			return m.formError(0, err)
		}
		tags, err := parseTags(m.formInputs[1].Value())
		if err != nil {
			return m.formError(1, err)
		}
		create = func() (uuid.UUID, string, error) {
			result, err := m.client.CreateBookmark(m.ctx, url, tags)
			return result.Bookmark.Metadata().ID(), result.Path, err
		}
		createdMessage = "Bookmark created"
	default:
		return m, nil
	}

	m.problem = ""
	m.screen = screenBusy
	return m, func() tea.Msg {
		id, path, err := create()
		if err != nil {
			return operationResult{returnScreen: screenForm, err: err}
		}
		if !edit {
			return operationResult{message: createdMessage}
		}
		source, err := m.client.OpenNote(m.ctx, id)
		if err != nil {
			return operationResult{err: fmt.Errorf("created %s, but cannot open it for editing: %w", path, err)}
		}
		return noteOpenResult{
			id: id, source: source, createdPath: path, returnScreen: screenMenu,
			unchangedMessage: createdMessage,
		}
	}
}

func parseTags(value string) (domain.Tags, error) {
	if strings.TrimSpace(value) == "" {
		return domain.Tags{}, nil
	}
	parts := strings.Split(value, ",")
	tags := make([]domain.Tag, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return domain.Tags{}, fmt.Errorf("tags must be comma-separated without empty entries")
		}
		tag, err := domain.NewTag(part)
		if err != nil {
			return domain.Tags{}, err
		}
		tags = append(tags, tag)
	}
	return domain.NewTags(tags...)
}
