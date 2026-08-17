package createpage

import (
	"errors"
	"strings"
	"testing"
)

func TestPageAlreadyExistsError(t *testing.T) {
	err := &PageAlreadyExistsError{Path: "page/page.adoc"}
	if !errors.Is(err, ErrPageAlreadyExists) || err.Error() != `page already exists at "page/page.adoc"` {
		t.Fatalf("error = %v", err)
	}
}

func TestPageCreationIncompleteErrorPreservesCauses(t *testing.T) {
	indexErr := errors.New("index unavailable")
	cleanupErr := errors.New("delete unavailable")
	err := &PageCreationIncompleteError{
		Path:       "page/page.adoc",
		indexErr:   indexErr,
		cleanupErr: cleanupErr,
	}

	if !errors.Is(err, ErrPageCreationIncomplete) || !errors.Is(err, indexErr) || !errors.Is(err, cleanupErr) {
		t.Fatalf("error tree = %v", err)
	}
	if !strings.Contains(err.Error(), `page creation is incomplete at "page/page.adoc"`) {
		t.Fatalf("error = %v", err)
	}
}

func TestNewPageCreationIncompleteErrorHasNoOperationalCauses(t *testing.T) {
	err := NewPageCreationIncompleteError("page/page.adoc")
	if !errors.Is(err, ErrPageCreationIncomplete) || len(err.Unwrap()) != 0 {
		t.Fatalf("error = %v, causes = %v", err, err.Unwrap())
	}
}
