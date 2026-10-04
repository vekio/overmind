package codecs_test

import (
	"strings"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/infra/codecs"
)

func TestHabitDecodeReadsMetadataOnlyFromHeader(t *testing.T) {
	source := []byte(`= Beber agua
:overmind-id: 11111111-1111-4111-8111-111111111111
:overmind-type: habit
:overmind-amount: 2
:overmind-unit: litros
:overmind-period: day
:overmind-tags: Salud
:overmind-created-at: 2026-10-04T12:00:00Z
:overmind-updated-at: 2026-10-04T12:00:00Z

:overmind-type: person
:overmind-id: 00000000-0000-0000-0000-000000000000
----
Unclosed block in the body
`)
	habit, err := (codecs.HabitCodec{}).Decode(source)
	if err != nil {
		t.Fatal(err)
	}
	if habit.ID() != uuid.MustParse("11111111-1111-4111-8111-111111111111") || habit.Goal().Amount() != 2 || habit.Title().String() != "Beber agua" || habit.Tags().Strings()[0] != "salud" {
		t.Fatalf("unexpected decoded habit: %s", habit.Summary())
	}
	for _, replacement := range []struct{ old, new string }{
		{":overmind-type: habit", ":overmind-type: person"},
		{":overmind-amount: 2", ":overmind-amount: -2"},
		{":overmind-created-at: 2026-10-04T12:00:00Z", ":overmind-created-at: invalid"},
	} {
		invalid := []byte(strings.Replace(string(source), replacement.old, replacement.new, 1))
		if _, err := (codecs.HabitCodec{}).Decode(invalid); err == nil {
			t.Fatalf("accepted invalid header: %s", replacement.new)
		}
	}
}
