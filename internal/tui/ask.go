package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type askOptionID string

type askOption struct {
	id       askOptionID
	label    string
	icon     string
	shortcut string
}

type ask struct {
	question string
	options  []askOption
	selected int
}

type askAnswer struct {
	option askOptionID
}

var (
	askQuestionStyle = lipgloss.NewStyle().Bold(true)
	askSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#F780E2")).
				Padding(0, 2)
	askOptionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("237")).
			Padding(0, 2)
)

func newAsk(question string, defaultOption askOptionID, options ...askOption) ask {
	a := ask{
		question: question,
		options:  options,
	}
	for index, option := range a.options {
		if option.id == defaultOption {
			a.selected = index
			break
		}
	}
	return a
}

// Update wraps option selection and emits an answer only on confirmation or a shortcut.
func (a ask) Update(message tea.Msg) (ask, tea.Cmd) {
	key, ok := message.(tea.KeyPressMsg)
	if !ok {
		return a, nil
	}

	pressed := key.String()
	if len(a.options) == 0 {
		return a, nil
	}
	switch pressed {
	case "left", "h":
		a.selected = (a.selected - 1 + len(a.options)) % len(a.options)
		return a, nil
	case "right", "l", "tab":
		a.selected = (a.selected + 1) % len(a.options)
		return a, nil
	case "enter":
		return a, answer(a.options[a.selected].id)
	}

	for _, option := range a.options {
		if option.shortcut != "" && strings.EqualFold(pressed, option.shortcut) {
			return a, answer(option.id)
		}
	}
	return a, nil
}

// View renders the question and highlights the selected option.
func (a ask) View() string {
	buttons := make([]string, 0, len(a.options))
	for index, option := range a.options {
		label := strings.TrimSpace(option.icon + " " + option.label)
		style := askOptionStyle
		if index == a.selected {
			style = askSelectedStyle
		}
		buttons = append(buttons, style.Render(label))
	}
	return askQuestionStyle.Render(a.question) + "\n\n" + strings.Join(buttons, "   ")
}

// Controls returns navigation hints and configured option shortcuts.
func (a ask) Controls() []controlHint {
	controls := []controlHint{{"←/→", "choose"}, {"enter", "confirm"}}
	for _, option := range a.options {
		if option.shortcut != "" {
			controls = append(controls, controlHint{option.shortcut, strings.ToLower(option.label)})
		}
	}
	return controls
}

func answer(option askOptionID) tea.Cmd {
	return func() tea.Msg { return askAnswer{option: option} }
}
