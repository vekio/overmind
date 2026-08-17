package createpage

import (
	"errors"
	"fmt"
)

// ErrPageAlreadyExists identifies an expected conflict with an existing page.
var (
	ErrPageAlreadyExists      = errors.New("page already exists")
	ErrPageCreationIncomplete = errors.New("page creation incomplete")
)

// PageAlreadyExistsError reports the logical path that prevented creation.
// It translates a storage conflict into terminology understood by every input
// adapter.
type PageAlreadyExistsError struct {
	Path string
}

// Error describes the creation conflict.
func (err *PageAlreadyExistsError) Error() string {
	return fmt.Sprintf("page already exists at %q", err.Path)
}

// Is allows callers to classify the error without depending on its details.
func (err *PageAlreadyExistsError) Is(target error) bool {
	return target == ErrPageAlreadyExists
}

// PageCreationIncompleteError reports that the page was stored but could not
// be indexed, and removing the stored page also failed. The document may
// therefore exist without a corresponding index entry.
type PageCreationIncompleteError struct {
	Path       string
	indexErr   error
	cleanupErr error
}

// NewPageCreationIncompleteError recreates the semantic error when a remote
// adapter reports the same uncertain state without exposing server causes.
func NewPageCreationIncompleteError(path string) *PageCreationIncompleteError {
	return &PageCreationIncompleteError{Path: path}
}

// Error describes the uncertain state left by a failed compensation.
func (err *PageCreationIncompleteError) Error() string {
	return fmt.Sprintf("page creation is incomplete at %q: document may exist without an index entry", err.Path)
}

// Is allows callers to classify the incomplete creation.
func (err *PageCreationIncompleteError) Is(target error) bool {
	return target == ErrPageCreationIncomplete
}

// Unwrap preserves both operational causes for logs and diagnostics.
func (err *PageCreationIncompleteError) Unwrap() []error {
	causes := make([]error, 0, 2)
	if err.indexErr != nil {
		causes = append(causes, err.indexErr)
	}
	if err.cleanupErr != nil {
		causes = append(causes, err.cleanupErr)
	}
	return causes
}
