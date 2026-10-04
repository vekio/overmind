package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"uuid"
)

func TestReindexRestoresAllProjectionsFromFilesAndRemovesStaleNotes(t *testing.T) {
	fixture := newCommandFixture(t)
	specs := []struct {
		kind string
		args []string
		tags []string
	}{
		{"habit", []string{"habit", "--amount", "2", "--unit", "litros", "--period", "day", "--tag", "salud", "Agua"}, []string{"salud"}},
		{"person", []string{"person", "--group", "amigos", "--tag", "salud", "Ana"}, []string{"salud"}},
		{"bookmark", []string{"bookmark", "https://example.com"}, nil},
		{"inbox", []string{"inbox", "--", "\nUna idea\n:overmind-type: body\n\nOtra línea.\n"}, nil},
		{"page", []string{"page", "--area", "work/ideas", "Plan"}, nil},
		{"journal", []string{"journal", "--date", "2026-10-04"}, nil},
	}
	var documents []createdNote
	for _, spec := range specs {
		documents = append(documents, fixture.create(spec.kind, spec.args, "", spec.tags))
	}
	stale := fixture.create("page", []string{"page", "Deleted"}, "", nil)
	if err := os.Remove(stale.Path); err != nil {
		t.Fatal(err)
	}
	db := fixture.db()
	// Leave only a stale index row; all six real notes must be recovered from files.
	if _, err := db.Exec("DELETE FROM notes WHERE id <> ?", stale.ID.String()); err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		output, err := fixture.run([]string{"reindex"}, "")
		if err != nil || output != "Indexed 6 notes.\n" {
			t.Fatalf("output=%q err=%v", output, err)
		}
		var count int
		if err := db.QueryRow("SELECT count(*) FROM notes").Scan(&count); err != nil || count != 6 {
			t.Fatalf("count=%d err=%v", count, err)
		}
		if err := db.QueryRow("SELECT count(*) FROM notes WHERE id=?", stale.ID.String()).Scan(&count); err != nil || count != 0 {
			t.Fatal("stale note survived")
		}
		for _, document := range documents {
			source, err := os.ReadFile(document.Path)
			if err != nil || !bytes.Equal(source, document.Source) {
				t.Fatal("reindex modified a source file")
			}
			var path, created, updated string
			if err := db.QueryRow("SELECT path,created_at,updated_at FROM notes WHERE id=?", document.ID.String()).Scan(&path, &created, &updated); err != nil {
				t.Fatal(err)
			}
			if path != document.Path || created != updated {
				t.Fatal("metadata not recovered")
			}
		}
		var amount float64
		if err := db.QueryRow("SELECT target_amount FROM habits").Scan(&amount); err != nil || amount != 2 {
			t.Fatalf("amount=%v err=%v", amount, err)
		}
		var groups, tag, area, content, url, date string
		for _, item := range []struct {
			sql   string
			value *string
			want  string
		}{
			{"SELECT group_name FROM person_groups", &groups, "amigos"},
			{"SELECT tag FROM note_tags WHERE note_id IN (SELECT note_id FROM habits)", &tag, "salud"},
			{"SELECT area FROM pages", &area, "work/ideas"},
			{"SELECT content FROM inbox", &content, "\nUna idea\n:overmind-type: body\n\nOtra línea.\n"},
			{"SELECT url FROM bookmarks", &url, "https://example.com"},
			{"SELECT date FROM journals", &date, "2026-10-04"},
		} {
			if err := db.QueryRow(item.sql).Scan(item.value); err != nil || *item.value != item.want {
				t.Fatalf("%s: value=%q err=%v", item.sql, *item.value, err)
			}
		}
	}
}

func TestReindexFailuresPreservePreviousIndex(t *testing.T) {
	for _, failure := range []string{"malformed", "filename mismatch", "duplicate ID", "duplicate journal date", "invalid habit goal", "unsupported type", "invalid filename"} {
		t.Run(failure, func(t *testing.T) {
			fixture := newCommandFixture(t)
			page := fixture.create("page", []string{"page", "Plan"}, "", nil)
			journal := fixture.create("journal", []string{"journal", "--date", "2026-10-04"}, "", nil)
			habit := fixture.create("habit", []string{"habit", "--amount", "2", "--unit", "litros", "--period", "day", "Agua"}, "", nil)
			before, err := fixture.run([]string{"list"}, "")
			if err != nil {
				t.Fatal(err)
			}
			switch failure {
			case "malformed":
				if err := os.WriteFile(page.Path, []byte("= Invalid\n"), 0644); err != nil {
					t.Fatal(err)
				}
			case "filename mismatch":
				source := bytes.Replace(page.Source, []byte(page.ID.String()), []byte(uuid.New().String()), 1)
				if err := os.WriteFile(page.Path, source, 0644); err != nil {
					t.Fatal(err)
				}
			case "duplicate ID":
				path := filepath.Join(filepath.Dir(page.Path), "copy", filepath.Base(page.Path))
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, page.Source, 0644); err != nil {
					t.Fatal(err)
				}
			case "duplicate journal date":
				id := uuid.New()
				source := bytes.Replace(journal.Source, []byte(journal.ID.String()), []byte(id.String()), 1)
				if err := os.WriteFile(filepath.Join(filepath.Dir(journal.Path), id.String()+".adoc"), source, 0644); err != nil {
					t.Fatal(err)
				}
			case "unsupported type":
				source := bytes.Replace(page.Source, []byte(":overmind-type: page"), []byte(":overmind-type: unknown"), 1)
				if err := os.WriteFile(page.Path, source, 0644); err != nil {
					t.Fatal(err)
				}
			case "invalid filename":
				if err := os.WriteFile(filepath.Join(filepath.Dir(page.Path), "invalid.adoc"), page.Source, 0644); err != nil {
					t.Fatal(err)
				}
			case "invalid habit goal":
				source := bytes.Replace(habit.Source, []byte(":overmind-amount: 2"), []byte(":overmind-amount: -2"), 1)
				if err := os.WriteFile(habit.Path, source, 0644); err != nil {
					t.Fatal(err)
				}
			}
			if output, err := fixture.run([]string{"reindex"}, ""); err == nil || output != "" || !strings.Contains(err.Error(), "previous index preserved") {
				t.Fatalf("output=%q err=%v", output, err)
			}
			after, err := fixture.run([]string{"list"}, "")
			if err != nil || after != before {
				t.Fatalf("index changed: %s err=%v", after, err)
			}
		})
	}
}

func TestReindexEmptyVaultAndUnexpectedArguments(t *testing.T) {
	fixture := newCommandFixture(t)
	output, err := fixture.run([]string{"reindex"}, "")
	if err != nil || output != "Indexed 0 notes.\n" {
		t.Fatalf("output=%q err=%v", output, err)
	}
	if _, err := fixture.run([]string{"reindex", "unexpected"}, ""); err == nil {
		t.Fatal("unexpected argument accepted")
	}
}
