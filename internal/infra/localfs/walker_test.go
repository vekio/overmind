package localfs_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"git.casta.me/alberto/overmind/internal/infra/localfs"
)

func TestWalkerVisitsOnlyAsciiDocFilesAndHonorsCancellation(t *testing.T) {
	root := t.TempDir()
	for path, content := range map[string]string{
		"a.adoc":        "one",
		"nested/b.ADOC": "two",
		"nested/c.txt":  "ignored",
	} {
		fullPath := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var visited []string
	err := localfs.NewWalker(root).Walk(context.Background(), func(path string, content []byte) error {
		visited = append(visited, filepath.Base(path)+":"+string(content))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(visited)
	if len(visited) != 2 || visited[0] != "a.adoc:one" || visited[1] != "b.ADOC:two" {
		t.Fatalf("visited files = %v", visited)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := localfs.NewWalker(root).Walk(ctx, func(string, []byte) error {
		t.Fatal("walk visited after cancellation")
		return nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled walk = %v", err)
	}
}
