package tui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type notificationKind uint8

const (
	notificationInfo notificationKind = iota
	notificationSuccess
	notificationError
)

const (
	notificationDuration      = 3 * time.Second
	errorNotificationDuration = 5 * time.Second
)

type notification struct {
	id   uint64
	kind notificationKind
	text string
}

type dismissNotification struct {
	id uint64
}

var (
	notificationInfoStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#7571F9"))
	notificationSuccessStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFFFFF")).
					Background(lipgloss.Color("#02BA84"))
	notificationErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#ED567A"))
)

func (n notification) View(width int) string {
	if width <= 0 {
		return ""
	}
	var icon string
	var style lipgloss.Style
	switch n.kind {
	case notificationSuccess:
		icon, style = "✓ ", notificationSuccessStyle
	case notificationError:
		icon, style = "✕ ", notificationErrorStyle
	default:
		icon, style = "• ", notificationInfoStyle
	}
	message := strings.Join(strings.Fields(n.text), " ")
	line := ansi.Truncate(icon+message, width, "…")
	line += strings.Repeat(" ", max(0, width-lipgloss.Width(line)))
	return style.Render(line)
}

func (m *model) notify(text string, kind notificationKind) tea.Cmd {
	m.notification.id++
	m.notification.kind = kind
	m.notification.text = text

	duration := notificationDuration
	if kind == notificationError {
		duration = errorNotificationDuration
	}
	id := m.notification.id
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return dismissNotification{id: id}
	})
}
