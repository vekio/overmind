// Package domain contains Overmind's domain model.
package domain

import "time"

// Page is an AsciiDoc page stored in the knowledge base.
type Page struct {
	id        DocumentID
	title     Title
	area      Area
	createdAt time.Time
}

// NewPage creates a Page from validated value objects.
func NewPage(id DocumentID, title Title, area Area, createdAt time.Time) Page {
	if id.String() == "" {
		panic("page requires document id")
	}
	if title.IsZero() {
		panic("page requires title")
	}
	if createdAt.IsZero() {
		panic("page requires creation time")
	}
	return Page{id: id, title: title, area: area, createdAt: createdAt}
}

// ID returns the page's stable document identifier.
func (page Page) ID() DocumentID { return page.id }

// Title returns the page title.
func (page Page) Title() Title { return page.title }

// Area returns the page area.
func (page Page) Area() Area { return page.area }

// CreatedAt returns when the page was created.
func (page Page) CreatedAt() time.Time { return page.createdAt }
