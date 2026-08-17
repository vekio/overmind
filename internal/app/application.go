// Package app exposes the application's use-case handlers to input adapters.
package app

// Application groups the use-case handlers exposed by Overmind. It contains
// no transport or infrastructure logic; CLI and HTTP may consume the same
// instance.
type Application struct {
	Commands Commands
	Queries  Queries
}

// Commands contains the application's state-changing handlers.
type Commands struct {
	CreatePage     CreatePageHandler
	RebuildIndex   RebuildIndexHandler
	UpdateDocument UpdateDocumentHandler
}

// Queries contains the application's read-only handlers.
type Queries struct {
	GetDocument   GetDocumentHandler
	ListDocuments ListDocumentsHandler
}
