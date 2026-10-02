package parser_test

import (
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

func TestParseDocumentAttributeEntries(t *testing.T) {
	result := parse([]byte(":toc:\n= Document\n:sectnums:\n\nbody\n:imagesdir: assets\n:!old:"))
	if len(result.Diagnostics) != 0 || result.Document.Title == nil {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Document.Blocks) != 5 {
		t.Fatalf("blocks = %+v", result.Document.Blocks)
	}
	toc := requireBlock[*ast.AttributeEntry](t, result.Document.Blocks[0])
	images := requireBlock[*ast.AttributeEntry](t, result.Document.Blocks[3])
	unset := requireBlock[*ast.AttributeEntry](t, result.Document.Blocks[4])
	if toc.Operation != ast.AttributeSet || toc.Name != "toc" || toc.Value != "" || !toc.Header {
		t.Fatalf("toc = %+v", toc)
	}
	if images.Name != "imagesdir" || images.Value != "assets" || images.ValueSource == (ast.Span{}) || images.Header {
		t.Fatalf("imagesdir = %+v", images)
	}
	if unset.Operation != ast.AttributeUnset || unset.Name != "old" || unset.ValueSource != (ast.Span{}) {
		t.Fatalf("unset = %+v", unset)
	}
}
