package tui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	appconfig "github.com/vekio/overmind/internal/config"
)

func TestSetupShowsEditableDefaultsAndSubmitsNormalizedSettings(t *testing.T) {
	root := t.TempDir()
	defaults := appconfig.Settings{Mode: appconfig.ModeLocal, VaultPath: "/tmp/vault"}
	m := newSetupModel(defaults, filepath.Join(root, "config.yml"))
	m.Init()
	view := ansi.Strip(m.View().Content)
	for _, text := range []string{"Overmind / Setup", "Mode > local", "Vault > /tmp/vault", "Config >", "ctrl+s", "esc"} {
		if !strings.Contains(view, text) {
			t.Fatalf("setup omitted %q: %s", text, view)
		}
	}
	if !m.form.fields[0].input.Focused() {
		t.Fatal("setup did not focus the first input")
	}
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = next.(setupModel)
	if !m.form.fields[1].input.Focused() {
		t.Fatal("Tab did not focus Vault")
	}
	m.form.fields[1].input.SetValue("  " + filepath.Join(root, "edited") + "/  ")
	next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = next.(setupModel)
	if submit == nil || m.submitted {
		t.Fatal("setup should only complete after receiving submission")
	}
	next, quit := m.Update(submit())
	m = next.(setupModel)
	if !m.submitted || m.settings.Mode != appconfig.ModeLocal || m.settings.VaultPath != filepath.Join(root, "edited") {
		t.Fatalf("unexpected settings: %+v", m.settings)
	}
	if _, ok := quit().(tea.QuitMsg); !ok {
		t.Fatal("submitted setup did not close")
	}
}

func TestSetupValidationKeepsFormAndFocusesInvalidField(t *testing.T) {
	for _, tc := range []struct {
		mode, vault string
		field       int
	}{
		{"api", t.TempDir(), 0},
		{"", t.TempDir(), 0},
		{"local", " ", 1},
		{"local", "~someone/vault", 1},
	} {
		t.Run(tc.mode+tc.vault, func(t *testing.T) {
			m := newSetupModel(appconfig.Settings{Mode: appconfig.ModeLocal, VaultPath: t.TempDir()}, "/tmp/config.yml")
			setFormValue(&m.form, "mode", tc.mode)
			setFormValue(&m.form, "vault", tc.vault)
			next, cmd := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			m = next.(setupModel)
			if m.submitted || m.form.focus != tc.field || m.form.fields[tc.field].problem == "" {
				t.Fatal("invalid input did not remain editable with error and focus")
			}
			if cmd != nil {
				if _, ok := cmd().(formSubmitted); ok {
					t.Fatal("invalid setup submitted")
				}
			}
		})
	}
}

func TestSetupCancellationDoesNotSubmit(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{{Code: tea.KeyEscape}, {Code: 'c', Mod: tea.ModCtrl}} {
		m := newSetupModel(appconfig.Settings{VaultPath: t.TempDir()}, "/tmp/config.yml")
		next, cmd := m.Update(key)
		m = next.(setupModel)
		message := cmd()
		if _, cancelled := message.(formCancelled); cancelled {
			next, cmd = m.Update(message)
			m = next.(setupModel)
			message = cmd()
		}
		if _, quit := message.(tea.QuitMsg); !quit || m.submitted {
			t.Fatal("cancellation submitted configuration or failed to quit")
		}
	}
}
