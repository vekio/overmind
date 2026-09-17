package listdocuments

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

type indexListerStub struct {
	filter    ports.ListIndexedDocumentsFilter
	documents []ports.IndexedDocumentSummary
	err       error
	calls     int
}

func (index *indexListerStub) List(
	_ context.Context,
	filter ports.ListIndexedDocumentsFilter,
) ([]ports.IndexedDocumentSummary, error) {
	index.calls++
	index.filter = filter
	return index.documents, index.err
}

func TestHandlerNormalizesFiltersAndReturnsSummaries(t *testing.T) {
	id, _ := domain.NewDocumentID("page-id")
	index := &indexListerStub{documents: []ports.IndexedDocumentSummary{{
		ID: id, Kind: domain.DocumentKindPage, Title: "Page", Area: "knowledge/go", Tags: []string{"go", "ddd"},
	}}}
	handler := NewListDocumentsHandler(index)

	result, err := handler.Handle(context.Background(), ListDocumentsQuery{
		Type: " page ", Title: " Page ", Area: " Knowledge ",
		Tags: []string{"Go", "Diseño de dominio"},
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if index.filter.Kind != domain.DocumentKindPage || index.filter.Title != "Page" || index.filter.Area.String() != "knowledge" ||
		!reflect.DeepEqual(index.filter.Tags.Strings(), []string{"go", "diseno-de-dominio"}) {
		t.Fatalf("filter = %+v", index.filter)
	}
	if len(result.Documents) != 1 || result.Documents[0].ID != id ||
		result.Documents[0].Title != "Page" || result.Documents[0].Area != "knowledge/go" || result.Documents[0].Type != domain.DocumentKindPage ||
		!reflect.DeepEqual(result.Documents[0].Tags, []string{"go", "ddd"}) {
		t.Fatalf("result = %+v", result)
	}

	index.documents[0].Tags[0] = "changed"
	if result.Documents[0].Tags[0] != "go" {
		t.Fatal("result tags alias the index representation")
	}
}

func TestHandlerRejectsInvalidFiltersBeforeListing(t *testing.T) {
	for name, query := range map[string]ListDocumentsQuery{
		"type":          {Type: "journal"},
		"area":          {Area: "///"},
		"tag":           {Tags: []string{"---"}},
		"duplicate tag": {Tags: []string{"Go", "go"}},
	} {
		t.Run(name, func(t *testing.T) {
			index := &indexListerStub{}
			_, err := NewListDocumentsHandler(index).Handle(context.Background(), query)
			if err == nil {
				t.Fatal("Handle() error = nil")
			}
			if index.calls != 0 {
				t.Fatalf("index calls = %d", index.calls)
			}
		})
	}
}

func TestHandlerPreservesIndexFailure(t *testing.T) {
	indexErr := errors.New("index unavailable")
	_, err := NewListDocumentsHandler(&indexListerStub{err: indexErr}).Handle(
		context.Background(),
		ListDocumentsQuery{},
	)
	if !errors.Is(err, indexErr) {
		t.Fatalf("Handle() error = %v", err)
	}
}

func TestNewListDocumentsHandlerRequiresIndex(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewListDocumentsHandler() did not panic")
		}
	}()
	NewListDocumentsHandler(nil)
}
