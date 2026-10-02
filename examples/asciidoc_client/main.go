// Command asciidoc_client demonstrates how an application can consume the
// public parser, semantic analysis, and AST traversal APIs.
package main

import (
	"fmt"
	"os"

	"github.com/vekio/overmind/pkg/asciidoc"
	"github.com/vekio/overmind/pkg/asciidoc/ast"
	"github.com/vekio/overmind/pkg/asciidoc/edit"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s document.adoc\n", os.Args[0])
		os.Exit(2)
	}
	if err := inspect(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func inspect(path string) error {
	source, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	result := asciidoc.Process(source)
	for _, item := range result.Diagnostics {
		fmt.Println("diagnostic:", item)
	}
	if result.HasErrors() {
		return fmt.Errorf("cannot inspect a document with errors")
	}
	analysis := result.Analysis

	printHeader(analysis)
	printAnchorsAndReferences(analysis)
	if _, err := summarySection(analysis); err != nil {
		return err
	}
	if _, ok := analysis.Header.Attributes.Lookup("youtube-url"); !ok {
		return fmt.Errorf("youtube note requires the youtube-url header attribute")
	}
	editor, err := edit.New(source)
	if err != nil {
		return err
	}
	if err := editor.SetHeaderAttribute("summary-status", "pending"); err != nil {
		return err
	}
	updated := editor.Bytes()
	if err := validateUpdatedSource(updated); err != nil {
		return err
	}

	fmt.Println("\n--- updated AsciiDoc (not written to disk) ---")
	_, err = os.Stdout.Write(updated)
	if err == nil && (len(updated) == 0 || updated[len(updated)-1] != '\n') {
		_, err = fmt.Fprintln(os.Stdout)
	}
	return err
}

func printHeader(analysis asciidoc.Analysis) {
	if analysis.Header.Title != nil {
		fmt.Println("document:", analysis.Header.Title.Text)
	}
	noteType, isNote := analysis.Header.Attributes.Lookup("note-type")
	videoURL, hasVideoURL := analysis.Header.Attributes.Lookup("youtube-url")
	fmt.Printf("note type: %q (defined=%t)\n", noteType, isNote)
	fmt.Printf("youtube URL: %q (defined=%t)\n", videoURL, hasVideoURL)
	if isNote && noteType == "youtube" && hasVideoURL {
		fmt.Println("action: request an AI summary for the YouTube video")
	}
}

func printAnchorsAndReferences(analysis asciidoc.Analysis) {
	fmt.Println("anchors:")
	for _, anchor := range analysis.Anchors.All() {
		fmt.Printf("  %s -> %T at %s\n", anchor.ID, anchor.Node, anchor.Source)
	}
	fmt.Println("references:")
	for _, reference := range analysis.References {
		fmt.Printf("  %s -> %s", reference.Target, reference.Status)
		if reference.Anchor != nil {
			fmt.Printf(" (%T)", reference.Anchor.Node)
		}
		fmt.Println()
	}
}

func summarySection(analysis asciidoc.Analysis) (*ast.Section, error) {
	title, ok := analysis.Header.Attributes.Lookup("summary-section")
	if !ok {
		return nil, fmt.Errorf("youtube note requires the summary-section header attribute")
	}
	var destination *ast.Section
	ast.Walk(analysis.Document, func(node ast.Node) bool {
		if destination != nil {
			return false
		}
		section, isSection := node.(*ast.Section)
		if isSection && section.Title == title {
			destination = section
			return false
		}
		return true
	})
	if destination == nil {
		return nil, fmt.Errorf("summary section %q was not found", title)
	}
	fmt.Printf("summary destination: %q at %s with %d block(s)\n", destination.Title, destination.Source, len(destination.Blocks))
	return destination, nil
}

func validateUpdatedSource(source []byte) error {
	result := asciidoc.Process(source)
	if result.HasErrors() {
		return fmt.Errorf("generated edit produced errors: %v", result.Diagnostics)
	}
	return nil
}
