package domain_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/vekio/overmind/internal/domain"
)

func TestTagsRejectDuplicatesAndProtectItems(t *testing.T) {
	first, _ := domain.NewTag("Work Notes")
	duplicate, _ := domain.NewTag("work-notes")
	if _, err := domain.NewTags(first, duplicate); !errors.Is(err, domain.ErrDuplicateTag) {
		t.Fatalf("normalized duplicate tags = %v", err)
	}
	tags, err := domain.NewTags(first)
	if err != nil {
		t.Fatal(err)
	}
	items := tags.Items()
	items[0] = domain.Tag{}
	if !reflect.DeepEqual(tags.Strings(), []string{"work-notes"}) {
		t.Fatalf("tag collection changed through Items: %v", tags.Strings())
	}
}
