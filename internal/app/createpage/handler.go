// Package createpage implements the use case for creating an AsciiDoc page.
package createpage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"git.casta.me/alberto/overmind/internal/domain"
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
	logger        *slog.Logger
}

// NewCreatePageHandler creates the use-case handler.
func NewCreatePageHandler(
	blobWriter ports.BlobWriter,
	renderer ports.Renderer,
	idGenerator ports.IDGenerator,
	clock ports.Clock,
	documentIndex ports.DocumentIndexWriter,
	logger *slog.Logger,
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
	if logger == nil {
		panic("create page handler requires logger")
	}

	return &CreatePageHandler{
		blobWriter:    blobWriter,
		renderer:      renderer,
		idGenerator:   idGenerator,
		clock:         clock,
		documentIndex: documentIndex,
		logger:        logger,
	}
}

// Handle validates, renders, persists, and indexes a new page.
func (handler *CreatePageHandler) Handle(ctx context.Context, command CreatePageCommand) (CreatePageResult, error) {
	page, err := handler.newPageFromCommand(command)
	if err != nil {
		if isExpectedValidationError(err) {
			handler.logger.DebugContext(ctx, "page validation failed", "error", err)
		} else {
			handler.logger.DebugContext(ctx, "page initialization failed", "error", err)
		}
		return CreatePageResult{}, fmt.Errorf("create page: %w", err)
	}
	documentKey := pageDocumentKey(page)
	handler.logger.DebugContext(ctx, "creating page", "path", documentKey)

	renderedPage, err := handler.renderer.Render(ctx, page.Kind().String(), page)
	if err != nil {
		handler.logger.DebugContext(ctx, "page rendering failed", "path", documentKey, "error", err)
		return CreatePageResult{}, fmt.Errorf("render page %q: %w", documentKey, err)
	}
	if err := handler.blobWriter.Create(ctx, ports.Blob{Path: documentKey, Content: renderedPage}); err != nil {
		if errors.Is(err, ports.ErrBlobAlreadyExists) {
			handler.logger.DebugContext(ctx, "page already exists", "path", documentKey)
			return CreatePageResult{}, &PageAlreadyExistsError{Path: documentKey}
		}
		handler.logger.DebugContext(ctx, "page storage failed", "path", documentKey, "error", err)
		return CreatePageResult{}, fmt.Errorf("store page %q: %w", documentKey, err)
	}
	if err := handler.documentIndex.Upsert(ctx, indexEntryForPage(page, documentKey)); err != nil {
		rollbackErr := handler.rollbackStoredPage(ctx, documentKey, err)
		handler.logger.DebugContext(ctx, "page indexing failed", "path", documentKey, "error", rollbackErr)
		return CreatePageResult{}, rollbackErr
	}

	handler.logger.DebugContext(ctx, "page created", "id", page.ID(), "path", documentKey)
	return CreatePageResult{ID: page.ID(), Path: documentKey}, nil
}

func isExpectedValidationError(err error) bool {
	return errors.Is(err, domain.ErrInvalidTitle) ||
		errors.Is(err, domain.ErrInvalidArea) ||
		errors.Is(err, domain.ErrInvalidTag) ||
		errors.Is(err, domain.ErrDuplicateTag)
}

// rollbackStoredPage attempts to remove the blob after an index failure and
// preserves both errors when the compensation also fails.
func (handler *CreatePageHandler) rollbackStoredPage(ctx context.Context, documentKey string, indexErr error) error {
	if cleanupErr := handler.blobWriter.Delete(ctx, documentKey); cleanupErr != nil {
		return &PageCreationIncompleteError{
			Path:       documentKey,
			indexErr:   fmt.Errorf("index page: %w", indexErr),
			cleanupErr: fmt.Errorf("remove created page: %w", cleanupErr),
		}
	}
	return fmt.Errorf("index page %q: %w", documentKey, indexErr)
}
