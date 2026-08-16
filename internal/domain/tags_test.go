package domain

import (
	"errors"
	"slices"
	"testing"
)

func TestNewTagsPreservesNormalizedInputOrder(t *testing.T) {
	goTag, _ := NewTag("Go")
	domainDesignTag, _ := NewTag("Diseño de dominio")
	tags, err := NewTags(goTag, domainDesignTag)
	if err != nil {
		t.Fatalf("NewTags() error = %v", err)
	}
	if got, want := tags.Strings(), []string{"go", "diseno-de-dominio"}; !slices.Equal(got, want) {
		t.Fatalf("Tags.Strings() = %v, want %v", got, want)
	}
	if tags.Len() != 2 || tags.IsEmpty() {
		t.Fatalf("Tags length = %d, empty = %t", tags.Len(), tags.IsEmpty())
	}
}

func TestNewTagsRejectsDuplicatesAfterNormalization(t *testing.T) {
	first, _ := NewTag("Go Lang")
	second, _ := NewTag("go-lang")
	if _, err := NewTags(first, second); !errors.Is(err, ErrDuplicateTag) {
		t.Fatalf("NewTags() error = %v, want %v", err, ErrDuplicateTag)
	}
}

func TestTagsItemsReturnsACopy(t *testing.T) {
	goTag, _ := NewTag("go")
	tags, _ := NewTags(goTag)
	items := tags.Items()
	items[0] = Tag{}

	tag, _ := NewTag("go")
	if !tags.Contains(tag) {
		t.Fatal("mutating Items() changed the collection")
	}
}

func TestNewTagsRejectsZeroTag(t *testing.T) {
	if _, err := NewTags(Tag{}); !errors.Is(err, ErrInvalidTag) {
		t.Fatalf("NewTags() error = %v, want %v", err, ErrInvalidTag)
	}
}

func TestTagsZeroValueIsEmpty(t *testing.T) {
	var tags Tags
	if !tags.IsEmpty() || tags.Len() != 0 || len(tags.Items()) != 0 || len(tags.Strings()) != 0 {
		t.Fatalf("zero Tags = %+v", tags)
	}
}
