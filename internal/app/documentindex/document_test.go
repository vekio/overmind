package documentindex

import (
	"testing"
	"time"

	"git.casta.me/alberto/overmind/internal/app/documentparser"
	"git.casta.me/alberto/overmind/internal/domain"
)

func TestFromParsedCreatesIndependentIndexEntry(t *testing.T) {
	id, _ := domain.NewDocumentID("page-id")
	title, _ := domain.NewTitle("Page")
	area, _ := domain.NewArea("Knowledge")
	tag, _ := domain.NewTag("Go")
	tags, _ := domain.NewTags(tag)
	createdAt := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	document := documentparser.ParsedDocument{
		ID: id, Kind: domain.DocumentKindPage, Title: title, Area: area, Tags: tags, CreatedAt: createdAt, UpdatedAt: createdAt,
		Attributes: map[string]string{"custom": "value"},
	}

	indexed := FromParsed(document)
	if indexed.ID != id || indexed.Kind != domain.DocumentKindPage || indexed.Area != "knowledge" ||
		indexed.Title != "Page" || len(indexed.Tags) != 1 || indexed.Tags[0] != "go" ||
		!indexed.CreatedAt.Equal(createdAt) || !indexed.UpdatedAt.Equal(createdAt) || indexed.Attributes["custom"] != "value" {
		t.Fatalf("FromParsed() = %+v", indexed)
	}
	document.Attributes["custom"] = "changed"
	if indexed.Attributes["custom"] != "value" {
		t.Fatal("index entry attributes alias parsed document attributes")
	}
}
