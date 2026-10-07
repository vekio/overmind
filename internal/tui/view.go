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

// View renders the active screen and its contextual controls in the alternate screen.
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
		if m.reindexing {
			lines.WriteString("\nReindexing notes…")
		}
		content = lines.String()
		controls = []controlHint{{"↑/↓", "move"}, {"enter", "select"}, {"esc/q", "quit"}}
	case screenAsk:
		content = m.ask.View()
		controls = append(m.ask.Controls(), controlHint{"esc/q", "back"})
	case screenForm:
		content = m.form.View()
		controls = m.form.Controls()
	case screenRawEdit:
		content = "Overmind / Raw edit\n\n" + m.rawStatus()
		controls = m.rawControls()
	case screenEdit:
		content = "Overmind / Edit " + m.action.String() + "\n\nLoading…"
		controls = []controlHint{{"esc", "back"}}
	case screenJournal:
		content = m.journalView()
		controls = []controlHint{{"r", "refresh"}, {"esc/q", "back"}}
	case screenNotes:
		notes := m.notes
		if m.problem != "" {
			notes.SetHeight(max(4, notesTableHeight(m.height, m.noteCount)-2))
		}
		status := fmt.Sprintf("%d notes · page %d · %s", m.noteCount, m.offset/notesPageSize+1, m.filterSummary())
		if m.loading {
			status += " · searching…"
		}
		if m.deleting {
			status += " · deleting…"
		}
		content = fmt.Sprintf("Overmind / Notes (%s)\n\n%s", status, notes.View())
		if m.noteCount == 0 && !m.loading && m.problem == "" {
			content += "\n\nNo notes found."
		}
		controls = []controlHint{{"↑/↓", "move"}, {"enter/e", "edit"}, {"ctrl+e", "raw edit"}, {"d", "delete"}, {"f", "filter"}, {"r", "refresh"}, {"n/p", "page"}, {"esc/q", "back"}}
	}
	if m.problem != "" {
		content += "\n\nError: " + m.problem
	}
	if len(controls) > 0 {
		// Keep the notification row reserved so controls do not move when it appears.
		footerHeight := m.height - 1
		footer := renderControls(controls, m.width)
		if m.screen == screenForm || m.screen == screenRawEdit {
			footer = ansi.Wrap(formatControls(controls), max(1, m.width), "")
		}
		content = withFooter(content, footer, footerHeight)
		content += "\n"
		if m.notification.text != "" {
			content += m.notification.View(m.width)
		}
	}
	if m.filterOpen {
		modal := m.filterView()
		canvas := lipgloss.NewCanvas(max(1, m.width), max(1, m.height))
		canvas.Compose(lipgloss.NewCompositor(
			lipgloss.NewLayer(content),
			lipgloss.NewLayer(modal).X(max(0, (m.width-lipgloss.Width(modal))/2)).Y(max(0, (m.height-lipgloss.Height(modal))/2)).Z(1),
		))
		content = canvas.Render()
	}
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func withFooter(content, footer string, height int) string {
	content = strings.TrimRight(content, "\n")
	lines := 1 + strings.Count(content, "\n")
	footerLines := 1 + strings.Count(footer, "\n")
	gap := max(2, height-lines-footerLines+1)
	return content + strings.Repeat("\n", gap) + footer
}

func renderControls(controls []controlHint, width int) string {
	return ansi.Truncate(formatControls(controls), max(0, width), "…")
}

func formatControls(controls []controlHint) string {
	items := make([]string, 0, len(controls))
	for _, control := range controls {
		items = append(items,
			controlKeyStyle.Render(control.key)+" "+
				controlDescriptionStyle.Render(strings.ToLower(control.description)),
		)
	}
	return strings.Join(items, controlSeparatorStyle.Render("  ·  "))
}
