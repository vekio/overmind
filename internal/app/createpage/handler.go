// Package createpage implements the use case for creating an AsciiDoc page.
package createpage

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"git.casta.me/alberto/overmind/internal/domain"
	"git.casta.me/alberto/overmind/internal/ports"
)

const pageTemplate = "page"

// CreatePageHandler creates AsciiDoc pages without replacing existing
// documents.
type CreatePageHandler struct {
	blobs    ports.BlobWriter
	renderer ports.Renderer
	ids      ports.IDGenerator
	clock    ports.Clock
	index    ports.DocumentIndexWriter
}

// NewCreatePageHandler creates the use-case handler.
func NewCreatePageHandler(
	blobs ports.BlobWriter,
	renderer ports.Renderer,
	ids ports.IDGenerator,
	clock ports.Clock,
	index ports.DocumentIndexWriter,
) *CreatePageHandler {
	if blobs == nil {
		panic("create page handler requires blob writer")
	}
	if renderer == nil {
		panic("create page handler requires template renderer")
	}
	if ids == nil {
		panic("create page handler requires id generator")
	}
	if clock == nil {
		panic("create page handler requires clock")
	}
	if index == nil {
		panic("create page handler requires document index writer")
	}

	return &CreatePageHandler{blobs: blobs, renderer: renderer, ids: ids, clock: clock, index: index}
}

// Handle validates, renders, and persists a page.
func (handler *CreatePageHandler) Handle(ctx context.Context, command CreatePageCommand) (CreatePageResult, error) {
	id, err := handler.ids.Generate()
	if err != nil {
		return CreatePageResult{}, fmt.Errorf("create page id: %w", err)
	}
	page, err := newPage(id, handler.clock.Now(), command)
	if err != nil {
		return CreatePageResult{}, fmt.Errorf("create page: %w", err)
	}

	content, err := handler.renderer.Render(ctx, pageTemplate, page)
	documentKey := pageDocumentKey(page)
	if err != nil {
		return CreatePageResult{}, fmt.Errorf("create page %q: %w", documentKey, err)
	}
	if err := handler.blobs.Create(ctx, ports.Blob{Path: documentKey, Content: content}); err != nil {
		return CreatePageResult{}, fmt.Errorf("create page %q: %w", documentKey, err)
	}
	if err := handler.index.Upsert(ctx, indexedPage(page, documentKey, content)); err != nil {
		cleanupErr := handler.blobs.Delete(ctx, documentKey)
		return CreatePageResult{}, fmt.Errorf("create page %q index: %w", documentKey, errors.Join(err, cleanupErr))
	}

	return CreatePageResult{ID: page.ID()}, nil
}

func indexedPage(page domain.Page, documentPath string, content []byte) ports.IndexedDocument {
	return ports.IndexedDocument{
		ID:      page.ID(),
		Path:    documentPath,
		Content: append([]byte(nil), content...),
		Attributes: map[string]string{
			domain.AttributeID:        page.ID().String(),
			domain.AttributeType:      "page",
			domain.AttributeTitle:     page.Title().String(),
			domain.AttributeArea:      page.Area().String(),
			domain.AttributeCreatedAt: page.CreatedAt().UTC().Format(time.RFC3339),
		},
	}
}

// pageDocumentKey returns the store-relative logical key. Storage adapters are
// responsible for translating it to their concrete path or object key.
func pageDocumentKey(page domain.Page) string {
	return path.Join(page.Area().String(), page.Title().Slug()+".adoc")
}

func newPage(id string, createdAt time.Time, command CreatePageCommand) (domain.Page, error) {
	documentID, err := domain.NewDocumentID(id)
	if err != nil {
		return domain.Page{}, err
	}

	title, err := domain.NewTitle(command.Title)
	if err != nil {
		return domain.Page{}, err
	}

	area := domain.Area{}
	if strings.TrimSpace(command.Area) != "" {
		area, err = domain.NewArea(command.Area)
		if err != nil {
			return domain.Page{}, err
		}
	}

	return domain.NewPage(documentID, title, area, createdAt), nil
}
