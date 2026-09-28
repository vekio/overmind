package tui

import (
	"testing"

	"charm.land/bubbletea/v2"
)

func TestAskUsesDefaultArrowsAndShortcuts(t *testing.T) {
	a := newAsk("Delete?", cancelOption,
		askOption{id: deleteOption, label: "Yes", shortcut: "y"},
		askOption{id: cancelOption, label: "No", shortcut: "n"},
	)
	if a.options[a.selected].id != cancelOption {
		t.Fatal("unsafe default answer")
	}
	_, answerCmd := a.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if answerCmd().(askAnswer).option != cancelOption {
		t.Fatal("enter did not use selected default")
	}
	a, _ = a.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	_, answerCmd = a.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if answerCmd().(askAnswer).option != deleteOption {
		t.Fatal("left arrow did not select Yes")
	}
	_, answerCmd = a.Update(tea.KeyPressMsg{Code: 'n'})
	if answerCmd().(askAnswer).option != cancelOption {
		t.Fatal("No shortcut did not override selection")
	}
}
