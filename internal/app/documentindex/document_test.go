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
	tag, _ := domain.NewTag("Go")
	tags, _ := domain.NewTags(tag)
	createdAt := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	document := documentparser.ParsedDocument{
		ID: id, Kind: domain.DocumentKindPage, Title: title, Tags: tags, CreatedAt: createdAt, UpdatedAt: createdAt,
		Attributes: map[string]string{"area": "knowledge"},
	}

	indexed := FromParsed("page/knowledge/page.adoc", document)
	if indexed.ID != id || indexed.Path != "page/knowledge/page.adoc" || indexed.Kind != domain.DocumentKindPage ||
		indexed.Title != "Page" || len(indexed.Tags) != 1 || indexed.Tags[0] != "go" ||
		!indexed.CreatedAt.Equal(createdAt) || !indexed.UpdatedAt.Equal(createdAt) || indexed.Attributes["area"] != "knowledge" {
		t.Fatalf("FromParsed() = %+v", indexed)
	}
	document.Attributes["area"] = "changed"
	if indexed.Attributes["area"] != "knowledge" {
		t.Fatal("index entry attributes alias parsed document attributes")
	}
}
