package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"uuid"
)

func TestDeleteCommandRemovesEveryNoteTypeAndItsAssociations(t *testing.T) {
	for _, spec := range []struct {
		kind string
		args []string
		tags []string
	}{
		{"habit", []string{"habit", "--amount", "2", "--unit", "litros", "--period", "day", "--tag", "salud", "Agua"}, []string{"salud"}},
		{"person", []string{"person", "--group", "amigos", "--tag", "salud", "Ana"}, []string{"salud"}},
		{"bookmark", []string{"bookmark", "--tag", "salud", "https://example.com"}, []string{"salud"}},
		{"inbox", []string{"inbox", "--tag", "salud", "Una idea"}, []string{"salud"}},
		{"page", []string{"page", "--tag", "salud", "Plan"}, []string{"salud"}},
		{"journal", []string{"journal", "--tag", "salud", "--date", "2026-10-04"}, []string{"salud"}},
	} {
		t.Run(spec.kind, func(t *testing.T) {
			fixture := newCommandFixture(t)
			note := fixture.create(spec.kind, spec.args, "", spec.tags)
			survivor := fixture.create("page", []string{"page", "Keep me"}, "", nil)
			for attempt := 0; attempt < 2; attempt++ {
				output, err := fixture.run([]string{"delete", note.ID.String()}, "")
				if err != nil || output != fmt.Sprintf("Deleted note %s.\n", note.ID) {
					t.Fatalf("output=%q err=%v", output, err)
				}
				if _, err := os.Stat(note.Path); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("document survived: %v", err)
				}
				if _, err := os.Stat(survivor.Path); err != nil {
					t.Fatal("another document was deleted")
				}
				for _, table := range []string{"habits", "persons", "bookmarks", "inbox", "pages", "journals", "note_tags"} {
					var count int
					if err := fixture.db().QueryRow("SELECT count(*) FROM "+table+" WHERE note_id=?", note.ID.String()).Scan(&count); err != nil || count != 0 {
						t.Fatalf("%s count=%d err=%v", table, count, err)
					}
				}
				var groups int
				if err := fixture.db().QueryRow("SELECT count(*) FROM person_groups WHERE person_id=?", note.ID.String()).Scan(&groups); err != nil || groups != 0 {
					t.Fatalf("groups=%d err=%v", groups, err)
				}
				list, err := fixture.run([]string{"list"}, "")
				if err != nil || strings.Contains(list, note.ID.String()) || !strings.Contains(list, survivor.ID.String()) {
					t.Fatalf("list=%s err=%v", list, err)
				}
			}
			// Reindex cannot resurrect a note whose source document has been deleted.
			if output, err := fixture.run([]string{"reindex"}, ""); err != nil || output != "Indexed 1 notes.\n" {
				t.Fatalf("reindex=%q err=%v", output, err)
			}
			if spec.kind == "journal" {
				if _, err := fixture.run([]string{"journal", "--date", "2026-10-04"}, ""); err != nil {
					t.Fatalf("deleted journal still reserves date: %v", err)
				}
			}
		})
	}
}

func TestDeleteCommandValidationAndUnindexedDocument(t *testing.T) {
	fixture := newCommandFixture(t)
	note := fixture.create("page", []string{"page", "Keep me"}, "", nil)
	for _, args := range [][]string{
		{"delete"}, {"delete", "invalid"}, {"delete", uuid.Nil().String()}, {"delete", note.ID.String(), "extra"},
	} {
		if _, err := fixture.run(args, ""); err == nil {
			t.Fatalf("accepted %v", args)
		}
		if _, err := os.Stat(note.Path); err != nil {
			t.Fatal("invalid command removed source")
		}
		fixture.requireSingleNote()
	}
	// A damaged document can be deleted without decoding it or reading the index.
	if err := os.WriteFile(note.Path, []byte("broken document"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.db().Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.db().Exec("DELETE FROM notes WHERE id=?", note.ID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.run([]string{"delete", note.ID.String()}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(note.Path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("unindexed document survived")
	}
}

func TestDeleteFindsNestedNotesAndRejectsDuplicateIdentity(t *testing.T) {
	fixture := newCommandFixture(t)
	note := fixture.create("page", []string{"page", "Nested"}, "", nil)
	nested := filepath.Join(filepath.Dir(note.Path), "subdirectory", filepath.Base(note.Path))
	if err := os.MkdirAll(filepath.Dir(nested), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nested, note.Source, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.run([]string{"delete", note.ID.String()}, ""); err == nil {
		t.Fatal("duplicate identity must not be deleted ambiguously")
	}
	for _, path := range []string{note.Path, nested} {
		if _, err := os.Stat(path); err != nil {
			t.Fatal("duplicate deletion modified source files")
		}
	}
	if err := os.Remove(note.Path); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.run([]string{"delete", note.ID.String()}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(nested); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("nested note survived")
	}
	if output, err := fixture.run([]string{"reindex"}, ""); err != nil || output != "Indexed 0 notes.\n" {
		t.Fatalf("nested note restored: %q %v", output, err)
	}
}

func TestDeleteCleansStaleIndexWhenDocumentIsAlreadyAbsent(t *testing.T) {
	fixture := newCommandFixture(t)
	note := fixture.create("page", []string{"page", "Absent"}, "", nil)
	if err := os.Remove(note.Path); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.run([]string{"delete", note.ID.String()}, ""); err != nil {
		t.Fatal(err)
	}
	if output, err := fixture.run([]string{"list"}, ""); err != nil || output != "No notes found.\n" {
		t.Fatalf("stale note survived: %q %v", output, err)
	}
}
