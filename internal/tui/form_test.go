package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestFormTagsValidationUsesNormalizedDomainRules(t *testing.T) {
	for _, raw := range []string{"Ideas, ideas", "trabajo personal, trabajo-personal", "ideas,", ",ideas", "!!!"} {
		t.Run(raw, func(t *testing.T) {
			f := newInboxForm()
			f.Resize(80, 24)
			f.Focus()
			f.fields[1].input.SetValue(raw)
			f, cmd := f.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			if f.fields[1].problem == "" || f.focus != 1 || !f.fields[1].input.Focused() {
				t.Fatal("invalid tags did not receive an error and focus")
			}
			if cmd != nil {
				if _, submitted := cmd().(formSubmitted); submitted {
					t.Fatal("invalid tags were submitted")
				}
			}
			f.fields[1].input.SetValue("ideas, trabajo personal")
			f, cmd = f.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			if cmd == nil || f.fields[1].problem != "" {
				t.Fatal("corrected tags could not be submitted")
			}
			if _, submitted := cmd().(formSubmitted); !submitted {
				t.Fatal("valid tags were not submitted")
			}
		})
	}
}

func TestFormTextFieldCanBeReusedAndFocusWraps(t *testing.T) {
	f := newForm("Page", fieldSpec{id: "title", label: "Title", kind: fieldText})
	f.Resize(80, 24)
	f.Focus()
	f, _ = f.Update(tea.PasteMsg{Content: "A title"})
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if !f.fields[0].input.Focused() || f.focus != 0 {
		t.Fatal("Tab did not wrap back to the only field")
	}
	f, submit := f.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if submit == nil || submit().(formSubmitted).values["title"] != "A title" {
		t.Fatal("text field lost its value")
	}
}

func TestInboxFormResizesAndKeepsLongMultilineContent(t *testing.T) {
	m := openInbox(t)
	var next tea.Model
	text := strings.Repeat("una línea\n", 120) + "fin"
	next, _ = m.Update(tea.PasteMsg{Content: text})
	m = next.(model)
	for _, size := range []tea.WindowSizeMsg{{Width: 45, Height: 22}, {Width: 100, Height: 40}} {
		next, _ = m.Update(size)
		m = next.(model)
		if m.form.fields[0].Value() != text || m.form.fields[0].textarea.Width() > size.Width || (m.form.fields[0].textarea.Height() < 1 || m.form.fields[0].textarea.Height() > 10) {
			t.Fatal("resizing lost content or did not resize textarea")
		}
		view := ansi.Strip(m.View().Content)
		if !strings.Contains(view, "Tags > Optional") || strings.Index(view, "Content") > strings.Index(view, "Tags >") {
			t.Fatal("content must appear first and tags must keep their inline placeholder")
		}
		for _, removed := range []string{"Save", "Cancel", "Free-form"} {
			if strings.Contains(view, removed) {
				t.Fatalf("form still displays %q", removed)
			}
		}
		for _, want := range []string{"Overmind / Inbox", "Content", "Tags", "ctrl+s", "new line"} {
			if !strings.Contains(view, want) {
				t.Fatalf("form view omitted %q", want)
			}
		}
	}
}

func TestAllFormsKeepSpacingContentAndTagsLast(t *testing.T) {
	for _, f := range []form{newInboxForm(), newHabitForm(), newPersonForm(), newBookmarkForm(), newPageForm(), newJournalForm("2020-01-02")} {
		t.Run(f.title, func(t *testing.T) {
			f.Resize(100, 40)
			f.Focus()
			view := ansi.Strip(f.View())
			for i := 1; i < len(f.fields); i++ {
				if !strings.Contains(view, "\n\n"+f.fields[i].spec.label) {
					t.Fatalf("missing blank line before %s: %s", f.fields[i].spec.label, view)
				}
			}
			if f.fields[len(f.fields)-1].spec.id != "tags" {
				t.Fatal("Tags must be the last input")
			}
			expectedContentHeight := 10
			if f.title == "Page" || f.title == "Journal" || f.title == "Person" {
				expectedContentHeight = 20
			}
			found := false
			for _, field := range f.fields {
				if field.spec.id == "content" {
					found = true
					if field.spec.kind != fieldMultiline || field.textarea.Height() != expectedContentHeight {
						t.Fatal("Content must use the shared large textarea")
					}
				}
			}
			if !found {
				t.Fatal("missing Content")
			}
		})
	}
}
