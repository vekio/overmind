package edit_test

import (
	"fmt"

	"git.casta.me/alberto/overmind/pkg/asciidoc"
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/edit"
)

func ExampleReplaceSectionContent() {
	source := []byte("== Summary\n\nPending.\n\n== Notes\n\nKeep this.\n")
	processed := asciidoc.Process(source)

	var summary *ast.Section
	ast.Walk(processed.Document, func(node ast.Node) bool {
		section, ok := node.(*ast.Section)
		if ok && section.Title == "Summary" {
			summary = section
			return false
		}
		return true
	})

	change, err := edit.ReplaceSectionContent(source, summary, []byte("Generated summary."))
	if err != nil {
		panic(err)
	}
	updated, err := edit.Apply(source, []edit.TextEdit{change})
	if err != nil {
		panic(err)
	}
	if reparsed := asciidoc.Process(updated); reparsed.HasErrors() {
		panic(reparsed.Diagnostics)
	}
	fmt.Print(string(updated))

	// Output:
	// == Summary
	//
	// Generated summary.
	//
	// == Notes
	//
	// Keep this.
}
