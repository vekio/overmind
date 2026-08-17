package domain

import (
	"testing"
	"time"
)

func TestMetadataExposesCommonDocumentValues(t *testing.T) {
	id, _ := NewDocumentID("document-id")
	title, _ := NewTitle("Document")
	goTag, _ := NewTag("Go")
	dddTag, _ := NewTag("DDD")
	tags, _ := NewTags(goTag, dddTag)
	createdAt := time.Date(2026, time.August, 16, 12, 0, 0, 0, time.UTC)

	metadata := newMetadata(id, DocumentKindPage, title, tags, createdAt)

	if metadata.ID() != id || metadata.Kind() != DocumentKindPage || metadata.Title() != title || metadata.Tags().Len() != 2 ||
		!metadata.CreatedAt().Equal(createdAt) || !metadata.UpdatedAt().Equal(createdAt) {
		t.Fatalf("metadata = %#v", metadata)
	}
}
