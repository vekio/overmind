package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"git.casta.me/alberto/overmind/internal/domain"
)

type field struct {
	label       string
	placeholder string
	value       string
}

func (m model) advance() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.input.Value())
	if err := validateField(m.action, m.fieldIndex, value); err != nil {
		m.problem = err.Error()
		return m, nil
	}
	m.problem = ""
	m.fields[m.fieldIndex].value = value
	if m.fieldIndex+1 < len(m.fields) {
		m.fieldIndex++
		m.input.SetValue(m.fields[m.fieldIndex].value)
		m.input.Placeholder = m.fields[m.fieldIndex].placeholder
		return m, nil
	}
	return m.submitForm()
}

func validateField(selected action, index int, value string) error {
	switch {
	case selected == actionPage && index == 0:
		_, err := domain.NewTitle(value)
		return err
	case selected == actionPage && index == 1 && value != "":
		_, err := domain.NewArea(value)
		return err
	case selected == actionBookmark && index == 0:
		_, err := domain.NewURL(value)
		return err
	default:
		_, err := parseTags(value)
		return err
	}
}

func (m model) submitForm() (tea.Model, tea.Cmd) {
	var operation func() (string, error)
	switch m.action {
	case actionPage:
		title, _ := domain.NewTitle(m.fields[0].value)
		var area domain.Area
		if m.fields[1].value != "" {
			area, _ = domain.NewArea(m.fields[1].value)
		}
		tags, _ := parseTags(m.fields[2].value)
		operation = func() (string, error) { return m.client.CreatePage(m.ctx, title, area, tags) }
	case actionBookmark:
		url, _ := domain.NewURL(m.fields[0].value)
		tags, _ := parseTags(m.fields[1].value)
		operation = func() (string, error) { return m.client.CreateBookmark(m.ctx, url, tags) }
	case actionJournal:
		tags, _ := parseTags(m.fields[0].value)
		operation = func() (string, error) { return m.client.CreateJournal(m.ctx, tags) }
	}
	m.screen = screenBusy
	return m, func() tea.Msg {
		path, err := operation()
		return operationResult{message: "Created " + path, err: err}
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
