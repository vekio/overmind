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
	goTag, err := domain.NewTag("Go")
	if err != nil {
		t.Fatalf("NewTag() error = %v", err)
	}
	domainDesignTag, err := domain.NewTag("Diseño de dominio")
	if err != nil {
		t.Fatalf("NewTag() error = %v", err)
	}
	tags, err := domain.NewTags(goTag, domainDesignTag)
	if err != nil {
		t.Fatalf("NewTags() error = %v", err)
	}
	createdAt := time.Date(2026, time.August, 14, 10, 30, 0, 0, time.FixedZone("CEST", 2*60*60))
	page := domain.NewPage(id, title, tags, area, createdAt)
	content, err := renderer.Render(context.Background(), "page", page)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got, want := string(content), "= My page\n:overmind-id: page-id\n:overmind-type: page\n:overmind-area: knowledge\n:overmind-tags: go, diseno-de-dominio\n:overmind-created-at: 2026-08-14T08:30:00Z\n\n"; got != want {
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
	page := domain.NewPage(id, title, domain.Tags{}, domain.Area{}, createdAt)
	content, err := renderer.Render(context.Background(), "page", page)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got, want := string(content), "= My page\n:overmind-id: page-id\n:overmind-type: page\n:overmind-area:\n:overmind-tags:\n:overmind-created-at: 2026-08-14T10:30:00Z\n\n"; got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}
