package semantic_test

import (
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc"
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/diagnostic"
	"git.casta.me/alberto/overmind/pkg/asciidoc/semantic"
)

func TestAnalyzeIndexesAnchorsAndResolvesForwardInternalReferences(t *testing.T) {
	parsed := asciidoc.Parse([]byte("See <<summary>> and xref:#code[the code].\n\n[[summary,Summary section]]\n== Summary\n\n[[code]]\n[source,go]\n----\nfmt.Println()\n----"))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parser diagnostics = %+v", parsed.Diagnostics)
	}
	analysis := asciidoc.Analyze(parsed.Document)
	if len(analysis.Diagnostics) != 0 || analysis.HasErrors() {
		t.Fatalf("analysis diagnostics = %+v", analysis.Diagnostics)
	}
	anchors := analysis.Anchors.All()
	if len(anchors) != 2 || anchors[0].ID != "summary" || anchors[1].ID != "code" {
		t.Fatalf("anchors = %+v", anchors)
	}
	if _, ok := anchors[0].Node.(*ast.Section); !ok {
		t.Fatalf("summary owner = %T, want *ast.Section", anchors[0].Node)
	}
	if _, ok := anchors[1].Node.(*ast.DelimitedBlock); !ok {
		t.Fatalf("code owner = %T, want *ast.DelimitedBlock", anchors[1].Node)
	}
	if len(analysis.References) != 2 {
		t.Fatalf("references = %+v", analysis.References)
	}
	for _, reference := range analysis.References {
		if reference.Status != semantic.ReferenceResolved || reference.Anchor == nil || reference.Anchor.ID != reference.LocalID {
			t.Fatalf("reference = %+v", reference)
		}
	}
}

func TestAnalyzeReportsDuplicateAnchorsAndKeepsFirst(t *testing.T) {
	parsed := asciidoc.Parse([]byte("[[same]]\n== First\n\n[[same]]\n== Second\n\nSee <<same>>."))
	analysis := asciidoc.Analyze(parsed.Document)
	if !analysis.HasErrors() || len(analysis.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %+v", analysis.Diagnostics)
	}
	item := analysis.Diagnostics[0]
	if item.Severity != diagnostic.SeverityError || !strings.Contains(item.Message, `duplicate anchor "same"`) || item.Source.Start.Line != 4 {
		t.Fatalf("diagnostic = %+v", item)
	}
	history := analysis.Anchors.History("same")
	anchor, ok := analysis.Anchors.Lookup("same")
	if !ok || len(history) != 2 || anchor.Source != history[0].Source {
		t.Fatalf("anchor = %+v, history = %+v", anchor, history)
	}
	if len(analysis.References) != 1 || analysis.References[0].Anchor == nil || analysis.References[0].Anchor.Source != history[0].Source {
		t.Fatalf("reference = %+v", analysis.References)
	}
}

func TestAnalyzeWarnsForUnresolvedLocalReferenceAndIgnoresExternalTargets(t *testing.T) {
	parsed := asciidoc.Parse([]byte("See <<missing>>, xref:other.adoc#section[external], and xref:chapter.adoc[]."))
	analysis := asciidoc.Analyze(parsed.Document)
	if analysis.HasErrors() || len(analysis.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %+v", analysis.Diagnostics)
	}
	if item := analysis.Diagnostics[0]; item.Severity != diagnostic.SeverityWarning || item.Message != `unresolved internal reference "missing"` || item.Source.Start.Line != 1 {
		t.Fatalf("diagnostic = %+v", item)
	}
	if len(analysis.References) != 3 || analysis.References[0].Status != semantic.ReferenceUnresolved || analysis.References[1].Status != semantic.ReferenceExternal || analysis.References[2].Status != semantic.ReferenceExternal {
		t.Fatalf("references = %+v", analysis.References)
	}
}

func TestReferenceStatusString(t *testing.T) {
	for _, test := range []struct {
		status semantic.ReferenceStatus
		want   string
	}{
		{semantic.ReferenceUnknown, "unknown"},
		{semantic.ReferenceResolved, "resolved"},
		{semantic.ReferenceUnresolved, "unresolved"},
		{semantic.ReferenceExternal, "external"},
		{semantic.ReferenceStatus(255), "unknown"},
	} {
		if got := test.status.String(); got != test.want {
			t.Errorf("String() = %q, want %q", got, test.want)
		}
	}
}
