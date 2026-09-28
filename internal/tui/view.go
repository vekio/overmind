package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type controlHint struct {
	key         string
	description string
}

var (
	controlKeyStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	controlDescriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	controlSeparatorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

func (m model) View() tea.View {
	var content string
	var controls []controlHint
	switch m.screen {
	case screenMenu:
		var lines strings.Builder
		lines.WriteString("Overmind\n\n")
		for index, item := range menuItems {
			marker := "  "
			if index == m.selected {
				marker = "> "
			}
			fmt.Fprintf(&lines, "%s%-16s %s\n", marker, item.title, item.description)
		}
		content = lines.String()
		controls = []controlHint{{"↑/↓", "move"}, {"enter", "select"}, {"esc/q", "quit"}}
	case screenForm:
		content = m.formView()
		controls = []controlHint{{"tab/↑/↓", "field"}, {"enter", "create"}, {"ctrl+e", "create & edit"}, {"esc", "back"}}
	case screenAsk:
		content = m.ask.View()
		controls = append(m.ask.Controls(), controlHint{"esc/q", "back"})
	case screenBusy:
		content = "Working…"
		controls = []controlHint{{"ctrl+c", "quit"}}
	case screenNotes:
		notes := m.notes
		if m.problem != "" {
			notes.SetHeight(max(4, notesTableHeight(m.height, m.noteCount)-2))
		}
		content = fmt.Sprintf("Overmind / Notes (%d)\n\n%s", m.noteCount, notes.View())
		if m.noteCount == 0 {
			content += "\n\nNo notes indexed yet. Create one or rebuild the index."
		}
		controls = []controlHint{{"↑/↓", "move"}, {"enter/e", "edit"}, {"r", "refresh"}, {"esc/q", "back"}}
	}
	if m.problem != "" {
		content += "\n\nError: " + m.problem
	}
	if len(controls) > 0 {
		content = withFooter(content, renderControls(controls, m.width), m.height)
		if m.notification.text != "" {
			content += "\n" + m.notification.View(m.width)
		}
	}
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func withFooter(content, footer string, height int) string {
	content = strings.TrimRight(content, "\n")
	lines := 1 + strings.Count(content, "\n")
	gap := max(2, height-lines-1)
	return content + strings.Repeat("\n", gap) + footer
}

func renderControls(controls []controlHint, width int) string {
	items := make([]string, 0, len(controls))
	for _, control := range controls {
		items = append(items,
			controlKeyStyle.Render(control.key)+" "+
				controlDescriptionStyle.Render(strings.ToLower(control.description)),
		)
	}
	return ansi.Truncate(strings.Join(items, controlSeparatorStyle.Render("  ·  ")), max(0, width), "…")
}
