package gotemplate

import (
	"context"
	"testing"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
)

func TestRendererRendersPageTemplate(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	title, err := domain.NewTitle("My page")
	if err != nil {
		t.Fatalf("NewTitle() error = %v", err)
	}
	area, err := domain.NewArea("Knowledge")
	if err != nil {
		t.Fatalf("NewArea() error = %v", err)
	}
	id, err := domain.NewDocumentID("page-id")
	if err != nil {
		t.Fatalf("NewDocumentID() error = %v", err)
	}
	createdAt := time.Date(2026, time.August, 14, 10, 30, 0, 0, time.FixedZone("CEST", 2*60*60))
	page := domain.NewPage(id, title, area, createdAt)
	content, err := renderer.Render(context.Background(), "page", page)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got, want := string(content), "= My page\n:overmind-id: page-id\n:overmind-type: page\n:overmind-title: My page\n:overmind-area: knowledge\n:overmind-created-at: 2026-08-14T08:30:00Z\n\n"; got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRendererRendersPageWithoutArea(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	title, err := domain.NewTitle("My page")
	if err != nil {
		t.Fatalf("NewTitle() error = %v", err)
	}
	id, err := domain.NewDocumentID("page-id")
	if err != nil {
		t.Fatalf("NewDocumentID() error = %v", err)
	}
	createdAt := time.Date(2026, time.August, 14, 10, 30, 0, 0, time.UTC)
	page := domain.NewPage(id, title, domain.Area{}, createdAt)
	content, err := renderer.Render(context.Background(), "page", page)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got, want := string(content), "= My page\n:overmind-id: page-id\n:overmind-type: page\n:overmind-title: My page\n:overmind-area:\n:overmind-created-at: 2026-08-14T10:30:00Z\n\n"; got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}
