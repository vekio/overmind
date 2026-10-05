package person

import (
	"context"
	"fmt"

	"github.com/vekio/overmind/internal/ports"
)

// ListGroupsQuery requests all group names currently assigned to a person.
type ListGroupsQuery struct{}

// ListGroupsResult contains normalized group names in alphabetical order.
type ListGroupsResult struct {
	Groups []string
}

// ListGroupsHandler reads existing memberships from the derived index.
type ListGroupsHandler struct {
	index ports.Index
}

// NewListGroupsHandler binds the group query to the note index.
func NewListGroupsHandler(index ports.Index) *ListGroupsHandler {
	return &ListGroupsHandler{index: index}
}

// Handle returns used group names; unused stored definitions are excluded by the index.
func (handler *ListGroupsHandler) Handle(ctx context.Context, _ ListGroupsQuery) (ListGroupsResult, error) {
	if err := ctx.Err(); err != nil {
		return ListGroupsResult{}, err
	}
	if handler.index == nil {
		return ListGroupsResult{}, fmt.Errorf("note index is not configured")
	}
	groups, err := handler.index.FindGroups(ctx)
	if err != nil {
		return ListGroupsResult{}, fmt.Errorf("list groups: %w", err)
	}
	return ListGroupsResult{Groups: groups}, nil
}
