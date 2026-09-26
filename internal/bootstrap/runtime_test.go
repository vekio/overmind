package bootstrap

import (
	"errors"
	"io"
	"slices"
	"testing"
)

type recordingCloser struct {
	name  string
	calls *[]string
	err   error
}

func (closer recordingCloser) Close() error {
	*closer.calls = append(*closer.calls, closer.name)
	return closer.err
}

func TestRuntimeClosesAllResourcesInReverseOrder(t *testing.T) {
	var calls []string
	firstErr := errors.New("first failed")
	secondErr := errors.New("second failed")
	runtime := &Runtime{closers: []io.Closer{
		recordingCloser{name: "first", calls: &calls, err: firstErr},
		recordingCloser{name: "second", calls: &calls, err: secondErr},
	}}

	err := runtime.Close()
	if !errors.Is(err, firstErr) || !errors.Is(err, secondErr) {
		t.Fatalf("Close() error = %v, want both resource errors", err)
	}
	if want := []string{"second", "first"}; !slices.Equal(calls, want) {
		t.Errorf("close order = %v, want %v", calls, want)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("second Close() = %v, want nil", err)
	}
	if len(calls) != 2 {
		t.Errorf("second Close() repeated resource close: %v", calls)
	}
}
