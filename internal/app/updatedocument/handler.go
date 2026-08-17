package updatedocument

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"git.casta.me/alberto/overmind/internal/app/documentindex"
	"git.casta.me/alberto/overmind/internal/app/documentparser"
	"git.casta.me/alberto/overmind/internal/ports"
	asciidocedit "git.casta.me/alberto/overmind/pkg/asciidoc/edit"
)

const updatedAtAttribute = "overmind-updated-at"

// UpdateDocumentHandler validates, conditionally replaces, and reindexes an
// existing managed document.
type UpdateDocumentHandler struct {
	blobReader    ports.BlobReader
	blobUpdater   ports.BlobUpdater
	clock         ports.Clock
	documentIndex ports.DocumentIndexWriter
}

// NewUpdateDocumentHandler creates the command handler.
func NewUpdateDocumentHandler(
	blobReader ports.BlobReader,
	blobUpdater ports.BlobUpdater,
	clock ports.Clock,
	documentIndex ports.DocumentIndexWriter,
) *UpdateDocumentHandler {
	if blobReader == nil {
		panic("update document handler requires blob reader")
	}
	if blobUpdater == nil {
		panic("update document handler requires blob updater")
	}
	if clock == nil {
		panic("update document handler requires clock")
	}
	if documentIndex == nil {
		panic("update document handler requires document index writer")
	}
	return &UpdateDocumentHandler{
		blobReader:    blobReader,
		blobUpdater:   blobUpdater,
		clock:         clock,
		documentIndex: documentIndex,
	}
}

// Handle replaces the document only when it still has the revision read by
// the caller. Stable metadata cannot be changed through raw editing.
func (handler *UpdateDocumentHandler) Handle(
	ctx context.Context,
	command UpdateDocumentCommand,
) (UpdateDocumentResult, error) {
	if command.ExpectedRevision == "" {
		return UpdateDocumentResult{}, ErrRevisionRequired
	}

	currentBlob, err := handler.blobReader.Get(ctx, command.Path)
	if err != nil {
		return UpdateDocumentResult{}, fmt.Errorf("update document %q: read current content: %w", command.Path, err)
	}
	if currentBlob.Revision != command.ExpectedRevision {
		return UpdateDocumentResult{}, fmt.Errorf("update document %q: %w", command.Path, ports.ErrBlobChanged)
	}

	currentDocument, err := parseManagedDocument(currentBlob.Content)
	if err != nil {
		return UpdateDocumentResult{}, fmt.Errorf("update document %q: current content: %w", command.Path, err)
	}
	if bytes.Equal(currentBlob.Content, command.Content) {
		return UpdateDocumentResult{Path: currentBlob.Path, Revision: currentBlob.Revision}, nil
	}
	updatedSource, err := sourceWithUpdatedAt(command.Content, handler.clock.Now())
	if err != nil {
		return UpdateDocumentResult{}, fmt.Errorf("update document %q: set update time: %w", command.Path, err)
	}
	updatedBlob := ports.Blob{Path: currentBlob.Path, Content: updatedSource}
	updatedDocument, err := parseManagedDocument(updatedBlob.Content)
	if err != nil {
		return UpdateDocumentResult{}, fmt.Errorf("update document %q: edited content: %w", command.Path, err)
	}
	if currentDocument.ID != updatedDocument.ID ||
		currentDocument.Kind != updatedDocument.Kind ||
		!currentDocument.CreatedAt.Equal(updatedDocument.CreatedAt) {
		return UpdateDocumentResult{}, fmt.Errorf("update document %q: %w", command.Path, ErrImmutableMetadata)
	}

	newRevision, err := handler.blobUpdater.Update(ctx, updatedBlob, command.ExpectedRevision)
	if err != nil {
		return UpdateDocumentResult{}, fmt.Errorf("update document %q: store edited content: %w", command.Path, err)
	}
	indexEntry := documentindex.FromParsed(updatedBlob.Path, updatedDocument)
	if err := handler.documentIndex.Upsert(ctx, indexEntry); err != nil {
		return UpdateDocumentResult{}, handler.rollbackBlob(ctx, currentBlob, newRevision, err)
	}
	return UpdateDocumentResult{Path: updatedBlob.Path, Revision: newRevision}, nil
}

func sourceWithUpdatedAt(source []byte, updatedAt time.Time) ([]byte, error) {
	if updatedAt.IsZero() {
		return nil, fmt.Errorf("update time is required")
	}
	editor, err := asciidocedit.New(source)
	if err != nil {
		return nil, err
	}
	if err := editor.SetHeaderAttribute(
		updatedAtAttribute,
		updatedAt.UTC().Format(time.RFC3339),
	); err != nil {
		return nil, err
	}
	return editor.Bytes(), nil
}

func parseManagedDocument(source []byte) (documentparser.ParsedDocument, error) {
	document, managed, err := documentparser.Parse(source)
	if err != nil {
		return documentparser.ParsedDocument{}, err
	}
	if !managed {
		return documentparser.ParsedDocument{}, ErrUnmanagedDocument
	}
	return document, nil
}

func (handler *UpdateDocumentHandler) rollbackBlob(
	ctx context.Context,
	current ports.Blob,
	updatedRevision string,
	indexErr error,
) error {
	_, rollbackErr := handler.blobUpdater.Update(ctx, ports.Blob{
		Path:    current.Path,
		Content: current.Content,
	}, updatedRevision)
	if rollbackErr != nil {
		indexErr = errors.Join(indexErr, fmt.Errorf("restore previous content: %w", rollbackErr))
	}
	return fmt.Errorf("update document %q: index edited content: %w", current.Path, indexErr)
}
