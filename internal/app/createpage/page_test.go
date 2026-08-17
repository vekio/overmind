package createpage

import (
	"errors"
	"strings"
	"testing"

	"git.casta.me/alberto/overmind/internal/domain"
)

func TestNewPageFromCommandCreatesValidatedDomainPage(t *testing.T) {
	ids := &idGeneratorStub{id: "page-id"}
	handler := &CreatePageHandler{idGenerator: ids, clock: clockStub{now: testCreatedAt}}

	page, err := handler.newPageFromCommand(CreatePageCommand{
		Title: " First page ",
		Area:  "Knowledge/Go",
		Tags:  []string{"Go", "Diseño de dominio"},
	})
	if err != nil {
		t.Fatalf("newPageFromCommand() error = %v", err)
	}
	if page.ID().String() != "page-id" || page.Title().String() != "First page" ||
		page.Title().Slug() != "first-page" || page.Area().String() != "knowledge/go" ||
		!page.CreatedAt().Equal(testCreatedAt) {
		t.Fatalf("page = %+v", page)
	}
	if got := page.Tags().Strings(); len(got) != 2 || got[0] != "go" || got[1] != "diseno-de-dominio" {
		t.Fatalf("tags = %v", got)
	}
}

func TestNewPageFromCommandValidatesInputBeforeGeneratingID(t *testing.T) {
	tests := map[string]struct {
		command CreatePageCommand
		cause   error
		context string
	}{
		"title":         {command: CreatePageCommand{}, cause: domain.ErrInvalidTitle, context: "validate title"},
		"area":          {command: CreatePageCommand{Title: "Page", Area: "///"}, cause: domain.ErrInvalidArea, context: "validate area"},
		"tag":           {command: CreatePageCommand{Title: "Page", Tags: []string{"---"}}, cause: domain.ErrInvalidTag, context: "tag 1"},
		"duplicate tag": {command: CreatePageCommand{Title: "Page", Tags: []string{"Go", "go"}}, cause: domain.ErrDuplicateTag, context: "validate tags"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ids := &idGeneratorStub{id: "page-id"}
			handler := &CreatePageHandler{idGenerator: ids, clock: clockStub{now: testCreatedAt}}
			_, err := handler.newPageFromCommand(test.command)
			if !errors.Is(err, test.cause) || !strings.Contains(err.Error(), test.context) {
				t.Fatalf("newPageFromCommand() error = %v", err)
			}
			if ids.calls != 0 {
				t.Fatalf("Generate() calls = %d, want 0", ids.calls)
			}
		})
	}
}

func TestNewPageFromCommandReportsIDGeneratorFailure(t *testing.T) {
	idErr := errors.New("random source unavailable")
	handler := &CreatePageHandler{
		idGenerator: &idGeneratorStub{err: idErr},
		clock:       clockStub{now: testCreatedAt},
	}

	_, err := handler.newPageFromCommand(CreatePageCommand{Title: "Page"})
	if !errors.Is(err, idErr) || !strings.Contains(err.Error(), "generate document id") {
		t.Fatalf("newPageFromCommand() error = %v", err)
	}
}

func TestNewPageFromCommandRejectsInvalidGeneratedID(t *testing.T) {
	handler := &CreatePageHandler{
		idGenerator: &idGeneratorStub{id: "invalid id"},
		clock:       clockStub{now: testCreatedAt},
	}

	_, err := handler.newPageFromCommand(CreatePageCommand{Title: "Page"})
	if !errors.Is(err, domain.ErrInvalidDocumentID) || !strings.Contains(err.Error(), "invalid generated document id") {
		t.Fatalf("newPageFromCommand() error = %v", err)
	}
}

func TestOptionalArea(t *testing.T) {
	for name, test := range map[string]struct {
		value string
		want  string
		err   error
	}{
		"omitted": {value: "", want: ""},
		"blank":   {value: "   ", want: ""},
		"present": {value: "Knowledge/Go", want: "knowledge/go"},
		"invalid": {value: "///", err: domain.ErrInvalidArea},
	} {
		t.Run(name, func(t *testing.T) {
			area, err := optionalArea(test.value)
			if !errors.Is(err, test.err) || area.String() != test.want {
				t.Fatalf("optionalArea(%q) = (%q, %v)", test.value, area, err)
			}
		})
	}
}

func TestPageDocumentKey(t *testing.T) {
	id, _ := domain.NewDocumentID("page-id")
	title, _ := domain.NewTitle("My Page")
	area, _ := domain.NewArea("Knowledge/Go")

	rootPage := domain.NewPage(id, title, domain.Tags{}, domain.Area{}, testCreatedAt)
	nestedPage := domain.NewPage(id, title, domain.Tags{}, area, testCreatedAt)
	if got, want := pageDocumentKey(rootPage), "page/my-page.adoc"; got != want {
		t.Fatalf("root page key = %q, want %q", got, want)
	}
	if got, want := pageDocumentKey(nestedPage), "page/knowledge/go/my-page.adoc"; got != want {
		t.Fatalf("nested page key = %q, want %q", got, want)
	}
}

func TestIndexEntryForPageSeparatesCommonAndSpecificMetadata(t *testing.T) {
	id, _ := domain.NewDocumentID("page-id")
	title, _ := domain.NewTitle("My Page")
	area, _ := domain.NewArea("Knowledge")
	goTag, _ := domain.NewTag("Go")
	tags, _ := domain.NewTags(goTag)
	page := domain.NewPage(id, title, tags, area, testCreatedAt)

	entry := indexEntryForPage(page, "page/knowledge/my-page.adoc")
	if entry.ID != id || entry.Path != "page/knowledge/my-page.adoc" || entry.Kind != domain.DocumentKindPage ||
		entry.Title != "My Page" || !entry.CreatedAt.Equal(testCreatedAt) || len(entry.Tags) != 1 || entry.Tags[0] != "go" {
		t.Fatalf("entry = %+v", entry)
	}
	if len(entry.Attributes) != 1 || entry.Attributes["area"] != "knowledge" {
		t.Fatalf("attributes = %+v", entry.Attributes)
	}
}
