package tui

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/domain"
)

type captureClient struct {
	Client
	id          uuid.UUID
	captureText string
	openCalls   int
}

func (client *captureClient) Capture(_ context.Context, content string) (app.CaptureResult, error) {
	client.captureText = content
	inbox, err := domain.NewInbox(client.id, content, time.Now())
	return app.CaptureResult{Inbox: inbox, Path: "/vault/inbox.adoc"}, err
}

func (client *captureClient) OpenNote(_ context.Context, id uuid.UUID) ([]byte, error) {
	client.openCalls++
	if id != client.id {
		return nil, errors.New("wrong capture ID")
	}
	return []byte("capture source"), nil
}

func TestCaptureCreatesEmptyNoteBeforeOpeningEditor(t *testing.T) {
	client := &captureClient{id: uuid.MustParse("11111111-1111-4111-8111-111111111111")}
	m := newModel(context.Background(), client)
	_, command := m.startCapture()
	message, ok := command().(noteOpenResult)
	if !ok || message.err != nil || message.id != client.id || message.createdPath != "/vault/inbox.adoc" || client.captureText != "" || client.openCalls != 1 {
		t.Fatalf("capture result = %+v, capture text %q, open calls %d", message, client.captureText, client.openCalls)
	}
}
