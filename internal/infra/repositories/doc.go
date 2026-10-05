// Package repositories persists typed entities in the vault and refreshes their index projections.
// Document writes precede indexing; an index failure can therefore leave a saved document.
// Typed updates re-encode the managed document; raw editing preserves unrelated source formatting.
package repositories
