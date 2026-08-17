// Package rebuildindex implements rebuilding the document read model.
package rebuildindex

// RebuildIndexCommand requests a complete index rebuild.
type RebuildIndexCommand struct{}

// RebuildIndexResult reports how many managed documents were indexed.
type RebuildIndexResult struct {
	Documents int
}
