// Package documentindex maps parsed Overmind documents to index entries.
package documentindex

import (
	"maps"

	"git.casta.me/alberto/overmind/internal/app/documentparser"
	"git.casta.me/alberto/overmind/internal/ports"
)

// FromParsed creates the read-model entry for a parsed document at path.
func FromParsed(path string, document documentparser.ParsedDocument) ports.IndexedDocument {
	return ports.IndexedDocument{
		ID:         document.ID,
		Path:       path,
		Kind:       document.Kind,
		Title:      document.Title.String(),
		Tags:       document.Tags.Strings(),
		CreatedAt:  document.CreatedAt,
		UpdatedAt:  document.UpdatedAt,
		Attributes: maps.Clone(document.Attributes),
	}
}
