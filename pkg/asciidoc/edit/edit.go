// Package edit creates and applies source-preserving changes to AsciiDoc text.
// It does not perform filesystem I/O.
package edit

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

// TextEdit replaces the half-open Source range with Replacement. An empty
// range inserts text and an empty replacement deletes the range.
type TextEdit struct {
	Source      ast.Span
	Replacement []byte
}

// Apply validates and applies edits to source. Edits may be supplied in any
// order, but they must not overlap or share an insertion point. The returned
// bytes do not alias source or any edit replacement. Every AST and analysis
// derived from source becomes stale after an edit is applied.
func Apply(source []byte, edits []TextEdit) ([]byte, error) {
	ordered := append([]TextEdit(nil), edits...)
	sort.SliceStable(ordered, func(left, right int) bool {
		return ordered[left].Source.Start.Offset < ordered[right].Source.Start.Offset
	})

	for index, change := range ordered {
		start, end := change.Source.Start.Offset, change.Source.End.Offset
		if start < 0 || end < start || end > len(source) {
			return nil, fmt.Errorf("edit %d has invalid byte range [%d, %d) for %d-byte source", index, start, end, len(source))
		}
		if index == 0 {
			continue
		}
		previous := ordered[index-1].Source
		if start < previous.End.Offset || start == previous.Start.Offset {
			return nil, fmt.Errorf("edit ranges [%d, %d) and [%d, %d) overlap", previous.Start.Offset, previous.End.Offset, start, end)
		}
	}

	capacity := len(source)
	for _, change := range ordered {
		capacity += len(change.Replacement) - (change.Source.End.Offset - change.Source.Start.Offset)
	}
	var result bytes.Buffer
	result.Grow(capacity)
	cursor := 0
	for _, change := range ordered {
		result.Write(source[cursor:change.Source.Start.Offset])
		result.Write(change.Replacement)
		cursor = change.Source.End.Offset
	}
	result.Write(source[cursor:])
	return result.Bytes(), nil
}
