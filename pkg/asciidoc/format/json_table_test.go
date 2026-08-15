package format_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc"
	"git.casta.me/alberto/overmind/pkg/asciidoc/format"
)

func TestWriteJSONSerializesTableRowsCellsAndInlines(t *testing.T) {
	result := asciidoc.Parse([]byte(".Inventory\n[cols=\"1,2\",options=\"header\"]\n|===\n|Name |Description\n\n|Parser |Builds the *AST*\n|==="))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	var output bytes.Buffer
	if err := format.WriteJSON(&output, result); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	document := decoded["document"].(map[string]any)
	table := document["blocks"].([]any)[0].(map[string]any)
	if table["kind"] != "table" || table["format"] != "psv" || table["separator"] != "|" || table["closed"] != true || table["header"] != true {
		t.Fatalf("table identity = %#v", table)
	}
	if table["content"] != "|Name |Description\n\n|Parser |Builds the *AST*\n" {
		t.Fatalf("content = %#v", table["content"])
	}
	columns := table["columns"].([]any)
	if len(columns) != 2 || columns[0].(map[string]any)["spec"] != "1" || columns[1].(map[string]any)["spec"] != "2" {
		t.Fatalf("columns = %#v", columns)
	}
	rows := table["rows"].([]any)
	if len(rows) != 2 || rows[0].(map[string]any)["kind"] != "table_row" || rows[0].(map[string]any)["header"] != true {
		t.Fatalf("rows = %#v", rows)
	}
	cell := rows[1].(map[string]any)["cells"].([]any)[1].(map[string]any)
	if cell["kind"] != "table_cell" || cell["text"] != "Builds the *AST*" {
		t.Fatalf("cell = %#v", cell)
	}
	inlines := cell["inlines"].([]any)
	if len(inlines) != 2 || inlines[1].(map[string]any)["kind"] != "strong" {
		t.Fatalf("cell inlines = %#v", inlines)
	}
	if _, ok := table["openingSource"]; !ok {
		t.Fatalf("openingSource missing: %#v", table)
	}
	if _, ok := table["closingSource"]; !ok {
		t.Fatalf("closingSource missing: %#v", table)
	}
}

func TestWriteJSONOmitsTableClosingSourceWhenUnclosed(t *testing.T) {
	result := asciidoc.Parse([]byte("|===\n|cell"))
	var output bytes.Buffer
	if err := format.WriteJSON(&output, result); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	table := decoded["document"].(map[string]any)["blocks"].([]any)[0].(map[string]any)
	if table["closed"] != false {
		t.Fatalf("closed = %#v", table["closed"])
	}
	if _, exists := table["closingSource"]; exists {
		t.Fatalf("unclosed table has closingSource: %#v", table)
	}
}
