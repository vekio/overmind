package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/page"
	"github.com/vekio/overmind/internal/domain/pages"
	"github.com/vekio/overmind/internal/domain/shared"
)

// CreatePageFunc connects form submissions to the page creation use case.
type CreatePageFunc func(context.Context, page.CreateCommand) (page.CreateResult, error)

func newPageForm() form {
	return newForm("Page",
		fieldSpec{id: "title", label: "Title", kind: fieldText, placeholder: "Required, page title", validate: func(raw string) error {
			_, err := shared.NewTitle(raw)
			return err
		}},
		fieldSpec{id: "area", label: "Area", kind: fieldText, placeholder: "Optional, example: work/projects", validate: func(raw string) error {
			if strings.TrimSpace(raw) == "" {
				return nil
			}
			_, err := pages.NewArea(raw)
			return err
		}},
		largeContentField("Write your page…"),
		tagsField(),
	)
}

func (m model) savePage(values map[string]string) tea.Cmd {
	command := page.CreateCommand{Title: values["title"], Content: values["content"], Area: strings.TrimSpace(values["area"]), Tags: formListValues(values["tags"])}
	return func() tea.Msg {
		if m.createPage == nil {
			return noteCreated{action: actionPage, err: fmt.Errorf("page creation is not configured")}
		}
		_, err := m.createPage(m.ctx, command)
		return noteCreated{action: actionPage, err: err}
	}
}
