package diagnostic

import (
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/ast"
)

func TestSeverityAndDiagnosticString(t *testing.T) {
	for _, test := range []struct {
		severity Severity
		want     string
	}{
		{SeverityUnknown, "unknown"},
		{SeverityWarning, "warning"},
		{SeverityError, "error"},
		{Severity(255), "unknown"},
	} {
		if got := test.severity.String(); got != test.want {
			t.Errorf("Severity(%d).String() = %q, want %q", test.severity, got, test.want)
		}
	}

	item := Diagnostic{
		Severity: SeverityError,
		Message:  "bad syntax",
		Source: ast.Span{
			Start: ast.Position{Line: 2, Column: 1, Offset: 4},
			End:   ast.Position{Line: 2, Column: 3, Offset: 6},
		},
	}
	if got, want := item.String(), "error: bad syntax at [2:1@4, 2:3@6)"; got != want {
		t.Fatalf("Diagnostic.String() = %q, want %q", got, want)
	}
}
