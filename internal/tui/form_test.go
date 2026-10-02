package tui

import (
	"context"
	"reflect"
	"testing"
	"time"
	"uuid"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
)

type formClient struct {
	Client
	id           uuid.UUID
	pageTitle    domain.Title
	pageArea     domain.Area
	pageTags     domain.Tags
	openedID     uuid.UUID
	created      int
	personName   domain.Title
	personGroups domain.Groups
}

func (client *formClient) CreatePage(_ context.Context, title domain.Title, area domain.Area, tags domain.Tags) (app.CreatePageResult, error) {
	client.created++
	client.pageTitle, client.pageArea, client.pageTags = title, area, tags
	page, err := domain.NewPage(client.id, title, area, tags, time.Now())
	return app.CreatePageResult{Page: page, Path: "/vault/page.adoc"}, err
}

func (client *formClient) OpenNote(_ context.Context, id uuid.UUID) ([]byte, error) {
	client.openedID = id
	return []byte("= Page\n\nOriginal\n"), nil
}

func (client *formClient) CreatePerson(_ context.Context, name domain.Title, groups domain.Groups, tags domain.Tags) (app.CreatePersonResult, error) {
	client.created++
	client.personName, client.personGroups = name, groups
	person, err := domain.NewPerson(client.id, name, groups, tags, time.Now())
	return app.CreatePersonResult{Person: person, Path: "/vault/person.adoc"}, err
}

func TestPersonFormValidatesGroupsAndCreatesForEditing(t *testing.T) {
	client := &formClient{id: uuid.New()}
	m := newModel(context.Background(), client)
	next, _ := m.begin(actionPerson)
	m = next.(model)
	m.formInputs[0].SetValue("Ana García")
	m.formInputs[1].SetValue("Work, work")
	next, command := m.submitForm(true)
	m = next.(model)
	if m.screen != screenForm || m.formFocus != 1 || m.problem == "" || client.created != 0 {
		t.Fatal("duplicate groups did not prevent creation")
	}
	m.formInputs[1].SetValue("Work, University")
	next, command = m.submitForm(true)
	if next.(model).screen != screenBusy || command == nil {
		t.Fatal("person form did not start creation")
	}
	opened, ok := command().(noteOpenResult)
	if !ok || opened.err != nil || opened.id != client.id || opened.createdPath != "/vault/person.adoc" {
		t.Fatalf("created person editor result = %+v", opened)
	}
	if client.personName.String() != "Ana García" || !reflect.DeepEqual(client.personGroups.Strings(), []string{"work", "university"}) {
		t.Fatalf("person inputs = %q, %v", client.personName, client.personGroups.Strings())
	}
	if got := noteName(app.ListedNote{Kind: domain.NoteKindPerson, Attributes: map[string]string{"name": "Ana García", "groups": "work, university"}}); got != "Ana García [work, university]" {
		t.Fatalf("person display = %q", got)
	}
}

func TestPageFormValidatesAndCanCreateForEditing(t *testing.T) {
	client := &formClient{id: uuid.MustParse("11111111-1111-4111-8111-111111111111")}
	m := newModel(context.Background(), client)
	m.action = actionPage
	next, _ := m.startForm([]formField{{"Title", "title"}, {"Area", "area"}, {"Tags", "tags"}})
	m = next.(model)
	m.formInputs[0].SetValue(" ")
	next, command := m.submitForm(true)
	m = next.(model)
	if m.screen != screenForm || m.formFocus != 0 || m.problem == "" || client.created != 0 {
		t.Fatalf("invalid form state: screen=%d focus=%d problem=%q created=%d", m.screen, m.formFocus, m.problem, client.created)
	}
	m.formInputs[0].SetValue("Project plan")
	m.formInputs[1].SetValue("Work/Ideas")
	m.formInputs[2].SetValue("one, two")
	next, command = m.submitForm(true)
	m = next.(model)
	if m.screen != screenBusy || command == nil {
		t.Fatal("valid page form did not start creation")
	}
	opened, ok := command().(noteOpenResult)
	if !ok || opened.err != nil || opened.id != client.id || opened.createdPath != "/vault/page.adoc" {
		t.Fatalf("created page editor result = %+v", opened)
	}
	if client.pageTitle.String() != "Project plan" || client.pageArea.String() != "work/ideas" || !reflect.DeepEqual(client.pageTags.Strings(), []string{"one", "two"}) || client.openedID != client.id {
		t.Fatalf("page inputs = %q, %q, %v, opened %s", client.pageTitle, client.pageArea, client.pageTags.Strings(), client.openedID)
	}
}
