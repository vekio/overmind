package domain

import (
	"testing"
	"time"
)

var pageCreatedAt = time.Date(2026, time.August, 14, 10, 30, 0, 0, time.UTC)

func TestNewPageComposesValidatedValues(t *testing.T) {
	id, _ := NewDocumentID("page-id")
	title, _ := NewTitle("Mi página")
	area, _ := NewArea("Knowledge/Go")

	page := NewPage(id, title, area, pageCreatedAt)

	if page.ID().String() != "page-id" {
		t.Fatalf("Page.ID() = %q, want %q", page.ID(), "page-id")
	}
	if page.Kind() != DocumentKindPage {
		t.Fatalf("Page.Kind() = %q, want %q", page.Kind(), "page")
	}
	if page.Title().String() != "Mi página" || page.Title().Slug() != "mi-pagina" {
		t.Fatalf("Page.Title() = value %q, slug %q", page.Title(), page.Title().Slug())
	}
	if page.Area().String() != "knowledge/go" {
		t.Fatalf("Page.Area() = %q, want %q", page.Area(), "knowledge/go")
	}
	if !page.CreatedAt().Equal(pageCreatedAt) {
		t.Fatalf("Page.CreatedAt() = %v, want %v", page.CreatedAt(), pageCreatedAt)
	}
}

func TestNewPageAllowsRootArea(t *testing.T) {
	id, _ := NewDocumentID("page-id")
	title, _ := NewTitle("Page")

	page := NewPage(id, title, Area{}, pageCreatedAt)

	if !page.Area().IsZero() {
		t.Fatalf("Page.Area() = %q, want root area", page.Area())
	}
}

func TestNewPageRequiresValidIDAndTitle(t *testing.T) {
	id, _ := NewDocumentID("page-id")
	title, _ := NewTitle("Page")

	for name, build := range map[string]func(){
		"document id":   func() { NewPage(DocumentID{}, title, Area{}, pageCreatedAt) },
		"title":         func() { NewPage(id, Title{}, Area{}, pageCreatedAt) },
		"creation time": func() { NewPage(id, title, Area{}, time.Time{}) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("NewPage() did not panic")
				}
			}()
			build()
		})
	}
}
