package person_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/vekio/overmind/internal/app"
	"github.com/vekio/overmind/internal/app/person"
	"github.com/vekio/overmind/internal/ports"
)

type groupFinder struct {
	ports.Index
	calls int
	ctx   context.Context
	err   error
}

func (finder *groupFinder) FindGroups(ctx context.Context) ([]string, error) {
	finder.calls++
	finder.ctx = ctx
	return []string{"friends", "work"}, finder.err
}

func TestListGroupsThroughApplication(t *testing.T) {
	finder := &groupFinder{}
	application := app.New(app.Dependencies{Index: finder})
	ctx := context.WithValue(context.Background(), struct{}{}, "groups")
	result, err := application.Queries.ListGroups.Handle(ctx, person.ListGroupsQuery{})
	if err != nil || finder.ctx != ctx || !reflect.DeepEqual(result.Groups, []string{"friends", "work"}) {
		t.Fatalf("result=%+v, error=%v", result, err)
	}
	failure := errors.New("unavailable")
	finder.err = failure
	if _, err := application.Queries.ListGroups.Handle(ctx, person.ListGroupsQuery{}); !errors.Is(err, failure) {
		t.Fatal("lookup error was not preserved")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := application.Queries.ListGroups.Handle(cancelled, person.ListGroupsQuery{}); !errors.Is(err, context.Canceled) || finder.calls != 2 {
		t.Fatal("cancelled lookup reached finder")
	}
	if _, err := person.NewListGroupsHandler(nil).Handle(ctx, person.ListGroupsQuery{}); err == nil {
		t.Fatal("missing finder was not reported")
	}
}
