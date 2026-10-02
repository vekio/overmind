package asciidoc_test

import (
	"strings"
	"testing"
	"testing/iotest"

	"github.com/vekio/overmind/pkg/asciidoc"
	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
)

func TestProcessReturnsASTAnalysisAndCombinedDiagnostics(t *testing.T) {
	source := []byte("= Note\n:note-type: youtube\n\nSee <<missing>>.\n")
	result := asciidoc.Process(source)
	if result.Document == nil || result.Analysis.Document != result.Document {
		t.Fatalf("document = %p, analysis document = %p", result.Document, result.Analysis.Document)
	}
	if value, ok := result.Analysis.Header.Attributes.Lookup("note-type"); !ok || value != "youtube" {
		t.Fatalf("note-type = %q, %t", value, ok)
	}
	if result.HasErrors() || len(result.ParseDiagnostics) != 0 || len(result.Diagnostics) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if result.Diagnostics[0].Severity != diagnostic.SeverityWarning || !strings.Contains(result.Diagnostics[0].Message, "unresolved internal reference") {
		t.Fatalf("diagnostic = %+v", result.Diagnostics[0])
	}
}

func TestProcessCombinesParserAndSemanticErrorsInSourceOrder(t *testing.T) {
	source := []byte("[[duplicate]]\n== First\n\n[[duplicate]]\n== Second\n\n= Late title\n")
	result := asciidoc.Process(source)
	if !result.HasErrors() || len(result.ParseDiagnostics) != 1 || len(result.Analysis.Diagnostics) != 1 || len(result.Diagnostics) != 2 {
		t.Fatalf("result = %+v", result)
	}
	if result.Diagnostics[0].Source.Start.Line != 4 || result.Diagnostics[1].Source.Start.Line != 7 {
		t.Fatalf("diagnostics are not in source order: %+v", result.Diagnostics)
	}
}

func TestProcessReaderPreservesPartialASTAfterReadError(t *testing.T) {
	result := asciidoc.ProcessReader(iotest.TimeoutReader(strings.NewReader(":toc:")))
	if !result.HasErrors() || result.Document == nil || !result.Analysis.Header.Attributes.Has("toc") {
		t.Fatalf("result = %+v", result)
	}
}
