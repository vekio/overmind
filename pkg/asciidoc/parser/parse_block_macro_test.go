package parser_test

import (
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

func TestParseBlockMacroWithMetadata(t *testing.T) {
	result := parse([]byte("[[diagram]]\n.Architecture\nimage::diagram.svg[Architecture diagram]"))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	macro := requireBlock[*ast.BlockMacro](t, result.Document.Blocks[0])
	if macro.Name != "image" || macro.Target != "diagram.svg" || macro.Attributes != "Architecture diagram" {
		t.Fatalf("macro = %+v", macro)
	}
	if macro.Metadata.Anchor == nil || macro.Metadata.Anchor.ID != "diagram" || macro.Metadata.Title == nil || macro.Metadata.Title.Text != "Architecture" {
		t.Fatalf("metadata = %+v", macro.Metadata)
	}
	if macro.TargetSource.Start.Offset >= macro.AttributesSource.Start.Offset {
		t.Fatalf("macro spans = target %s attributes %s", macro.TargetSource, macro.AttributesSource)
	}
}

func TestParseIncludeShapeWithoutReadingTarget(t *testing.T) {
	result := parse([]byte("include::missing-file.adoc[lines=1..3]"))
	macro := requireBlock[*ast.BlockMacro](t, result.Document.Blocks[0])
	if macro.Name != "include" || macro.Target != "missing-file.adoc" || result.HasErrors() {
		t.Fatalf("include-shaped macro = %+v diagnostics=%+v", macro, result.Diagnostics)
	}
}
