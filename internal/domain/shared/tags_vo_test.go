package shared_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/vekio/overmind/internal/domain/shared"
)

func TestTagsRejectDuplicatesAndProtectItems(t *testing.T) {
	first, _ := shared.NewTag("Work Notes")
	duplicate, _ := shared.NewTag("work-notes")
	if _, err := shared.NewTags(first, duplicate); !errors.Is(err, shared.ErrDuplicateTag) {
		t.Fatalf("normalized duplicate tags = %v", err)
	}
	if _, err := shared.NewTags(shared.Tag{}); !errors.Is(err, shared.ErrInvalidTag) {
		t.Fatalf("uninitialized tag = %v", err)
	}
	source := []shared.Tag{first}
	tags, err := shared.NewTags(source...)
	if err != nil {
		t.Fatal(err)
	}
	source[0] = shared.Tag{}
	items := tags.Items()
	items[0] = shared.Tag{}
	if !reflect.DeepEqual(tags.Strings(), []string{"work-notes"}) {
		t.Fatalf("tag collection changed through an external slice: %v", tags.Strings())
	}
}
