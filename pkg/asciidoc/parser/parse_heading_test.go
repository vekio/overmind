package parser_test

import "testing"

func TestParseRejectsDocumentTitleAfterBody(t *testing.T) {
	result := parse([]byte("body\n\n= Late title\n"))
	if !result.HasErrors() || len(result.Diagnostics) != 1 {
		t.Fatalf("result diagnostics = %+v, want one error", result.Diagnostics)
	}
	if result.Document.Title != nil {
		t.Fatalf("late document title was accepted: %+v", result.Document.Title)
	}
}

func TestParseRejectsDocumentTitleAfterAttributeHeaderEnds(t *testing.T) {
	result := parse([]byte(":toc:\n\n= Late title\n"))
	if !result.HasErrors() || result.Document.Title != nil {
		t.Fatalf("result = %+v, want late-title error", result)
	}
}
