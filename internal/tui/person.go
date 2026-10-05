package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/person"
	"github.com/vekio/overmind/internal/domain/persons"
	"github.com/vekio/overmind/internal/domain/shared"
)

// CreatePersonFunc connects form submissions to the person creation use case.
type CreatePersonFunc func(context.Context, person.CreateCommand) (person.CreateResult, error)

func newPersonForm() form {
	return newForm("Person",
		fieldSpec{
			id: "name", label: "Name", kind: fieldText, placeholder: "Required, person's name",
			validate: func(raw string) error {
				_, err := shared.NewTitle(raw)
				return err
			},
		},
		fieldSpec{
			id: "groups", label: "Groups", kind: fieldGroups,
			placeholder: "Search existing groups or type a new name",
		},
		largeContentField("Write your person…"),
		tagsField(),
	)
}

func validateFormGroups(raw string) error {
	var groups []persons.Group
	for _, value := range formListValues(raw) {
		group, err := persons.NewGroup(value)
		if err != nil {
			return err
		}
		groups = append(groups, group)
	}
	_, err := persons.NewGroups(groups...)
	return err
}

func (m model) savePerson(values map[string]string) tea.Cmd {
	command := person.CreateCommand{
		Name:    values["name"],
		Content: values["content"],
		Groups:  formListValues(values["groups"]),
		Tags:    formListValues(values["tags"]),
	}
	return func() tea.Msg {
		if m.createPerson == nil {
			return noteCreated{action: actionPerson, err: fmt.Errorf("person creation is not configured")}
		}
		_, err := m.createPerson(m.ctx, command)
		return noteCreated{action: actionPerson, err: err}
	}
}
