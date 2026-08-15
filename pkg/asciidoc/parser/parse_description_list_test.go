package parser_test

import (
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

func TestParseDescriptionListNestingAndContinuation(t *testing.T) {
	result := parse([]byte("CPU:: The *processor*\nRAM::: Nested memory\nDisk:: Storage\n+\nNOTE: Attached note\n"))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	list := requireBlock[*ast.DescriptionList](t, result.Document.Blocks[0])
	if list.Level != 1 || len(list.Items) != 2 {
		t.Fatalf("description list = %+v", list)
	}
	if list.Items[0].Term != "CPU" || inlineText(t, list.Items[0].DescriptionInlines) != "The processor" || len(list.Items[0].Blocks) != 1 {
		t.Fatalf("first item = %+v", list.Items[0])
	}
	nested := requireBlock[*ast.DescriptionList](t, list.Items[0].Blocks[0])
	if nested.Level != 2 || nested.Items[0].Term != "RAM" {
		t.Fatalf("nested description list = %+v", nested)
	}
	admonition := requireBlock[*ast.Admonition](t, list.Items[1].Blocks[0])
	if admonition.Kind != ast.AdmonitionNote {
		t.Fatalf("attached admonition = %+v", admonition)
	}
}
