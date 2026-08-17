package edit_test

import (
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/pkg/asciidoc/edit"
)

func TestEditorSetHeaderAttributeReplacesEffectiveDeclaration(t *testing.T) {
	source := []byte("= Page\n:updated-at: old\n:tags: go\n\nBody with :updated-at: untouched.\n")
	editor := newEditor(t, source)
	if err := editor.SetHeaderAttribute("updated-at", "new"); err != nil {
		t.Fatalf("SetHeaderAttribute() error = %v", err)
	}
	updated := editor.Bytes()
	want := "= Page\n:updated-at: new\n:tags: go\n\nBody with :updated-at: untouched.\n"
	if string(updated) != want {
		t.Fatalf("updated source = %q, want %q", updated, want)
	}
}

func TestEditorSetHeaderAttributeReplacesLastUnsetDeclaration(t *testing.T) {
	source := []byte(":updated-at: old\n:updated-at!:\n= Page\n\nBody\n")
	editor := newEditor(t, source)
	if err := editor.SetHeaderAttribute("updated-at", "new"); err != nil {
		t.Fatalf("SetHeaderAttribute() error = %v", err)
	}
	updated := editor.Bytes()
	if got, want := string(updated), ":updated-at: old\n:updated-at: new\n= Page\n\nBody\n"; got != want {
		t.Fatalf("updated source = %q, want %q", got, want)
	}
}

func TestEditorSetHeaderAttributeInsertsAfterHeaderUsingExistingLineEnding(t *testing.T) {
	for name, test := range map[string]struct {
		source string
		want   string
	}{
		"attributes": {
			source: "= Page\r\n:id: page-id\r\n\r\nBody\r\n",
			want:   "= Page\r\n:id: page-id\r\n:updated-at: now\r\n\r\nBody\r\n",
		},
		"title": {
			source: "= Page\n\nBody\n",
			want:   "= Page\n:updated-at: now\n\nBody\n",
		},
		"empty": {
			source: "",
			want:   ":updated-at: now\n",
		},
	} {
		t.Run(name, func(t *testing.T) {
			editor := newEditor(t, []byte(test.source))
			err := editor.SetHeaderAttribute("updated-at", "now")
			if updated := string(editor.Bytes()); err != nil || updated != test.want {
				t.Fatalf("SetHeaderAttribute() = (%q, %v), want %q", updated, err, test.want)
			}
		})
	}
}

func TestEditorSetHeaderAttributeRejectsInvalidInput(t *testing.T) {
	source := []byte("= Page\n")
	for _, test := range []struct {
		name  string
		value string
	}{
		{name: ""},
		{name: "invalid name"},
		{name: "updated-at", value: "first\nsecond"},
	} {
		editor := newEditor(t, source)
		if err := editor.SetHeaderAttribute(test.name, test.value); err == nil {
			t.Errorf("SetHeaderAttribute(%q, %q) error = nil", test.name, test.value)
		}
		if got := string(editor.Bytes()); got != string(source) {
			t.Errorf("source after rejected edit = %q, want %q", got, source)
		}
	}

	if _, err := edit.New([]byte("= First\n\n= Invalid\n")); err == nil || !strings.Contains(err.Error(), "document has errors") {
		t.Fatalf("New(invalid) error = %v", err)
	}
}

func TestEditorReparsesAfterEveryHeaderAttributeChange(t *testing.T) {
	editor := newEditor(t, []byte("= Page\n\nBody\n"))

	for _, attribute := range []struct {
		name  string
		value string
	}{
		{name: "first", value: "one"},
		{name: "second", value: "two"},
		{name: "first", value: "updated"},
	} {
		if err := editor.SetHeaderAttribute(attribute.name, attribute.value); err != nil {
			t.Fatalf("SetHeaderAttribute(%q, %q) error = %v", attribute.name, attribute.value, err)
		}
	}

	want := "= Page\n:first: updated\n:second: two\n\nBody\n"
	if got := string(editor.Bytes()); got != want {
		t.Fatalf("edited source = %q, want %q", got, want)
	}
}

func TestEditorOwnsItsSourceAndReturnedBytes(t *testing.T) {
	source := []byte("= Page\n")
	editor := newEditor(t, source)
	source[0] = '!'

	result := editor.Bytes()
	result[0] = '?'

	if got, want := string(editor.Bytes()), "= Page\n"; got != want {
		t.Fatalf("editor source = %q, want %q", got, want)
	}
}

func newEditor(t *testing.T, source []byte) *edit.Editor {
	t.Helper()
	editor, err := edit.New(source)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return editor
}
