package domain

import "time"

// Metadata contains the properties shared by every Overmind document.
type Metadata struct {
	id        DocumentID
	kind      DocumentKind
	title     Title
	tags      Tags
	createdAt time.Time
}

func newMetadata(id DocumentID, kind DocumentKind, title Title, tags Tags, createdAt time.Time) Metadata {
	if id.String() == "" {
		panic("document metadata requires id")
	}
	if !kind.IsValid() {
		panic("document metadata requires kind")
	}
	if title.IsZero() {
		panic("document metadata requires title")
	}
	if createdAt.IsZero() {
		panic("document metadata requires creation time")
	}
	return Metadata{id: id, kind: kind, title: title, tags: tags, createdAt: createdAt}
}

// ID returns the stable document identifier.
func (metadata Metadata) ID() DocumentID { return metadata.id }

// Kind returns the document kind.
func (metadata Metadata) Kind() DocumentKind { return metadata.kind }

// Title returns the document title.
func (metadata Metadata) Title() Title { return metadata.title }

// Tags returns the document tags.
func (metadata Metadata) Tags() Tags { return metadata.tags }

// CreatedAt returns when the document was created.
func (metadata Metadata) CreatedAt() time.Time { return metadata.createdAt }
