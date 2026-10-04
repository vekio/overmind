package cli

import (
	"strings"
	"testing"
)

func TestListCommandFiltersPersistedNotes(t *testing.T) {
	fixture := newCommandFixture(t)
	creates := [][]string{
		{"habit", "--amount", "2", "--unit", "litros", "--period", "day", "--tag", "salud", "Beber agua"},
		{"habit", "--amount", "20", "--unit", "paginas", "--period", "day", "Leer"},
		{"person", "--tag", "salud", "Ana"},
		{"bookmark", "https://example.com"},
		{"inbox", "Una idea\nOtra línea"},
		{"page", "Plan"},
		{"journal", "--date", "2026-10-04"},
	}
	for _, args := range creates {
		if _, err := fixture.run(args, ""); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name               string
		args               []string
		count              int
		contains, excludes []string
	}{
		{name: "all types", count: 7, contains: []string{"Beber agua", "Leer", "Ana", "https://example.com", "Una idea Otra línea", "Plan", "2026-10-04"}},
		{name: "type", args: []string{"--type", "habit"}, count: 2, contains: []string{"Beber agua", "Leer"}, excludes: []string{"Ana"}},
		{name: "tag", args: []string{"--tag", "Sálud"}, count: 2, contains: []string{"Beber agua", "Ana"}, excludes: []string{"Leer"}},
		{name: "combined", args: []string{"--type", "habit", "--tag", "salud"}, count: 1, contains: []string{"Beber agua"}, excludes: []string{"Leer", "Ana"}},
		{name: "empty", args: []string{"--tag", "unknown"}, count: 0, contains: []string{"No notes found."}},
		{name: "page", args: []string{"--limit", "2", "--offset", "2"}, count: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := fixture.run(append([]string{"list"}, tc.args...), "")
			if err != nil {
				t.Fatal(err)
			}
			if tc.count > 0 && len(strings.Split(strings.TrimSpace(output), "\n")) != tc.count+1 {
				t.Fatalf("output=%s", output)
			}
			for _, value := range tc.contains {
				if !strings.Contains(output, value) {
					t.Fatalf("missing %q: %s", value, output)
				}
			}
			for _, value := range tc.excludes {
				if strings.Contains(output, value) {
					t.Fatalf("unexpected %q: %s", value, output)
				}
			}
		})
	}
	for _, args := range [][]string{{"list", "--type", "unknown"}, {"list", "--tag", "!!!"}, {"list", "--offset", "-1"}, {"list", "unexpected"}} {
		if _, err := fixture.run(args, ""); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestListCommandPaginationDoesNotRepeatNotes(t *testing.T) {
	fixture := newCommandFixture(t)
	for _, title := range []string{"First", "Second", "Third"} {
		if _, err := fixture.run([]string{"page", title}, ""); err != nil {
			t.Fatal(err)
		}
	}
	var ids []string
	for _, offset := range []string{"0", "1", "2"} {
		output, err := fixture.run([]string{"list", "--limit", "1", "--offset", offset}, "")
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 2 {
			t.Fatal(output)
		}
		ids = append(ids, strings.Fields(lines[1])[0])
	}
	if ids[0] == ids[1] || ids[1] == ids[2] || ids[0] == ids[2] {
		t.Fatalf("repeated IDs: %v", ids)
	}
}
