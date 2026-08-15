package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/diagnostic"
)

func TestParsePSVTableRowsCellsMetadataAndInlines(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "table.adoc"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	result := parse(source)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	table := requireBlock[*ast.Table](t, result.Document.Blocks[0])
	if table.Format != ast.TableFormatPSV || table.Separator != "|" || table.Marker != "|===" || !table.Closed || !table.Header {
		t.Fatalf("table identity = %+v", table)
	}
	if table.Metadata.Title == nil || table.Metadata.Title.Text != "Component inventory" || table.Metadata.Anchor == nil || table.Metadata.Anchor.ID != "components" {
		t.Fatalf("metadata = %+v", table.Metadata)
	}
	if len(table.Columns) != 2 || table.Columns[0].Spec != "1" || table.Columns[1].Spec != "2" {
		t.Fatalf("columns = %+v", table.Columns)
	}
	if len(table.Rows) != 3 || !table.Rows[0].Header || table.Rows[1].Header || len(table.Rows[2].Cells) != 2 {
		t.Fatalf("rows = %+v", table.Rows)
	}
	if table.Rows[0].Cells[0].Text != "Name" || table.Rows[0].Cells[1].Text != "Description" {
		t.Fatalf("header cells = %+v", table.Rows[0].Cells)
	}
	strong := requireInline[*ast.Strong](t, table.Rows[1].Cells[1].Inlines[1])
	if inlineText(t, strong.Children) != "physical" {
		t.Fatalf("strong cell inline = %+v", strong)
	}
	monospace := requireInline[*ast.Monospace](t, table.Rows[2].Cells[1].Inlines[1])
	if inlineText(t, monospace.Children) != "AST" {
		t.Fatalf("monospace cell inline = %+v", monospace)
	}
	cell := table.Rows[2].Cells[1]
	if string(source[cell.MarkerSource.Start.Offset:cell.MarkerSource.End.Offset]) != "|" || string(source[cell.ContentSource.Start.Offset:cell.ContentSource.End.Offset]) != "Builds the `AST`" {
		t.Fatalf("cell spans = marker %s content %s", cell.MarkerSource, cell.ContentSource)
	}
}

func TestParseTableInfersColumnsAndImplicitHeader(t *testing.T) {
	result := parse([]byte("|===\n|Name |Value\n\n|one |two\n|==="))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	table := requireBlock[*ast.Table](t, result.Document.Blocks[0])
	if len(table.Columns) != 2 || len(table.Rows) != 2 || !table.Header || !table.Rows[0].Header {
		t.Fatalf("table = %+v", table)
	}
}

func TestParseTableColumnMultiplierAndRaggedRowWarning(t *testing.T) {
	result := parse([]byte("[cols=3*]\n|===\n|one |two |three\n|four |five\n|==="))
	if result.HasErrors() || len(result.Diagnostics) != 1 || !strings.Contains(result.Diagnostics[0].Message, "2 cells; expected 3") {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	table := requireBlock[*ast.Table](t, result.Document.Blocks[0])
	if len(table.Columns) != 3 || len(table.Rows) != 2 || len(table.Rows[1].Cells) != 2 {
		t.Fatalf("table = %+v", table)
	}
}

func TestParseUnclosedTableReturnsPartialAST(t *testing.T) {
	source := []byte("[cols=2*]\n|===\n|one |two\n")
	result := parse(source)
	if !result.HasErrors() || len(result.Diagnostics) != 1 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	item := result.Diagnostics[0]
	if item.Severity != diagnostic.SeverityError || item.Message != `unclosed table; expected "|==="` {
		t.Fatalf("diagnostic = %+v", item)
	}
	table := requireBlock[*ast.Table](t, result.Document.Blocks[0])
	if table.Closed || table.ClosingSource != (ast.Span{}) || table.Source.End.Offset != len(source) || len(table.Rows) != 1 {
		t.Fatalf("partial table = %+v", table)
	}
}

func TestParseUnsupportedCSVTablePreservesContent(t *testing.T) {
	result := parse([]byte(",===\nname,value\none,two\n,==="))
	if result.HasErrors() || len(result.Diagnostics) != 1 || !strings.Contains(result.Diagnostics[0].Message, "csv table data") {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	table := requireBlock[*ast.Table](t, result.Document.Blocks[0])
	if table.Format != ast.TableFormatCSV || table.Separator != "," || table.Content != "name,value\none,two\n" || len(table.Rows) != 0 {
		t.Fatalf("CSV table = %+v", table)
	}
}

func TestParseTablePreservesUnicodeCRLFSpans(t *testing.T) {
	source := []byte("[cols=2*]\r\n|===\r\n|Café |*valor*\r\n|===\r\n")
	result := parse(source)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	table := requireBlock[*ast.Table](t, result.Document.Blocks[0])
	if table.Content != "|Café |*valor*\r\n" || len(table.Rows) != 1 {
		t.Fatalf("table = %+v", table)
	}
	cell := table.Rows[0].Cells[1]
	if cell.Text != "*valor*" || string(source[cell.ContentSource.Start.Offset:cell.ContentSource.End.Offset]) != "*valor*" {
		t.Fatalf("cell = %+v", cell)
	}
	if cell.ContentSource.Start != (ast.Position{Offset: 25, Line: 3, Column: 8}) {
		t.Fatalf("cell content starts at %s, want 3:8@25", cell.ContentSource.Start)
	}
	strong := requireInline[*ast.Strong](t, cell.Inlines[0])
	if string(source[strong.Source.Start.Offset:strong.Source.End.Offset]) != "*valor*" {
		t.Fatalf("strong source = %s", strong.Source)
	}
}

func TestParseTableSupportsCustomSingleByteSeparator(t *testing.T) {
	result := parse([]byte("[cols=2*,separator=;]\n|===\n;one ;two\n|==="))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	table := requireBlock[*ast.Table](t, result.Document.Blocks[0])
	if table.Separator != ";" || len(table.Rows) != 1 || table.Rows[0].Cells[1].Text != "two" {
		t.Fatalf("table = %+v", table)
	}
}

func TestParseTableInfersImplicitHeaderWithCREndings(t *testing.T) {
	result := parse([]byte("|===\r|Name |Value\r\r|one |two\r|==="))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", result.Diagnostics)
	}
	table := requireBlock[*ast.Table](t, result.Document.Blocks[0])
	if !table.Header || len(table.Rows) != 2 || !table.Rows[0].Header {
		t.Fatalf("table = %+v", table)
	}
}
