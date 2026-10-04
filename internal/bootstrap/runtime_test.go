package bootstrap

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/vekio/overmind/internal/app"
	appconfig "github.com/vekio/overmind/internal/config"
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

func TestRuntimeLoadsConfigurationLazilyAndReusesApplication(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	configFile, err := appconfig.NewFile()
	if err != nil {
		t.Fatal(err)
	}
	if err := configFile.SetPath(filepath.Join(root, "config.yml")); err != nil {
		t.Fatal(err)
	}
	runtime := New(configFile)
	if _, err := runtime.Application(ctx); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("application before setup = %v", err)
	}
	if err := configFile.Create(appconfig.Settings{Mode: appconfig.ModeLocal, VaultPath: filepath.Join(root, "vault")}); err != nil {
		t.Fatal(err)
	}
	first, err := runtime.Application(ctx)
	if err != nil {
		t.Fatal(err)
	}
	second, err := runtime.Application(ctx)
	if err != nil || first != second {
		t.Fatalf("runtime returned another application: %p, %p, %v", first, second, err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeConcurrentQueriesShareOneApplication(t *testing.T) {
	root := t.TempDir()
	configFile, err := appconfig.NewFile()
	if err != nil {
		t.Fatal(err)
	}
	if err := configFile.SetPath(filepath.Join(root, "config.yml")); err != nil {
		t.Fatal(err)
	}
	if err := configFile.Create(appconfig.Settings{Mode: appconfig.ModeLocal, VaultPath: filepath.Join(root, "vault")}); err != nil {
		t.Fatal(err)
	}
	runtime := New(configFile)
	t.Cleanup(func() { _ = runtime.Close() })
	type response struct {
		application *app.Application
		err         error
	}
	results := make(chan response, 8)
	for range 8 {
		go func() {
			application, err := runtime.Application(context.Background())
			results <- response{application, err}
		}()
	}
	var first *app.Application
	for range 8 {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		if first == nil {
			first = result.application
		}
		if result.application != first {
			t.Fatal("concurrent queries initialized multiple application instances")
		}
	}
}
