package semantic

import (
	"sort"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
)

// Anchor associates an explicit AsciiDoc anchor with the AST node to which
// its block metadata is attached.
type Anchor struct {
	ID      string
	RefText string
	// Source covers the complete anchor declaration.
	Source        ast.Span
	IDSource      ast.Span
	RefTextSource ast.Span
	// Node is the block or section to which the anchor is attached.
	Node ast.Node
}

// AnchorIndex contains explicit anchors indexed by their exact ID. When an ID
// is duplicated, Lookup returns the first declaration and History returns all
// declarations in source order.
type AnchorIndex struct {
	first   map[string]Anchor
	history map[string][]Anchor
}

// Lookup returns the first anchor declared with id.
func (i AnchorIndex) Lookup(id string) (Anchor, bool) {
	anchor, ok := i.first[id]
	return anchor, ok
}

// Has reports whether id has an indexed anchor.
func (i AnchorIndex) Has(id string) bool {
	_, ok := i.first[id]
	return ok
}

// History returns all declarations for id in source order. The returned slice
// is independent from the index.
func (i AnchorIndex) History(id string) []Anchor {
	return append([]Anchor(nil), i.history[id]...)
}

// All returns the first declaration for each anchor ID in source order.
func (i AnchorIndex) All() []Anchor {
	result := make([]Anchor, 0, len(i.first))
	for _, anchor := range i.first {
		result = append(result, anchor)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].Source.Start.Offset == result[right].Source.Start.Offset {
			return result[left].ID < result[right].ID
		}
		return result[left].Source.Start.Offset < result[right].Source.Start.Offset
	})
	return result
}

func analyzeAnchors(document *ast.Document) (AnchorIndex, []diagnostic.Diagnostic) {
	index := AnchorIndex{
		first:   make(map[string]Anchor),
		history: make(map[string][]Anchor),
	}
	var diagnostics []diagnostic.Diagnostic
	ast.Walk(document, func(node ast.Node) bool {
		metadata, ok := nodeMetadata(node)
		if !ok || metadata.Anchor == nil {
			return true
		}
		value := metadata.Anchor
		anchor := Anchor{
			ID:            value.ID,
			RefText:       value.RefText,
			Source:        value.Source,
			IDSource:      value.IDSource,
			RefTextSource: value.RefTextSource,
			Node:          node,
		}
		index.history[anchor.ID] = append(index.history[anchor.ID], anchor)
		if first, duplicate := index.first[anchor.ID]; duplicate {
			diagnostics = append(diagnostics, diagnostic.Diagnostic{
				Severity: diagnostic.SeverityError,
				Message:  "duplicate anchor \"" + anchor.ID + "\"; first declared at " + first.IDSource.Start.String(),
				Source:   anchor.IDSource,
			})
			return true
		}
		index.first[anchor.ID] = anchor
		return true
	})
	return index, diagnostics
}

func nodeMetadata(node ast.Node) (ast.BlockMetadata, bool) {
	switch node := node.(type) {
	case *ast.Section:
		return node.Metadata, true
	case *ast.DelimitedBlock:
		return node.Metadata, true
	case *ast.Table:
		return node.Metadata, true
	case *ast.List:
		return node.Metadata, true
	case *ast.DescriptionList:
		return node.Metadata, true
	case *ast.Admonition:
		return node.Metadata, true
	case *ast.BlockMacro:
		return node.Metadata, true
	default:
		return ast.BlockMetadata{}, false
	}
}
