package parser_test

import (
	"testing"

	"github.com/vekio/overmind/pkg/asciidoc/diagnostic"
	parserpkg "github.com/vekio/overmind/pkg/asciidoc/parser"
)

func TestResultHasErrors(t *testing.T) {
	result := parserpkg.Result{Diagnostics: []diagnostic.Diagnostic{{Severity: diagnostic.SeverityWarning}}}
	if result.HasErrors() {
		t.Fatal("warning-only result reports errors")
	}
	result.Diagnostics = append(result.Diagnostics, diagnostic.Diagnostic{Severity: diagnostic.SeverityError})
	if !result.HasErrors() {
		t.Fatal("error result does not report errors")
	}
}
