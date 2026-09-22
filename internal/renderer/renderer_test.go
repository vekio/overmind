package renderer_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"git.casta.me/alberto/overmind/internal/renderer"
)

func TestRenderPage(t *testing.T) {
	documentRenderer := newRenderer(t)
	createdAt := time.Date(2026, time.September, 22, 20, 30, 0, 0, time.FixedZone("CEST", 2*60*60))
	updatedAt := createdAt.Add(15 * time.Minute)

	content, err := documentRenderer.Render(context.Background(), renderer.PageTemplate, renderer.Page{
		ID:        "4f962271-9927-4b4c-9301-4dd87a671d14",
		Title:     "Sillita bebé",
		Area:      "bebé",
		Tags:      []string{"silla", "compras"},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := "= Sillita bebé\n" +
		":overmind-id: 4f962271-9927-4b4c-9301-4dd87a671d14\n" +
		":overmind-type: page\n" +
		":overmind-area: bebé\n" +
		":overmind-tags: silla, compras\n" +
		":overmind-created-at: 2026-09-22T18:30:00Z\n" +
		":overmind-updated-at: 2026-09-22T18:45:00Z\n\n"
	if got := string(content); got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderPageWithoutOptionalMetadata(t *testing.T) {
	documentRenderer := newRenderer(t)
	timestamp := time.Date(2026, time.September, 22, 18, 30, 0, 0, time.UTC)

	content, err := documentRenderer.Render(context.Background(), renderer.PageTemplate, renderer.Page{
		ID:        "dfe42d0b-f697-412e-8d07-a49e926393e0",
		Title:     "Inbox",
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := "= Inbox\n" +
		":overmind-id: dfe42d0b-f697-412e-8d07-a49e926393e0\n" +
		":overmind-type: page\n" +
		":overmind-area:\n" +
		":overmind-tags:\n" +
		":overmind-created-at: 2026-09-22T18:30:00Z\n" +
		":overmind-updated-at: 2026-09-22T18:30:00Z\n\n"
	if got := string(content); got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderReturnsCanceledContext(t *testing.T) {
	documentRenderer := newRenderer(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	content, err := documentRenderer.Render(ctx, renderer.PageTemplate, renderer.Page{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Render() error = %v, want context.Canceled", err)
	}
	if content != nil {
		t.Fatalf("Render() content = %q, want nil", content)
	}
}

func TestRenderWrapsTemplateErrors(t *testing.T) {
	documentRenderer := newRenderer(t)

	content, err := documentRenderer.Render(context.Background(), "unknown", nil)
	if err == nil {
		t.Fatal("Render() error = nil, want unknown template error")
	}
	if !strings.Contains(err.Error(), `render template "unknown"`) {
		t.Fatalf("Render() error = %q, want template name", err)
	}
	if content != nil {
		t.Fatalf("Render() content = %q, want nil", content)
	}
}

func TestRenderRejectsDataThatDoesNotMatchTemplate(t *testing.T) {
	documentRenderer := newRenderer(t)

	content, err := documentRenderer.Render(context.Background(), renderer.PageTemplate, struct{}{})
	if err == nil {
		t.Fatal("Render() error = nil, want incompatible data error")
	}
	if !strings.Contains(err.Error(), `render template "page"`) {
		t.Fatalf("Render() error = %q, want template name", err)
	}
	if content != nil {
		t.Fatalf("Render() content = %q, want nil", content)
	}
}

func newRenderer(t *testing.T) *renderer.Renderer {
	t.Helper()

	documentRenderer, err := renderer.New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return documentRenderer
}
