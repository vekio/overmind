package index

import (
	"context"
)

// FindGroups returns alphabetical normalized names assigned to at least one person.
func (index *Index) FindGroups(ctx context.Context) ([]string, error) {
	return index.queries.FindGroups(ctx)
}
