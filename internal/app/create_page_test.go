package app_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app"
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
	result, err := handler.Handle(context.Background(), app.CreatePageCommand{Title: " Project plan ", Area: "Work/Ideas", Tags: []string{"Important"}})
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
	for _, command := range []app.CreatePageCommand{{}, {Title: "!!!"}, {Title: "Page", Area: "Work//Ideas"}, {Title: "Page", Tags: []string{"!!!"}}, {Title: "Page", Tags: []string{"One", "one"}}} {
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
	command := app.CreatePageCommand{Title: "Page"}
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
