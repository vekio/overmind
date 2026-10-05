package inbox_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"uuid"

	"github.com/vekio/overmind/internal/app"
	appinbox "github.com/vekio/overmind/internal/app/inbox"
	"github.com/vekio/overmind/internal/domain/inbox"
)

type inboxTestRepository struct {
	saved *inbox.Inbox
	err   error
}

func (repo *inboxTestRepository) Save(_ context.Context, note *inbox.Inbox) error {
	if repo.err != nil {
		return repo.err
	}
	repo.saved = note
	return nil
}
func TestCreateInbox(t *testing.T) {
	repo := &inboxTestRepository{}
	handler := app.New(app.Dependencies{Inbox: repo, IDGenerator: testIDGenerator{}}).Commands.CreateInbox
	content := "Una idea\n:overmind-type: arbitrary body\n\nOtra línea.\n"
	result, err := handler.Handle(context.Background(), appinbox.CreateCommand{Content: content, Tags: []string{"Ideas"}})
	if err != nil {
		t.Fatal(err)
	}
	entity := result.Inbox
	if entity != repo.saved || entity.ID() == uuid.Nil() || entity.Content() != content || !reflect.DeepEqual(entity.Tags().Strings(), []string{"ideas"}) {
		t.Fatal("raw content or tags changed")
	}
	if entity.Metadata().IsZero() || !entity.Metadata().CreatedAt().Equal(entity.Metadata().UpdatedAt()) {
		t.Fatal("invalid initial metadata")
	}
}
func TestCreateInboxRejectsInvalidTagsWithoutSaving(t *testing.T) {
	for _, tags := range [][]string{{"!!!"}, {"One", "one"}} {
		repo := &inboxTestRepository{}
		handler := app.New(app.Dependencies{Inbox: repo, IDGenerator: testIDGenerator{}}).Commands.CreateInbox
		result, err := handler.Handle(context.Background(), appinbox.CreateCommand{Content: "idea", Tags: tags})
		if err == nil || result.Inbox != nil || repo.saved != nil {
			t.Fatal("invalid command saved")
		}
	}
}
func TestCreateInboxAllowsEmptyContentAndPropagatesFailures(t *testing.T) {
	repo := &inboxTestRepository{}
	handler := app.New(app.Dependencies{Inbox: repo, IDGenerator: testIDGenerator{}}).Commands.CreateInbox
	command := appinbox.CreateCommand{}
	result, err := handler.Handle(context.Background(), command)
	if err != nil || result.Inbox.Content() != "" || !result.Inbox.Tags().IsEmpty() {
		t.Fatal(err)
	}
	failure := errors.New("disk failure")
	repo.err = failure
	if result, err := handler.Handle(context.Background(), command); !errors.Is(err, failure) || result.Inbox != nil {
		t.Fatal("save failure not propagated")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := handler.Handle(ctx, command); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := app.New(app.Dependencies{}).Commands.CreateInbox.Handle(context.Background(), command); err == nil {
		t.Fatal("missing dependencies accepted")
	}
}

// testIDGenerator keeps ID generation independent of infrastructure.
type testIDGenerator struct{}

func (testIDGenerator) Generate() uuid.UUID { return uuid.New() }

func (repo *inboxTestRepository) Update(ctx context.Context, entity *inbox.Inbox) error {
	return repo.Save(ctx, entity)
}

func (repo *inboxTestRepository) ByID(_ context.Context, _ uuid.UUID) (*inbox.Inbox, error) {
	return repo.saved, repo.err
}
