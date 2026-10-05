package bookmark_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	appbookmark "github.com/vekio/overmind/internal/app/bookmark"
	"github.com/vekio/overmind/internal/domain/bookmarks"
)

type bookmarkTestRepository struct {
	saved *bookmarks.Bookmark
	err   error
}

func (repo *bookmarkTestRepository) Save(_ context.Context, bookmark *bookmarks.Bookmark) error {
	if repo.err != nil {
		return repo.err
	}
	repo.saved = bookmark
	return nil
}
func TestCreateBookmark(t *testing.T) {
	repo := &bookmarkTestRepository{}
	handler := app.New(app.Dependencies{Bookmarks: repo, IDGenerator: testIDGenerator{}}).Commands.CreateBookmark
	result, err := handler.Handle(context.Background(), appbookmark.CreateCommand{URL: " https://example.com/a?q=1&lang=es ", Tags: []string{"Lectura"}})
	if err != nil {
		t.Fatal(err)
	}
	entity := result.Bookmark
	if entity != repo.saved || entity.ID() == uuid.Nil() || entity.URL().String() != "https://example.com/a?q=1&lang=es" || !reflect.DeepEqual(entity.Tags().Strings(), []string{"lectura"}) {
		t.Fatal("raw input was not validated and saved")
	}
	if entity.Metadata().IsZero() || !entity.Metadata().CreatedAt().Equal(entity.Metadata().UpdatedAt()) {
		t.Fatal("invalid initial metadata")
	}
}
func TestCreateBookmarkRejectsInvalidInputWithoutSaving(t *testing.T) {
	for _, command := range []appbookmark.CreateCommand{
		{}, {URL: "file:///tmp/note"}, {URL: "https://example.com", Tags: []string{"!!!"}}, {URL: "https://example.com", Tags: []string{"One", "one"}},
	} {
		repo := &bookmarkTestRepository{}
		handler := app.New(app.Dependencies{Bookmarks: repo, IDGenerator: testIDGenerator{}}).Commands.CreateBookmark
		result, err := handler.Handle(context.Background(), command)
		if err == nil || result.Bookmark != nil || repo.saved != nil {
			t.Fatal("invalid command saved")
		}
	}
}
func TestCreateBookmarkOptionalTagsAndFailures(t *testing.T) {
	repo := &bookmarkTestRepository{}
	handler := app.New(app.Dependencies{Bookmarks: repo, IDGenerator: testIDGenerator{}}).Commands.CreateBookmark
	command := appbookmark.CreateCommand{URL: "https://example.com"}
	result, err := handler.Handle(context.Background(), command)
	if err != nil || !result.Bookmark.Tags().IsEmpty() {
		t.Fatal(err)
	}
	failure := errors.New("disk failure")
	repo.err = failure
	if result, err := handler.Handle(context.Background(), command); !errors.Is(err, failure) || result.Bookmark != nil {
		t.Fatal("save failure not propagated")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := handler.Handle(ctx, command); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := app.New(app.Dependencies{}).Commands.CreateBookmark.Handle(context.Background(), command); err == nil {
		t.Fatal("missing dependencies accepted")
	}
}

// testIDGenerator keeps ID generation independent of infrastructure.
type testIDGenerator struct{}

func (testIDGenerator) Generate() uuid.UUID { return uuid.New() }

func (repo *bookmarkTestRepository) Update(ctx context.Context, entity *bookmarks.Bookmark) error {
	return repo.Save(ctx, entity)
}

func (repo *bookmarkTestRepository) ByID(_ context.Context, _ uuid.UUID) (*bookmarks.Bookmark, error) {
	return repo.saved, repo.err
}
