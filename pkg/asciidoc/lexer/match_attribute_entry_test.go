package lexer

import "testing"

func TestMatchAttributeEntry(t *testing.T) {
	tests := []struct {
		raw  string
		want AttributeEntryPayload
	}{
		{raw: ":toc:", want: AttributeEntryPayload{Operation: AttributeSet, Name: "toc", NameByteOffset: 1, ValueByteOffset: 5}},
		{raw: ":toc: left", want: AttributeEntryPayload{Operation: AttributeSet, Name: "toc", NameByteOffset: 1, Value: "left", ValueByteOffset: 6}},
		{raw: ":name:\t value ", want: AttributeEntryPayload{Operation: AttributeSet, Name: "name", NameByteOffset: 1, Value: "value ", ValueByteOffset: 8}},
		{raw: ":!toc:", want: AttributeEntryPayload{Operation: AttributeUnset, Name: "toc", NameByteOffset: 2, ValueByteOffset: 6}},
		{raw: ":toc!:", want: AttributeEntryPayload{Operation: AttributeUnset, Name: "toc", NameByteOffset: 1, ValueByteOffset: 6}},
	}

	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			token, matched := matchAttributeEntry(test.raw)
			if !matched || token.Kind != LineAttributeEntry || token.Raw != test.raw || token.AttributeEntry != test.want {
				t.Fatalf("matchAttributeEntry(%q) = (%+v, %t), want payload %+v", test.raw, token, matched, test.want)
			}
		})
	}
}

func TestMatchAttributeEntryRejectsInvalidForms(t *testing.T) {
	for _, raw := range []string{":", "::", ":name", ":bad name:", ":bad.name:", ":$name:", ":!name!: ", ":!name: value", ":name:value"} {
		if _, matched := matchAttributeEntry(raw); matched {
			t.Errorf("matchAttributeEntry(%q) matched, want false", raw)
		}
	}
}

func TestAttributeOperationString(t *testing.T) {
	if got := AttributeOperationUnknown.String(); got != "UNKNOWN" {
		t.Fatalf("AttributeOperationUnknown.String() = %q, want UNKNOWN", got)
	}
	if got := AttributeSet.String(); got != "SET" {
		t.Fatalf("AttributeSet.String() = %q, want SET", got)
	}
	if got := AttributeUnset.String(); got != "UNSET" {
		t.Fatalf("AttributeUnset.String() = %q, want UNSET", got)
	}
	if got := AttributeOperation(255).String(); got != "UNKNOWN" {
		t.Fatalf("unknown AttributeOperation.String() = %q, want UNKNOWN", got)
	}
}
