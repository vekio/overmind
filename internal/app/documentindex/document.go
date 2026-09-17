// Package documentindex maps parsed Overmind documents to index entries.
package documentindex

import (
	"maps"

	"git.casta.me/alberto/overmind/internal/app/documentparser"
	"git.casta.me/alberto/overmind/internal/ports"
)

// FromParsed creates a read-model entry from a parsed document.
func FromParsed(document documentparser.ParsedDocument) ports.IndexedDocument {
	return ports.IndexedDocument{
		ID:         document.ID,
		Kind:       document.Kind,
		Title:      document.Title.String(),
		Area:       document.Area.String(),
		Tags:       document.Tags.Strings(),
		CreatedAt:  document.CreatedAt,
		UpdatedAt:  document.UpdatedAt,
		Attributes: maps.Clone(document.Attributes),
	}
}
