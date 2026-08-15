package semantic_test

import (
	"reflect"
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc"
	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
	"git.casta.me/alberto/overmind/pkg/asciidoc/semantic"
)

func TestAnalyzeBuildsEffectiveHeaderAttributesAndHistory(t *testing.T) {
	parsed := asciidoc.Parse([]byte("= Video note\n:note-type: draft\n:youtube-url:\n:note-type: youtube\n:temporary: value\n:temporary!:\n\n== Summary\n\nPending.\n\n:body-only: ignored"))
	analysis := asciidoc.Analyze(parsed.Document)

	if analysis.Document != parsed.Document || analysis.Header.Title == nil || analysis.Header.Title.Text != "Video note" {
		t.Fatalf("analysis header = %+v", analysis.Header)
	}
	if value, ok := analysis.Header.Attributes.Lookup("note-type"); !ok || value != "youtube" {
		t.Fatalf("note-type = %q, %t, want youtube, true", value, ok)
	}
	if value, ok := analysis.Header.Attributes.Lookup("youtube-url"); !ok || value != "" {
		t.Fatalf("youtube-url = %q, %t, want empty, true", value, ok)
	}
	if analysis.Header.Attributes.Has("temporary") || analysis.Header.Attributes.Has("body-only") {
		t.Fatalf("effective attributes = %+v", analysis.Header.Attributes.All())
	}

	history := analysis.Header.Attributes.History("note-type")
	if len(history) != 2 || history[0].Value != "draft" || history[1].Value != "youtube" {
		t.Fatalf("note-type history = %+v", history)
	}
	effective, ok := analysis.Header.Attributes.Get("note-type")
	if !ok || effective.Source != history[1].Source || effective.Operation != ast.AttributeSet {
		t.Fatalf("effective note-type = %+v, %t", effective, ok)
	}
	if got := []string{analysis.Header.Attributes.All()[0].Name, analysis.Header.Attributes.All()[1].Name}; !reflect.DeepEqual(got, []string{"youtube-url", "note-type"}) {
		t.Fatalf("effective order = %v", got)
	}
	if len(analysis.Diagnostics) != 0 {
		t.Fatalf("Diagnostics = %+v", analysis.Diagnostics)
	}
}

func TestAnalyzeAcceptsNilDocument(t *testing.T) {
	analysis := semantic.Analyze(nil)
	if analysis.Document != nil || analysis.Header.Title != nil || len(analysis.Header.Attributes.All()) != 0 || len(analysis.Diagnostics) != 0 {
		t.Fatalf("analysis = %+v", analysis)
	}
}

func TestApplicationLocatesSectionWithWalk(t *testing.T) {
	document := asciidoc.Parse([]byte("= Note\n\n== Notes\n\nText.\n\n== Summary\n\nPending.")).Document
	var summary *ast.Section
	ast.Walk(document, func(node ast.Node) bool {
		section, ok := node.(*ast.Section)
		if ok && section.Title == "Summary" {
			summary = section
			return false
		}
		return true
	})
	if summary == nil || summary.Title != "Summary" || len(summary.Blocks) != 1 {
		t.Fatalf("summary = %+v", summary)
	}
}
