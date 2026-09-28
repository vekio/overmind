package tui

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
)

type journalClient struct {
	Client
	id          uuid.UUID
	exists      bool
	createErr   error
	findCalls   int
	createCalls int
	openCalls   int
}

func (client *journalClient) FindJournal(context.Context, domain.Date) (uuid.UUID, bool, error) {
	client.findCalls++
	if client.findCalls > 1 && errors.Is(client.createErr, app.ErrJournalAlreadyExists) {
		return client.id, true, nil
	}
	return client.id, client.exists, nil
}

func (client *journalClient) CreateJournal(context.Context, domain.Tags) (app.CreateJournalResult, error) {
	client.createCalls++
	if client.createErr != nil {
		return app.CreateJournalResult{}, client.createErr
	}
	journal, err := domain.NewJournal(client.id, domain.DateFromTime(time.Now()), domain.Tags{}, time.Now())
	return app.CreateJournalResult{Journal: journal, Path: "/vault/journal.adoc"}, err
}

func (client *journalClient) OpenNote(_ context.Context, id uuid.UUID) ([]byte, error) {
	client.openCalls++
	if id != client.id {
		return nil, errors.New("wrong journal ID")
	}
	return []byte("journal source"), nil
}

func TestJournalOpensExistingCreatesMissingAndHandlesCreateRace(t *testing.T) {
	for _, tc := range []struct {
		name        string
		exists      bool
		createErr   error
		wantCreates int
		wantFinds   int
		wantPath    string
	}{
		{"existing", true, nil, 0, 1, ""},
		{"new", false, nil, 1, 1, "/vault/journal.adoc"},
		{"created concurrently", false, app.ErrJournalAlreadyExists, 1, 2, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &journalClient{id: uuid.MustParse("11111111-1111-4111-8111-111111111111"), exists: tc.exists, createErr: tc.createErr}
			m := newModel(context.Background(), client)
			_, command := m.startJournal()
			message, ok := command().(noteOpenResult)
			if !ok || message.err != nil || message.id != client.id || message.createdPath != tc.wantPath || string(message.source) != "journal source" {
				t.Fatalf("journal result = %+v", message)
			}
			if client.createCalls != tc.wantCreates || client.findCalls != tc.wantFinds || client.openCalls != 1 {
				t.Fatalf("calls: create=%d find=%d open=%d", client.createCalls, client.findCalls, client.openCalls)
			}
		})
	}
}
