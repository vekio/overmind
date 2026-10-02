package main

import (
	"bytes"
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc"
	"github.com/vekio/overmind/pkg/asciidoc/edit"
)

func TestEditorUsesTheHighLevelEditingAPI(t *testing.T) {
	source := []byte("= Note\r\n:note-type: youtube\r\n\r\nBody\r\n")
	editor, err := edit.New(source)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := editor.SetHeaderAttribute("summary-status", "pending"); err != nil {
		t.Fatalf("SetHeaderAttribute() error = %v", err)
	}
	updated := editor.Bytes()
	if !bytes.Contains(updated, []byte(":summary-status: pending\r\n")) {
		t.Fatalf("updated source = %q", updated)
	}
	if parsed := asciidoc.Parse(updated); parsed.HasErrors() {
		t.Fatalf("reparsed diagnostics = %+v", parsed.Diagnostics)
	}
}
