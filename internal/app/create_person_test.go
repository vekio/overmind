package app_test

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"git.casta.me/alberto/overmind/internal/app"
	"git.casta.me/alberto/overmind/internal/domain"
)

func TestPersonCreationEditingAndRebuild(t *testing.T) {
	ctx := context.Background()
	fixture := newAppFixture(t)
	name, _ := domain.NewTitle("Ana García")
	work, _ := domain.NewGroup("Trabajo")
	university, _ := domain.NewGroup("Universidad")
	groups, _ := domain.NewGroups(work, university)
	tag, _ := domain.NewTag("Amiga")
	tags, _ := domain.NewTags(tag)
	result, err := fixture.application.Commands.CreatePerson.Handle(ctx, app.CreatePersonCommand{Name: name, Groups: groups, Tags: tags})
	if err != nil {
		t.Fatal(err)
	}
	id := result.Person.Metadata().ID()
	original, err := fixture.application.Queries.OpenNote.Handle(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"= Ana García", ":overmind-type: person", ":overmind-groups: trabajo, universidad", "== Contact", "== Context", "== Notes"} {
		if !bytes.Contains(original, []byte(text)) {
			t.Fatalf("template missing %q: %s", text, original)
		}
	}
	listed, err := fixture.application.Queries.ListNotes.Handle(ctx)
	if err != nil {
		t.Fatal(err)
	}
	note := listedNote(t, listed.Notes, id)
	if note.Kind != domain.NoteKindPerson || note.Attributes["name"] != "Ana García" || note.Attributes["groups"] != "trabajo, universidad" || !reflect.DeepEqual(note.Tags, []string{"amiga"}) {
		t.Fatalf("indexed person = %+v", note)
	}
	invalid := bytes.Replace(original, []byte("trabajo, universidad"), []byte("Trabajo, trabajo"), 1)
	if _, err := fixture.application.Commands.UpdateNote.Handle(ctx, app.UpdateNoteCommand{ID: id, Original: original, Source: invalid}); err == nil {
		t.Fatal("edit accepted duplicate groups")
	}
	edited := bytes.Replace(original, []byte("= Ana García"), []byte("= Ana García López"), 1)
	edited = bytes.Replace(edited, []byte("trabajo, universidad"), []byte("universidad"), 1)
	updated, err := fixture.application.Commands.UpdateNote.Handle(ctx, app.UpdateNoteCommand{ID: id, Original: original, Source: edited})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(updated.Source, []byte("= Ana García López")) {
		t.Fatal("edit lost name")
	}
	if err := fixture.index.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	rebuilt, err := fixture.application.Commands.RebuildIndex.Handle(ctx)
	if err != nil || rebuilt.Count != 1 {
		t.Fatalf("rebuild = %+v, %v", rebuilt, err)
	}
	listed, err = fixture.application.Queries.ListNotes.Handle(ctx)
	if err != nil {
		t.Fatal(err)
	}
	note = listedNote(t, listed.Notes, id)
	if note.Attributes["name"] != "Ana García López" || note.Attributes["groups"] != "universidad" || !reflect.DeepEqual(note.Tags, []string{"amiga"}) {
		t.Fatalf("rebuilt person = %+v", note)
	}
}
