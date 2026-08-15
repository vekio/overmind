package lexer

import "testing"

func TestMatchBlockMacro(t *testing.T) {
	for _, test := range []struct {
		raw  string
		want BlockMacroPayload
	}{
		{raw: "image::sunset.jpg[]", want: BlockMacroPayload{Name: "image", NameByteOffset: 0, Target: "sunset.jpg", TargetByteOffset: 7, AttributesByteOffset: 18}},
		{raw: "include::app.rb[lines=1..3]", want: BlockMacroPayload{Name: "include", NameByteOffset: 0, Target: "app.rb", TargetByteOffset: 9, Attributes: "lines=1..3", AttributesByteOffset: 16}},
		{raw: "toc::[]", want: BlockMacroPayload{Name: "toc", NameByteOffset: 0, TargetByteOffset: 5, AttributesByteOffset: 6}},
		{raw: "custom-block::target[one,two]", want: BlockMacroPayload{Name: "custom-block", NameByteOffset: 0, Target: "target", TargetByteOffset: 14, Attributes: "one,two", AttributesByteOffset: 21}},
	} {
		token, matched := matchBlockMacro(test.raw)
		if !matched || token.Kind != LineBlockMacro || token.Raw != test.raw || token.BlockMacro != test.want {
			t.Errorf("matchBlockMacro(%q) = (%+v, %t), want payload %+v", test.raw, token, matched, test.want)
		}
	}
}

func TestMatchBlockMacroRejectsInvalidForms(t *testing.T) {
	for _, raw := range []string{"image:target[]", "1image::target[]", "bad.name::target[]", "image:: target[]", "image::target[", "image::target[] trailing"} {
		if _, matched := matchBlockMacro(raw); matched {
			t.Errorf("matchBlockMacro(%q) matched, want false", raw)
		}
	}
}
