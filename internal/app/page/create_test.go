package page_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/app/page"
	"github.com/vekio/overmind/internal/domain/pages"
)

type pageTestRepository struct {
	saved *pages.Page
	err   error
}

func (repo *pageTestRepository) Save(_ context.Context, entity *pages.Page) error {
	if repo.err != nil {
		return repo.err
	}
	repo.saved = entity
	return nil
}
func TestCreatePage(t *testing.T) {
	repo := &pageTestRepository{}
	handler := app.New(app.Dependencies{Pages: repo, IDGenerator: testIDGenerator{}}).Commands.CreatePage
	result, err := handler.Handle(context.Background(), page.CreateCommand{Title: " Project plan ", Area: "Work/Ideas", Tags: []string{"Important"}})
	if err != nil {
		t.Fatal(err)
	}
	entity := result.Page
	if entity != repo.saved || entity.ID() == uuid.Nil() || entity.Title().String() != "Project plan" || entity.Area().String() != "work/ideas" || !reflect.DeepEqual(entity.Tags().Strings(), []string{"important"}) {
		t.Fatal("raw input was not validated and saved")
	}
	if entity.Metadata().IsZero() || !entity.Metadata().CreatedAt().Equal(entity.Metadata().UpdatedAt()) {
		t.Fatal("invalid initial metadata")
	}
}
func TestCreatePageRejectsInvalidInputsWithoutSaving(t *testing.T) {
	for _, command := range []page.CreateCommand{{}, {Title: "!!!"}, {Title: "Page", Area: "Work//Ideas"}, {Title: "Page", Tags: []string{"!!!"}}, {Title: "Page", Tags: []string{"One", "one"}}} {
		repo := &pageTestRepository{}
		handler := app.New(app.Dependencies{Pages: repo, IDGenerator: testIDGenerator{}}).Commands.CreatePage
		result, err := handler.Handle(context.Background(), command)
		if err == nil || result.Page != nil || repo.saved != nil {
			t.Fatal("invalid command saved")
		}
	}
}
func TestCreatePageOptionalAreaAndTagsAndFailures(t *testing.T) {
	repo := &pageTestRepository{}
	handler := app.New(app.Dependencies{Pages: repo, IDGenerator: testIDGenerator{}}).Commands.CreatePage
	command := page.CreateCommand{Title: "Page"}
	result, err := handler.Handle(context.Background(), command)
	if err != nil || !result.Page.Area().IsZero() || !result.Page.Tags().IsEmpty() {
		t.Fatal(err)
	}
	failure := errors.New("disk failure")
	repo.err = failure
	if result, err := handler.Handle(context.Background(), command); !errors.Is(err, failure) || result.Page != nil {
		t.Fatal("save failure not propagated")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := handler.Handle(ctx, command); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := app.New(app.Dependencies{}).Commands.CreatePage.Handle(context.Background(), command); err == nil {
		t.Fatal("missing dependencies accepted")
	}
}

// testIDGenerator keeps ID generation independent of infrastructure.
type testIDGenerator struct{}

func (testIDGenerator) Generate() uuid.UUID { return uuid.New() }

func (repo *pageTestRepository) Update(ctx context.Context, entity *pages.Page) error {
	return repo.Save(ctx, entity)
}

func (repo *pageTestRepository) ByID(_ context.Context, _ uuid.UUID) (*pages.Page, error) {
	return repo.saved, repo.err
}
