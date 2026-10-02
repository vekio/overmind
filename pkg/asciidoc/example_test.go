package asciidoc_test

import (
	"fmt"

	"github.com/vekio/overmind/pkg/asciidoc"
	"github.com/vekio/overmind/pkg/asciidoc/ast"
)

func ExampleProcess() {
	source := []byte("= Video note\n:note-type: youtube\n\n[[summary]]\n== Summary\n\nPending.")
	result := asciidoc.Process(source)
	if result.HasErrors() {
		fmt.Println(result.Diagnostics)
		return
	}

	noteType, defined := result.Analysis.Header.Attributes.Lookup("note-type")
	anchor, found := result.Analysis.Anchors.Lookup("summary")
	fmt.Println(result.Document.Title.Text)
	fmt.Println(noteType, defined)
	fmt.Println(anchor.ID, found)

	// Output:
	// Video note
	// youtube true
	// summary true
}

func ExampleProcessResult_walk() {
	result := asciidoc.Process([]byte("== Notes\n\nText.\n\n== Summary\n\nPending."))
	var summary *ast.Section
	ast.Walk(result.Document, func(node ast.Node) bool {
		section, ok := node.(*ast.Section)
		if ok && section.Title == "Summary" {
			summary = section
			return false
		}
		return true
	})
	fmt.Println(summary.Title)

	// Output:
	// Summary
}
