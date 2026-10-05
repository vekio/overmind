package tui

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/vekio/overmind/internal/app/bookmark"
	"github.com/vekio/overmind/internal/app/page"
)

func TestPageAndBookmarkCreateWithSharedForm(t *testing.T) {
	for _, tc := range []struct {
		action action
		values []string
	}{
		{actionPage, []string{"My page", "work/projects", "First line\nSecond line\n", "ideas, trabajo"}},
		{actionPage, []string{"Another page", "   ", "", ""}},
		{actionBookmark, []string{"https://example.com/article?q=one#two", "Bookmark details\nSecond line\n", "ideas, trabajo"}},
		{actionBookmark, []string{"http://example.com", "", ""}},
	} {
		t.Run(tc.action.String()+strings.Join(tc.values, "|"), func(t *testing.T) {
			next, _ := newModel().begin(tc.action)
			m := next.(model)
			if m.screen != screenForm || !m.form.fields[0].input.Focused() {
				t.Fatal("creation did not open with the required field focused")
			}
			view := ansi.Strip(m.View().Content)
			if !strings.Contains(view, "Tags > Optional") || strings.Index(view, m.form.fields[0].spec.label+" >") > strings.Index(view, "Tags >") || strings.Contains(view, "Demo:") || strings.Contains(view, "Save") || strings.Contains(view, "Cancel") {
				t.Fatalf("creation did not use the shared layout: %q", view)
			}
			calls := 0
			ctx := context.WithValue(context.Background(), struct{}{}, "create")
			m.ctx = ctx
			m.createPage = func(gotCtx context.Context, command page.CreateCommand) (page.CreateResult, error) {
				calls++
				if tc.action != actionPage || gotCtx != ctx {
					t.Fatal("wrong creation handler or context")
				}
				want := page.CreateCommand{Title: tc.values[0], Area: strings.TrimSpace(tc.values[1]), Content: tc.values[2], Tags: formListValues(tc.values[3])}
				if !reflect.DeepEqual(command, want) {
					t.Fatalf("page command=%+v, want=%+v", command, want)
				}
				return page.CreateResult{}, nil
			}
			m.createBookmark = func(gotCtx context.Context, command bookmark.CreateCommand) (bookmark.CreateResult, error) {
				calls++
				if tc.action != actionBookmark || gotCtx != ctx {
					t.Fatal("wrong creation handler or context")
				}
				want := bookmark.CreateCommand{URL: tc.values[0], Content: tc.values[1], Tags: formListValues(tc.values[2])}
				if !reflect.DeepEqual(command, want) {
					t.Fatalf("bookmark command=%+v, want=%+v", command, want)
				}
				return bookmark.CreateResult{}, nil
			}
			for _, value := range tc.values {
				next, _ = m.Update(tea.PasteMsg{Content: value})
				m = next.(model)
				next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				m = next.(model)
			}
			if m.form.focus != 0 || calls != 0 {
				t.Fatal("Tab did not wrap across fields without saving")
			}
			next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			m = next.(model)
			next, save := m.Update(submit())
			m = next.(model)
			if !m.form.saving || save == nil || calls != 0 {
				t.Fatal("save was not asynchronous")
			}
			next, _ = m.Update(save())
			m = next.(model)
			if calls != 1 || m.screen != screenMenu || m.notification.text != tc.action.String()+" created" {
				t.Fatal("creation did not return to menu with the correct notification")
			}
		})
	}
}

func TestPageAndBookmarkValidationFocusesInvalidField(t *testing.T) {
	for _, tc := range []struct {
		action action
		values []string
		field  int
	}{
		{actionPage, []string{"", "", "", ""}, 0},
		{actionPage, []string{"!!!", "", "", ""}, 0},
		{actionPage, []string{"Title", "work//projects", "", ""}, 1},
		{actionPage, []string{"Title", "", "", "Ideas, ideas"}, 3},
		{actionBookmark, []string{"", "", ""}, 0},
		{actionBookmark, []string{"example.com", "", ""}, 0},
		{actionBookmark, []string{"ftp://example.com", "", ""}, 0},
		{actionBookmark, []string{"https://example.com", "", "Ideas, ideas"}, 2},
	} {
		t.Run(tc.action.String()+strings.Join(tc.values, "|"), func(t *testing.T) {
			next, _ := newModel().begin(tc.action)
			m := next.(model)
			for i, value := range tc.values {
				setFormValue(&m.form, m.form.fields[i].spec.id, value)
			}
			next, cmd := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			m = next.(model)
			if m.form.saving || m.form.focus != tc.field || !m.form.fields[tc.field].input.Focused() || m.form.fields[tc.field].problem == "" {
				t.Fatal("validation did not focus the invalid field")
			}
			if cmd != nil {
				if _, submitted := cmd().(formSubmitted); submitted {
					t.Fatal("invalid form was submitted")
				}
			}
		})
	}
}

func TestPageAndBookmarkFailureRetryAndCancel(t *testing.T) {
	for _, selected := range []action{actionPage, actionBookmark} {
		t.Run(selected.String(), func(t *testing.T) {
			next, _ := newModel().begin(selected)
			m := next.(model)
			value := "My page"
			if selected == actionBookmark {
				value = "https://example.com"
			}
			m.form.fields[0].input.SetValue(value)
			// A missing callback is reported without losing the input.
			next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			m = next.(model)
			next, save := m.Update(submit())
			m = next.(model)
			next, _ = m.Update(noteCreated{action: actionInbox})
			m = next.(model)
			if !m.form.saving {
				t.Fatal("completion from another action changed the form")
			}
			next, _ = m.Update(save())
			m = next.(model)
			if m.form.saving || m.screen != screenForm || m.form.fields[0].Value() != value || !strings.Contains(m.form.problem, "not configured") {
				t.Fatal("failure did not preserve the form")
			}
			m.createPage = func(context.Context, page.CreateCommand) (page.CreateResult, error) {
				return page.CreateResult{}, errors.New("cannot save page")
			}
			m.createBookmark = func(context.Context, bookmark.CreateCommand) (bookmark.CreateResult, error) {
				return bookmark.CreateResult{}, errors.New("cannot save bookmark")
			}
			next, submit = m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			m = next.(model)
			next, save = m.Update(submit())
			m = next.(model)
			next, _ = m.Update(save())
			m = next.(model)
			if m.form.fields[0].Value() != value || !strings.Contains(m.View().Content, "cannot save") {
				t.Fatal("retry lost the input or omitted the service error")
			}
			next, cancel := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
			m = next.(model)
			next, _ = m.Update(cancel())
			m = next.(model)
			if m.screen != screenMenu || len(m.form.fields) != 0 {
				t.Fatal("cancel did not discard the form")
			}
		})
	}
}
