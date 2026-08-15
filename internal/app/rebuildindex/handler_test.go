package rebuildindex

import (
	"context"
	"testing"

	"git.casta.me/alberto/overmind/internal/infrastructure/localfs"
	"git.casta.me/alberto/overmind/internal/ports"
)

type indexWriterStub struct {
	documents []ports.IndexedDocument
}

func (index *indexWriterStub) Upsert(context.Context, ports.IndexedDocument) error { return nil }
func (index *indexWriterStub) ReplaceAll(_ context.Context, documents []ports.IndexedDocument) error {
	index.documents = documents
	return nil
}

func TestHandlerRebuildsManagedDocuments(t *testing.T) {
	blobs := localfs.New(t.TempDir())
	for _, blob := range []ports.Blob{
		{Path: "page.adoc", Content: []byte("= Page\n:overmind-id: page-id\n:overmind-type: page\n:overmind-area: knowledge\n:unrelated: ignored\n\nBody\n")},
		{Path: "unmanaged.adoc", Content: []byte("= External document\n\nBody\n")},
	} {
		if err := blobs.Create(context.Background(), blob); err != nil {
			t.Fatalf("Create(%q) error = %v", blob.Path, err)
		}
	}
	index := &indexWriterStub{}

	result, err := NewRebuildIndexHandler(blobs, index).Handle(context.Background(), RebuildIndexCommand{})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.Documents != 1 || len(index.documents) != 1 {
		t.Fatalf("result = %+v, indexed = %d", result, len(index.documents))
	}
	document := index.documents[0]
	if document.ID.String() != "page-id" || document.Attributes["overmind-area"] != "knowledge" {
		t.Fatalf("indexed document = %+v", document)
	}
	if _, exists := document.Attributes["unrelated"]; exists {
		t.Fatalf("unrelated attribute was indexed: %+v", document.Attributes)
	}
}

func TestHandlerIgnoresMalformedUnmanagedAsciiDoc(t *testing.T) {
	blobs := localfs.New(t.TempDir())
	if err := blobs.Create(context.Background(), ports.Blob{
		Path: "external.adoc", Content: []byte("= External\n:broken attribute\n"),
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	index := &indexWriterStub{}
	result, err := NewRebuildIndexHandler(blobs, index).Handle(context.Background(), RebuildIndexCommand{})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.Documents != 0 || len(index.documents) != 0 {
		t.Fatalf("result = %+v, indexed = %d", result, len(index.documents))
	}
}
