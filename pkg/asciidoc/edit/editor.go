package edit

import (
	"fmt"

	"github.com/vekio/overmind/pkg/asciidoc"
)

// Editor applies source-preserving semantic edits to an AsciiDoc document.
// Every successful operation reparses the updated source before committing it,
// so subsequent operations always use current syntax and source positions.
type Editor struct {
	source    []byte
	processed asciidoc.ProcessResult
}

// New creates an editor for a valid AsciiDoc source.
func New(source []byte) (*Editor, error) {
	ownedSource := append([]byte(nil), source...)
	processed := asciidoc.Process(ownedSource)
	if processed.HasErrors() {
		return nil, fmt.Errorf("create AsciiDoc editor: document has errors")
	}
	return &Editor{source: ownedSource, processed: processed}, nil
}

// SetHeaderAttribute sets a document-header attribute. An existing effective
// or unset declaration is replaced in place; otherwise it is appended to the
// current document header.
func (editor *Editor) SetHeaderAttribute(name, value string) error {
	if editor == nil {
		return fmt.Errorf("set header attribute: nil editor")
	}
	change, err := setHeaderAttributeEdit(editor.source, editor.processed, name, value)
	if err != nil {
		return err
	}
	updatedSource, err := Apply(editor.source, []TextEdit{change})
	if err != nil {
		return fmt.Errorf("set header attribute %q: %w", name, err)
	}
	processed := asciidoc.Process(updatedSource)
	if processed.HasErrors() {
		return fmt.Errorf("set header attribute %q: edited document has errors", name)
	}

	editor.source = updatedSource
	editor.processed = processed
	return nil
}

// Bytes returns an independent copy of the current edited source.
func (editor *Editor) Bytes() []byte {
	if editor == nil {
		return nil
	}
	return append([]byte(nil), editor.source...)
}
