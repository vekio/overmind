package lexer

import "testing"

func TestMatchDescriptionListItem(t *testing.T) {
	for _, test := range []struct {
		raw  string
		want DescriptionListPayload
	}{
		{raw: "CPU:: The processor", want: DescriptionListPayload{Term: "CPU", TermByteOffset: 0, Marker: "::", MarkerByteOffset: 3, Description: "The processor", DescriptionByteOffset: 6}},
		{raw: " Linux ::: kernel", want: DescriptionListPayload{Term: "Linux", TermByteOffset: 1, Marker: ":::", MarkerByteOffset: 7, Description: "kernel", DescriptionByteOffset: 11}},
		{raw: "Term::", want: DescriptionListPayload{Term: "Term", TermByteOffset: 0, Marker: "::", MarkerByteOffset: 4, DescriptionByteOffset: 6}},
	} {
		token, matched := matchDescriptionListItem(test.raw)
		if !matched || token.Kind != LineDescriptionListItem || token.Raw != test.raw || token.DescriptionList != test.want {
			t.Errorf("matchDescriptionListItem(%q) = (%+v, %t), want payload %+v", test.raw, token, matched, test.want)
		}
	}
}

func TestMatchDescriptionListItemRejectsInvalidForms(t *testing.T) {
	for _, raw := range []string{":: description", "term: description", "term::description", "image::target[]"} {
		if _, matched := matchDescriptionListItem(raw); matched {
			t.Errorf("matchDescriptionListItem(%q) matched, want false", raw)
		}
	}
}
