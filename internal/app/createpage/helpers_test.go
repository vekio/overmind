package createpage

import (
	"context"
	"time"

	"git.casta.me/alberto/overmind/internal/ports"
)

var testCreatedAt = time.Date(2026, time.August, 14, 10, 30, 0, 0, time.UTC)

type rendererStub struct {
	name  string
	data  any
	calls int
	err   error
}

func (renderer *rendererStub) Render(_ context.Context, name string, data any) ([]byte, error) {
	renderer.name = name
	renderer.data = data
	renderer.calls++
	if renderer.err != nil {
		return nil, renderer.err
	}
	return []byte("rendered page"), nil
}

type idGeneratorStub struct {
	id    string
	calls int
	err   error
}

func (generator *idGeneratorStub) Generate() (string, error) {
	generator.calls++
	return generator.id, generator.err
}

type clockStub struct{ now time.Time }

func (clock clockStub) Now() time.Time { return clock.now }

type blobWriterStub struct {
	created     ports.Blob
	createCalls int
	createErr   error
	deletedPath string
	deleteCalls int
	deleteErr   error
}

func (writer *blobWriterStub) Create(_ context.Context, blob ports.Blob) error {
	writer.created = blob
	writer.createCalls++
	return writer.createErr
}

func (writer *blobWriterStub) Put(context.Context, ports.Blob) error { return nil }

func (writer *blobWriterStub) Delete(_ context.Context, path string) error {
	writer.deletedPath = path
	writer.deleteCalls++
	return writer.deleteErr
}

type indexWriterStub struct {
	document ports.IndexedDocument
	calls    int
	err      error
}

func (index *indexWriterStub) Upsert(_ context.Context, document ports.IndexedDocument) error {
	index.document = document
	index.calls++
	return index.err
}

func (index *indexWriterStub) ReplaceAll(context.Context, []ports.IndexedDocument) error {
	return nil
}
