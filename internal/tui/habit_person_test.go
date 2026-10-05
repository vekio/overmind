package tui

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/vekio/overmind/internal/app/habit"
	"github.com/vekio/overmind/internal/app/person"
)

func TestHabitAndPersonCreation(t *testing.T) {
	for _, selected := range []action{actionHabit, actionPerson} {
		t.Run(selected.String(), func(t *testing.T) {
			next, _ := newModel().begin(selected)
			m := next.(model)
			ctx := context.WithValue(context.Background(), struct{}{}, selected)
			m.ctx = ctx
			calls := 0
			m.createHabit = func(gotCtx context.Context, command habit.CreateCommand) (habit.CreateResult, error) {
				calls++
				want := habit.CreateCommand{Title: "Exercise", Amount: 2.5, Unit: "hours", Period: "week", Content: "Habit details\nSecond line\n", Tags: []string{"health", "personal"}}
				if gotCtx != ctx || !reflect.DeepEqual(command, want) {
					t.Fatalf("unexpected habit command: %+v", command)
				}
				return habit.CreateResult{}, nil
			}
			m.createPerson = func(gotCtx context.Context, command person.CreateCommand) (person.CreateResult, error) {
				calls++
				want := person.CreateCommand{Name: "Alberto", Groups: []string{"work", "friends"}, Content: "Person details\nSecond line\n", Tags: []string{"people"}}
				if gotCtx != ctx || !reflect.DeepEqual(command, want) {
					t.Fatalf("unexpected person command: %+v", command)
				}
				return person.CreateResult{}, nil
			}
			values := []string{"Alberto", " work, friends ", "Person details\nSecond line\n", "people"}
			if selected == actionHabit {
				values = []string{"Exercise", "2.5", "hours", "", "Habit details\nSecond line\n", " health, personal "}
			}
			for _, value := range values {
				if m.form.fields[m.form.focus].spec.kind == fieldGroups {
					for _, name := range formListValues(value) {
						next, _ = m.Update(tea.PasteMsg{Content: name})
						m = next.(model)
						next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
						m = next.(model)
					}
					value = ""
				}
				next, _ = m.Update(tea.PasteMsg{Content: value})
				m = next.(model)
				if m.form.fields[m.form.focus].spec.kind == fieldSelect {
					next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
					m = next.(model)
				}
				next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				m = next.(model)
			}
			next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			m = next.(model)
			if submit == nil {
				t.Fatal("valid form did not submit")
			}
			next, save := m.Update(submit())
			m = next.(model)
			if save == nil || calls != 0 || !m.form.saving {
				t.Fatal("save must be asynchronous")
			}
			next, _ = m.Update(save())
			m = next.(model)
			if calls != 1 || m.screen != screenMenu || m.notification.text != selected.String()+" created" {
				t.Fatal("creation did not complete")
			}
		})
	}
}

func TestHabitPeriodSelection(t *testing.T) {
	f := newHabitForm()
	f.Resize(80, 24)
	f.Focus()
	for range 3 {
		f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	if f.focus != 3 || !reflect.DeepEqual(f.Controls()[3], controlHint{"←/→", "choose"}) {
		t.Fatal("period selector is not focusable")
	}
	for _, expected := range []string{"week", "month", "day"} {
		f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyRight})
		if f.fields[3].Value() != expected {
			t.Fatal("period did not advance or wrap")
		}
	}
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	f, _ = f.Update(tea.PasteMsg{Content: "year"})
	if f.fields[3].Value() != "month" {
		t.Fatal("period must only allow supported choices")
	}
	for _, key := range []rune{'h', 'j', 'k', 'l'} {
		f, _ = f.Update(tea.KeyPressMsg{Code: key, Text: string(key)})
		if f.fields[3].Value() != "month" {
			t.Fatalf("letter %c changed the period", key)
		}
	}
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if f.focus != 4 || !f.fields[4].textarea.Focused() {
		t.Fatal("Tab did not leave selector")
	}
}

func TestHabitAndPersonValidation(t *testing.T) {
	for _, raw := range []string{"", "0", "-1", "NaN", "+Inf", "abc"} {
		t.Run("amount="+raw, func(t *testing.T) {
			f := newHabitForm()
			f.fields[0].input.SetValue("Exercise")
			f.fields[1].input.SetValue(raw)
			f.fields[2].input.SetValue("hours")
			f, cmd := f.submit()
			if f.focus != 1 || f.fields[1].problem == "" {
				t.Fatal("invalid amount was not rejected")
			}
			if cmd != nil {
				if _, ok := cmd().(formSubmitted); ok {
					t.Fatal("invalid amount submitted")
				}
			}
		})
	}
	for _, raw := range []string{"Work, work", "work,", ",work", "!!!"} {
		t.Run("groups="+raw, func(t *testing.T) {
			f := newPersonForm()
			f.fields[0].input.SetValue("Alberto")
			f.fields[1].input.SetValue(raw)
			f, _ = f.submit()
			if f.focus != 1 || f.fields[1].problem == "" {
				t.Fatal("invalid groups were not rejected")
			}
		})
	}
	for _, f := range []form{newHabitForm(), newPersonForm()} {
		f, _ = f.submit()
		if f.focus != 0 || f.fields[0].problem == "" {
			t.Fatal("required title/name was not rejected")
		}
	}
	f := newPersonForm()
	f.fields[0].input.SetValue("Alberto")
	_, submit := f.submit()
	if submit == nil {
		t.Fatal("empty optional groups and tags should be allowed")
	}
}

func TestHabitAndPersonFailurePreservesDraft(t *testing.T) {
	for _, selected := range []action{actionHabit, actionPerson} {
		t.Run(selected.String(), func(t *testing.T) {
			next, _ := newModel().begin(selected)
			m := next.(model)
			m.form.fields[0].input.SetValue("Draft")
			if selected == actionHabit {
				m.form.fields[1].input.SetValue("1")
				m.form.fields[2].input.SetValue("hours")
			}
			for attempt := range 2 {
				next, submit := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
				m = next.(model)
				next, save := m.Update(submit())
				m = next.(model)
				next, _ = m.Update(save())
				m = next.(model)
				want := "not configured"
				if attempt == 1 {
					want = "cannot save"
				}
				if m.screen != screenForm || m.form.saving || m.form.fields[0].Value() != "Draft" || !strings.Contains(m.form.problem, want) {
					t.Fatal("failure lost draft or omitted error")
				}
				m.createHabit = func(context.Context, habit.CreateCommand) (habit.CreateResult, error) {
					return habit.CreateResult{}, errors.New("cannot save")
				}
				m.createPerson = func(context.Context, person.CreateCommand) (person.CreateResult, error) {
					return person.CreateResult{}, errors.New("cannot save")
				}
			}
			next, cancel := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
			m = next.(model)
			next, _ = m.Update(cancel())
			m = next.(model)
			if m.screen != screenMenu || len(m.form.fields) != 0 {
				t.Fatal("cancel did not discard draft")
			}
		})
	}
}
