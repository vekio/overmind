package domain

import (
	"testing"
	"time"
)

func TestMetadataExposesCommonDocumentValues(t *testing.T) {
	id, _ := NewDocumentID("document-id")
	title, _ := NewTitle("Document")
	createdAt := time.Date(2026, time.August, 16, 12, 0, 0, 0, time.UTC)

	metadata := newMetadata(id, DocumentKindPage, title, createdAt)

	if metadata.ID() != id || metadata.Kind() != DocumentKindPage || metadata.Title() != title || !metadata.CreatedAt().Equal(createdAt) {
		t.Fatalf("metadata = %#v", metadata)
	}
}
