package tui

import (
	"context"
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/person"
	"github.com/vekio/overmind/internal/domain/persons"
)

// ListGroupsFunc provides existing group names for person membership selection.
type ListGroupsFunc func(context.Context, person.ListGroupsQuery) (person.ListGroupsResult, error)

type groupsLoaded struct {
	revision uint64
	groups   []string
	err      error
}

type groupPicker struct {
	options []string
	chosen  []string
	cursor  int
	loading bool
	problem string
}

type groupChoice struct {
	name   string
	create bool
}

func (m *model) requestGroups() tea.Cmd {
	m.groupsRevision++
	revision, ctx, list := m.groupsRevision, m.ctx, m.listGroups
	m.form.fields[1].groups.loading = true
	m.form.fields[1].groups.problem = ""
	return func() tea.Msg {
		if list == nil {
			return groupsLoaded{revision: revision, err: fmt.Errorf("group lookup is not configured")}
		}
		result, err := list(ctx, person.ListGroupsQuery{})
		return groupsLoaded{revision: revision, groups: result.Groups, err: err}
	}
}

func (field formField) groupChoices() []groupChoice {
	query := strings.ToLower(strings.TrimSpace(field.input.Value()))
	normalized, err := persons.NewGroup(query)
	var choices []groupChoice
	exact := false
	options := slices.Clone(field.groups.options)
	for _, name := range field.groups.chosen {
		if !slices.Contains(options, name) {
			options = append(options, name)
		}
	}
	slices.Sort(options)
	for _, name := range options {
		if err == nil && name == normalized.String() {
			exact = true
		}
		if strings.Contains(name, query) || err == nil && strings.Contains(name, normalized.String()) {
			choices = append(choices, groupChoice{name: name})
		}
	}
	if query != "" && !exact && err == nil && !strings.Contains(query, ",") {
		choices = append(choices, groupChoice{name: normalized.String(), create: true})
	}
	return choices
}

func (field *formField) updateGroups(message tea.Msg) tea.Cmd {
	choices := field.groupChoices()
	if key, ok := message.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "up", "down":
			if len(choices) > 0 {
				step := 1
				if key.String() == "up" {
					step = -1
				}
				field.groups.cursor = (field.groups.cursor + step + len(choices)) % len(choices)
			}
			return nil
		case "enter":
			if len(choices) == 0 {
				field.problem = "Enter a valid group name"
				return nil
			}
			choice := choices[min(field.groups.cursor, len(choices)-1)]
			index := slices.Index(field.groups.chosen, choice.name)
			if index >= 0 {
				field.groups.chosen = slices.Delete(slices.Clone(field.groups.chosen), index, index+1)
			} else {
				field.groups.chosen = append(slices.Clone(field.groups.chosen), choice.name)
			}
			field.input.SetValue("")
			field.groups.cursor = 0
			field.problem = ""
			return nil
		}
	}
	before := field.input.Value()
	var cmd tea.Cmd
	field.input, cmd = field.input.Update(message)
	if before != field.input.Value() {
		field.groups.cursor = 0
		field.problem = ""
	}
	return cmd
}

func (field formField) groupsView(label string, focused bool) string {
	summary := strings.Join(field.groups.chosen, ", ")
	if summary == "" {
		summary = controlDescriptionStyle.Render("Optional, select or add groups")
	}
	view := label + " > " + summary
	if !focused {
		return view
	}
	field.input.Prompt = "Search > "
	view += "\n" + field.input.View()
	if field.groups.loading {
		view += "\n  Loading groups…"
	}
	if field.groups.problem != "" {
		view += "\n  " + field.groups.problem + " · ctrl+r retry"
	}
	choices := field.groupChoices()
	if len(choices) == 0 {
		if field.groups.loading {
			return view
		}
		if strings.TrimSpace(field.input.Value()) == "" {
			return view + "\n  Type a name to add your first group"
		}
		return view + "\n  No matches; enter a valid group name"
	}
	cursor := min(field.groups.cursor, len(choices)-1)
	start := max(0, cursor-3)
	end := min(len(choices), start+4)
	for i := start; i < end; i++ {
		choice := choices[i]
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}
		marker := "[ ] "
		if slices.Contains(field.groups.chosen, choice.name) {
			marker = "[x] "
		}
		if choice.create {
			marker = "+ Add "
		}
		view += "\n" + prefix + marker + choice.name
	}
	if len(choices) > 4 {
		view += fmt.Sprintf("\n  %d–%d of %d", start+1, end, len(choices))
	}
	return view
}
