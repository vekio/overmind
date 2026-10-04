package codecs_test

import (
	"strings"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/infra/codecs"
)

func TestPersonDecodeReadsMetadataOnlyFromHeader(t *testing.T) {
	source := []byte(`= Ana
:overmind-id: 22222222-2222-4222-8222-222222222222
:overmind-type: person
:overmind-groups: Amigos
:overmind-tags: Contactos
:overmind-created-at: 2026-10-04T12:00:00Z
:overmind-updated-at: 2026-10-04T12:00:00Z

:overmind-type: habit
:overmind-id: 00000000-0000-0000-0000-000000000000
----
Unclosed block in the body
`)
	person, err := (codecs.PersonCodec{}).Decode(source)
	if err != nil {
		t.Fatal(err)
	}
	if person.ID() != uuid.MustParse("22222222-2222-4222-8222-222222222222") || person.Name().String() != "Ana" || person.Groups().Strings()[0] != "amigos" || person.Tags().Strings()[0] != "contactos" {
		t.Fatal("unexpected decoded person")
	}
	for _, replacement := range []struct{ old, new string }{
		{":overmind-type: person", ":overmind-type: habit"},
		{":overmind-groups: Amigos", ":overmind-groups: !!!"},
		{"22222222-2222-4222-8222-222222222222", "00000000-0000-0000-0000-000000000000"},
	} {
		invalid := []byte(strings.Replace(string(source), replacement.old, replacement.new, 1))
		if _, err := (codecs.PersonCodec{}).Decode(invalid); err == nil {
			t.Fatalf("accepted invalid header: %s", replacement.new)
		}
	}
}
