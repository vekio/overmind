package index

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vekio/overmind/internal/ports"
)

// writeTransaction lets ordinary upserts own a transaction, while rebuild
// upserts borrow the single outer transaction and cannot commit it early.
type writeTransaction struct {
	*sql.Tx
	owned bool
}

// Commit finishes an owned transaction; borrowed rebuild transactions remain open.
func (transaction *writeTransaction) Commit() error {
	if !transaction.owned {
		return nil
	}
	return transaction.Tx.Commit()
}

// Rollback discards an owned transaction; borrowed rebuild transactions remain open.
func (transaction *writeTransaction) Rollback() error {
	if !transaction.owned {
		return nil
	}
	return transaction.Tx.Rollback()
}
func (index *Index) beginWrite(ctx context.Context) (*writeTransaction, error) {
	if index.transaction != nil {
		return &writeTransaction{Tx: index.transaction}, nil
	}
	transaction, err := index.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &writeTransaction{Tx: transaction, owned: true}, nil
}

// Rebuild replaces all projections within one transaction and rolls back on callback failure.
// The callback must use the supplied scoped index rather than the outer connection pool.
func (index *Index) Rebuild(ctx context.Context, populate func(ports.Index) error) error {
	if populate == nil {
		return fmt.Errorf("index population callback is required")
	}
	transaction, err := index.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	// Clear children explicitly, including orphaned projections left by external edits.
	queries := index.queries.WithTx(transaction)
	for _, clear := range []func(context.Context) error{
		queries.ClearAllNoteTags, queries.ClearAllPersonGroups,
		queries.ClearHabits, queries.ClearPersons, queries.ClearBookmarks,
		queries.ClearInbox, queries.ClearPages, queries.ClearJournals,
		queries.ClearNotes, queries.ClearTags, queries.ClearGroups,
	} {
		if err := clear(ctx); err != nil {
			return fmt.Errorf("clear note projections: %w", err)
		}
	}

	scoped := &Index{db: index.db, queries: queries, transaction: transaction}
	if err := populate(scoped); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return transaction.Commit()
}
