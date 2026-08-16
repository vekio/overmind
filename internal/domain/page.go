// Package domain contains Overmind's domain model.
package domain

import "time"

// Page is an AsciiDoc page stored in the knowledge base.
type Page struct {
	metadata Metadata
	area     Area
}

// NewPage creates a Page from validated value objects.
func NewPage(id DocumentID, title Title, tags Tags, area Area, createdAt time.Time) Page {
	return Page{metadata: newMetadata(id, DocumentKindPage, title, tags, createdAt), area: area}
}

// ID returns the page's stable document identifier.
func (page Page) ID() DocumentID { return page.metadata.ID() }

// Kind returns the document kind.
func (page Page) Kind() DocumentKind { return page.metadata.Kind() }

// Title returns the page title.
func (page Page) Title() Title { return page.metadata.Title() }

// Tags returns the page tags.
func (page Page) Tags() Tags { return page.metadata.Tags() }

// Area returns the page area.
func (page Page) Area() Area { return page.area }

// CreatedAt returns when the page was created.
func (page Page) CreatedAt() time.Time { return page.metadata.CreatedAt() }
