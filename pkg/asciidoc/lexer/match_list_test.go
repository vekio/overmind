package lexer

import "testing"

func TestMatchListItem(t *testing.T) {
	tests := []struct {
		raw  string
		want ListPayload
	}{
		{raw: "* water", want: ListPayload{Kind: ListUnordered, Level: 1, Marker: "*", MarkerByteOffset: 0, Principal: "water", PrincipalByteOffset: 2}},
		{raw: "** earth", want: ListPayload{Kind: ListUnordered, Level: 2, Marker: "**", MarkerByteOffset: 0, Principal: "earth", PrincipalByteOffset: 3}},
		{raw: "***\titem  ", want: ListPayload{Kind: ListUnordered, Level: 3, Marker: "***", MarkerByteOffset: 0, Principal: "item  ", PrincipalByteOffset: 4}},
		{raw: ". first", want: ListPayload{Kind: ListOrdered, Level: 1, Marker: ".", MarkerByteOffset: 0, Principal: "first", PrincipalByteOffset: 2}},
		{raw: ".. second", want: ListPayload{Kind: ListOrdered, Level: 2, Marker: "..", MarkerByteOffset: 0, Principal: "second", PrincipalByteOffset: 3}},
	}

	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			token, matched := matchListItem(test.raw)
			if !matched || token.Kind != LineListItem || token.Raw != test.raw || token.List != test.want {
				t.Fatalf("matchListItem(%q) = (%+v, %t), want payload %+v", test.raw, token, matched, test.want)
			}
		})
	}
}

func TestMatchListItemRejectsUnsupportedForms(t *testing.T) {
	for _, raw := range []string{"*", "*   ", "*missing", "..missing", "*. mixed", "- item", " - indented", "+ item"} {
		if _, matched := matchListItem(raw); matched {
			t.Errorf("matchListItem(%q) matched, want false", raw)
		}
	}
}
