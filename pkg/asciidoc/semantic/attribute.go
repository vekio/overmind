package semantic

import (
	"sort"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

// Attribute is one header attribute declaration. Operation distinguishes an
// assignment from an unset declaration.
type Attribute struct {
	Name        string
	Value       string
	Operation   ast.AttributeOperation
	Source      ast.Span
	NameSource  ast.Span
	ValueSource ast.Span
}

// AttributeSet contains the effective header attributes and their complete
// declaration history. Attribute names are matched exactly.
type AttributeSet struct {
	effective map[string]Attribute
	history   map[string][]Attribute
}

// Lookup returns an attribute's effective value. An explicitly set empty
// attribute returns an empty value and true.
func (s AttributeSet) Lookup(name string) (string, bool) {
	attribute, ok := s.effective[name]
	return attribute.Value, ok
}

// Get returns the declaration that produced an attribute's effective value.
func (s AttributeSet) Get(name string) (Attribute, bool) {
	attribute, ok := s.effective[name]
	return attribute, ok
}

// Has reports whether an attribute currently has an effective value.
func (s AttributeSet) Has(name string) bool {
	_, ok := s.effective[name]
	return ok
}

// History returns every header declaration for name in source order,
// including declarations that unset it. The returned slice is independent.
func (s AttributeSet) History(name string) []Attribute {
	return append([]Attribute(nil), s.history[name]...)
}

// All returns the effective attributes ordered by the source position of the
// declarations that produced their current values.
func (s AttributeSet) All() []Attribute {
	result := make([]Attribute, 0, len(s.effective))
	for _, attribute := range s.effective {
		result = append(result, attribute)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Source.Start.Offset == result[j].Source.Start.Offset {
			return result[i].Name < result[j].Name
		}
		return result[i].Source.Start.Offset < result[j].Source.Start.Offset
	})
	return result
}

func analyzeHeaderAttributes(document *ast.Document) AttributeSet {
	result := AttributeSet{
		effective: make(map[string]Attribute),
		history:   make(map[string][]Attribute),
	}
	if document == nil {
		return result
	}
	for _, block := range document.Blocks {
		entry, ok := block.(*ast.AttributeEntry)
		if !ok || entry == nil || !entry.Header {
			continue
		}
		attribute := Attribute{
			Name:        entry.Name,
			Value:       entry.Value,
			Operation:   entry.Operation,
			Source:      entry.Source,
			NameSource:  entry.NameSource,
			ValueSource: entry.ValueSource,
		}
		result.history[attribute.Name] = append(result.history[attribute.Name], attribute)
		switch attribute.Operation {
		case ast.AttributeSet:
			result.effective[attribute.Name] = attribute
		case ast.AttributeUnset:
			delete(result.effective, attribute.Name)
		}
	}
	return result
}
