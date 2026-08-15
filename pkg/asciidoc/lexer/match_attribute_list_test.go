package lexer

import (
	"reflect"
	"testing"
)

func TestMatchAttributeList(t *testing.T) {
	tests := []struct {
		raw     string
		entries []AttributeListEntry
	}{
		{raw: "[source,go]", entries: []AttributeListEntry{{Value: "source", ValueByteOffset: 1}, {Value: "go", ValueByteOffset: 8}}},
		{raw: "[,ruby]", entries: []AttributeListEntry{{ValueByteOffset: 1}, {Value: "ruby", ValueByteOffset: 2}}},
		{raw: "[#identifier]", entries: []AttributeListEntry{{Value: "#identifier", ValueByteOffset: 1}}},
		{raw: `[cols="1,1", options="header"]`, entries: []AttributeListEntry{{Value: `cols="1,1"`, ValueByteOffset: 1}, {Value: `options="header"`, ValueByteOffset: 13}}},
		{raw: "[]", entries: nil},
	}

	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			token, matched := matchAttributeList(test.raw)
			if !matched || token.Kind != LineAttributeList || token.Raw != test.raw {
				t.Fatalf("matchAttributeList(%q) = (%+v, %t), want attribute-list token", test.raw, token, matched)
			}
			if !reflect.DeepEqual(token.AttributeList.Entries, test.entries) {
				t.Fatalf("Entries = %#v, want %#v", token.AttributeList.Entries, test.entries)
			}
		})
	}
}

func TestMatchAttributeListRejectsIncompleteLines(t *testing.T) {
	for _, raw := range []string{"[source,go", "source,go]", "[source,go] trailing", "prefix [source,go]"} {
		if _, matched := matchAttributeList(raw); matched {
			t.Errorf("matchAttributeList(%q) matched, want false", raw)
		}
	}
}
