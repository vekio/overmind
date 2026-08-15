package lexer

import "testing"

func TestMatchBlockTitle(t *testing.T) {
	token, matched := matchBlockTitle(".Terminal Output")
	want := BlockTitlePayload{Title: "Terminal Output", TitleByteOffset: 1}
	if !matched || token.Kind != LineBlockTitle || token.Raw != ".Terminal Output" || token.BlockTitle != want {
		t.Fatalf("matchBlockTitle() = (%+v, %t), want payload %+v", token, matched, want)
	}
}

func TestMatchBlockTitleRejectsListAndEmptyForms(t *testing.T) {
	for _, raw := range []string{".", ". ordered item", " .Indented"} {
		if _, matched := matchBlockTitle(raw); matched {
			t.Errorf("matchBlockTitle(%q) matched, want false", raw)
		}
	}
}
