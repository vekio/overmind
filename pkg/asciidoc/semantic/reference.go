package semantic

import (
	"strings"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
)

// ReferenceStatus describes the result of resolving a cross-reference.
type ReferenceStatus uint8

const (
	ReferenceUnknown ReferenceStatus = iota
	ReferenceResolved
	ReferenceUnresolved
	ReferenceExternal
)

// String returns a stable lowercase reference status.
func (s ReferenceStatus) String() string {
	switch s {
	case ReferenceResolved:
		return "resolved"
	case ReferenceUnresolved:
		return "unresolved"
	case ReferenceExternal:
		return "external"
	default:
		return "unknown"
	}
}

// Reference is the semantic resolution of one CrossReference AST node.
// Anchor is populated only for a locally resolved reference.
type Reference struct {
	// Node is the original syntax node.
	Node *ast.CrossReference
	// Target is the target exactly as parsed.
	Target string
	// LocalID is populated for local references, without a leading '#'.
	LocalID string
	Status  ReferenceStatus
	// Anchor is non-nil only when Status is ReferenceResolved.
	Anchor       *Anchor
	Source       ast.Span
	TargetSource ast.Span
}

func analyzeReferences(document *ast.Document, anchors AnchorIndex) ([]Reference, []diagnostic.Diagnostic) {
	var references []Reference
	var diagnostics []diagnostic.Diagnostic
	ast.Walk(document, func(node ast.Node) bool {
		crossReference, ok := node.(*ast.CrossReference)
		if !ok {
			return true
		}
		reference := Reference{
			Node:         crossReference,
			Target:       crossReference.Target,
			Source:       crossReference.Source,
			TargetSource: crossReference.TargetSource,
		}
		localID, local := localReferenceID(crossReference.Target)
		if !local {
			reference.Status = ReferenceExternal
			references = append(references, reference)
			return true
		}
		reference.LocalID = localID
		if anchor, found := anchors.Lookup(localID); found {
			reference.Status = ReferenceResolved
			reference.Anchor = &anchor
		} else {
			reference.Status = ReferenceUnresolved
			diagnostics = append(diagnostics, diagnostic.Diagnostic{
				Severity: diagnostic.SeverityWarning,
				Message:  "unresolved internal reference \"" + crossReference.Target + "\"",
				Source:   crossReference.TargetSource,
			})
		}
		references = append(references, reference)
		return true
	})
	return references, diagnostics
}

func localReferenceID(target string) (string, bool) {
	if id, found := strings.CutPrefix(target, "#"); found {
		return id, id != ""
	}
	if strings.Contains(target, "#") || strings.ContainsAny(target, "/\\") || strings.HasSuffix(strings.ToLower(target), ".adoc") {
		return "", false
	}
	return target, target != ""
}
