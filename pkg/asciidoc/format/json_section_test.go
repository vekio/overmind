package format_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc"
	"github.com/vekio/overmind/pkg/asciidoc/format"
)

func TestWriteJSONSerializesSectionMetadata(t *testing.T) {
	result := asciidoc.Parse([]byte("[[summary,Summary reference]]\n== Summary\n\nPending."))
	var output bytes.Buffer
	if err := format.WriteJSON(&output, result); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	document := decoded["document"].(map[string]any)
	section := document["blocks"].([]any)[0].(map[string]any)
	headingSource := section["headingSource"].(map[string]any)
	if headingSource["start"].(map[string]any)["line"] != float64(2) {
		t.Fatalf("heading source = %#v", headingSource)
	}
	contentSource := section["contentSource"].(map[string]any)
	if contentSource["start"].(map[string]any)["line"] != float64(3) {
		t.Fatalf("content source = %#v", contentSource)
	}
	metadata := section["metadata"].(map[string]any)
	anchor := metadata["anchor"].(map[string]any)
	if anchor["id"] != "summary" || anchor["refText"] != "Summary reference" {
		t.Fatalf("section anchor = %#v", anchor)
	}
}
