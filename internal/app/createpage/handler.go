// Package createpage implements the use case for creating an AsciiDoc page.
package createpage

import (
	"context"
	"errors"
	"fmt"

	"git.casta.me/alberto/overmind/internal/ports"
)

// CreatePageHandler creates AsciiDoc pages without replacing existing
// documents.
type CreatePageHandler struct {
	blobWriter    ports.BlobWriter
	renderer      ports.Renderer
	idGenerator   ports.IDGenerator
	clock         ports.Clock
	documentIndex ports.DocumentIndexWriter
}

// NewCreatePageHandler creates the use-case handler.
func NewCreatePageHandler(
	blobWriter ports.BlobWriter,
	renderer ports.Renderer,
	idGenerator ports.IDGenerator,
	clock ports.Clock,
	documentIndex ports.DocumentIndexWriter,
) *CreatePageHandler {
	if blobWriter == nil {
		panic("create page handler requires blob writer")
	}
	if renderer == nil {
		panic("create page handler requires template renderer")
	}
	if idGenerator == nil {
		panic("create page handler requires id generator")
	}
	if clock == nil {
		panic("create page handler requires clock")
	}
	if documentIndex == nil {
		panic("create page handler requires document index writer")
	}

	return &CreatePageHandler{
		blobWriter:    blobWriter,
		renderer:      renderer,
		idGenerator:   idGenerator,
		clock:         clock,
		documentIndex: documentIndex,
	}
}

// Handle validates, renders, persists, and indexes a new page.
func (handler *CreatePageHandler) Handle(ctx context.Context, command CreatePageCommand) (CreatePageResult, error) {
	page, err := handler.newPageFromCommand(command)
	if err != nil {
		return CreatePageResult{}, fmt.Errorf("create page: %w", err)
	}
	documentKey := pageDocumentKey(page)

	renderedPage, err := handler.renderer.Render(ctx, page.Kind().String(), page)
	if err != nil {
		return CreatePageResult{}, fmt.Errorf("render page %q: %w", documentKey, err)
	}
	if err := handler.blobWriter.Create(ctx, ports.Blob{Path: documentKey, Content: renderedPage}); err != nil {
		return CreatePageResult{}, fmt.Errorf("store page %q: %w", documentKey, err)
	}
	if err := handler.documentIndex.Upsert(ctx, indexEntryForPage(page, documentKey)); err != nil {
		return CreatePageResult{}, handler.rollbackStoredPage(ctx, documentKey, err)
	}

	return CreatePageResult{ID: page.ID(), Path: documentKey}, nil
}

// rollbackStoredPage attempts to remove the blob after an index failure and
// preserves both errors when the compensation also fails.
func (handler *CreatePageHandler) rollbackStoredPage(ctx context.Context, documentKey string, indexErr error) error {
	if cleanupErr := handler.blobWriter.Delete(ctx, documentKey); cleanupErr != nil {
		return fmt.Errorf("index page %q: %w", documentKey, errors.Join(
			indexErr,
			fmt.Errorf("remove created page: %w", cleanupErr),
		))
	}
	return fmt.Errorf("index page %q: %w", documentKey, indexErr)
}
