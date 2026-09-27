package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m model) View() tea.View {
	var content string
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
		lines.WriteString("\n↑/↓: choose  ·  Enter: select  ·  q: quit")
		content = lines.String()
	case screenForm:
		content = fmt.Sprintf("Overmind / %s\n\n%s (%d/%d)\n%s\n\nEnter: next/save  ·  Esc: back",
			m.action, m.fields[m.fieldIndex].label, m.fieldIndex+1, len(m.fields), m.input.View())
	case screenCapture:
		content = "Overmind / Capture\n\n" + m.editor.View() + "\n\nCtrl+S: save  ·  Esc: back"
	case screenConfirm:
		content = "Rebuild the index from notes in the vault?\n\ny: rebuild  ·  n/Esc: back"
	case screenBusy:
		content = "Working…"
	case screenNotes:
		content = fmt.Sprintf("Overmind / Notes (%d)\n\n%s\n\n↑/↓: browse  ·  r: refresh  ·  Esc: back", m.noteCount, m.notes.View())
		if m.noteCount == 0 {
			content += "\n\nNo notes indexed yet. Create one or rebuild the index."
		}
	}
	if m.status != "" && m.screen == screenMenu {
		content += "\n\n" + m.status
	}
	if m.problem != "" {
		content += "\n\nError: " + m.problem
	}
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}
