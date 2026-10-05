package cli

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
	"uuid"

	configlib "github.com/vekio/config"
	"github.com/vekio/overmind/internal/bootstrap"
	appconfig "github.com/vekio/overmind/internal/config"
	"github.com/vekio/overmind/internal/infra/index/sqlitedb"
	"github.com/vekio/overmind/pkg/asciidoc"
)

func TestOvermindRegistersCreationCommandsAndDropsRetiredNames(t *testing.T) {
	config, err := appconfig.NewFile()
	if err != nil {
		t.Fatal(err)
	}
	root := New(config, fixedClient(&otherCommandsClient{}))
	for _, name := range []string{"habit", "person", "bookmark", "inbox", "page", "journal", "list", "reindex", "delete", "setup"} {
		if root.Command(name) == nil {
			t.Fatalf("missing creation command %q", name)
		}
	}
	for _, name := range []string{"capture", "rebuild"} {
		if root.Command(name) != nil {
			t.Fatalf("retired command %q still registered", name)
		}
	}
}

// commandFixture runs the root command with the real local runtime.
// Entity-specific integration assertions live in the corresponding command tests.
type commandFixture struct {
	t       *testing.T
	config  *configlib.ConfigFile[appconfig.Settings]
	factory ClientFactory
	vault   string
	index   *sql.DB
}
type createdNote struct {
	ID     uuid.UUID
	Path   string
	Source []byte
}

func newCommandFixture(t *testing.T) *commandFixture {
	t.Helper()
	root := t.TempDir()
	config, err := appconfig.NewFile()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.SetPath(filepath.Join(root, "config.yml")); err != nil {
		t.Fatal(err)
	}
	vault := filepath.Join(root, "vault")
	if err := config.Create(appconfig.Settings{Mode: appconfig.ModeLocal, VaultPath: vault}); err != nil {
		t.Fatal(err)
	}
	runtime := bootstrap.New(config)
	t.Cleanup(func() {
		if err := runtime.Close(); err != nil {
			t.Error(err)
		}
	})
	factory := func(ctx context.Context) (Client, error) {
		application, err := runtime.Application(ctx)
		if err != nil {
			return nil, err
		}
		return NewLocalClient(application), nil
	}
	return &commandFixture{t: t, config: config, factory: factory, vault: vault}
}
func (f *commandFixture) run(args []string, input string) (string, error) {
	f.t.Helper()
	command := New(f.config, f.factory)
	var output bytes.Buffer
	command.Writer, command.ErrWriter, command.Reader = &output, io.Discard, strings.NewReader(input)
	err := command.Run(context.Background(), append([]string{"overmind"}, args...))
	return output.String(), err
}
func (f *commandFixture) db() *sql.DB {
	f.t.Helper()
	if f.index != nil {
		return f.index
	}
	db, err := sql.Open("sqlite", filepath.Join(f.vault, "index.db"))
	if err != nil {
		f.t.Fatal(err)
	}
	f.t.Cleanup(func() {
		if err := db.Close(); err != nil {
			f.t.Error(err)
		}
	})
	f.index = db
	return db
}
func (f *commandFixture) create(kind string, args []string, input string, wantTags []string) createdNote {
	f.t.Helper()
	output, err := f.run(args, input)
	if err != nil {
		f.t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[1], "ID: ") {
		f.t.Fatalf("output=%q", output)
	}
	id, err := uuid.Parse(strings.TrimPrefix(lines[1], "ID: "))
	if err != nil {
		f.t.Fatal(err)
	}
	path := filepath.Join(f.vault, "notes", id.String()+".adoc")
	source, err := os.ReadFile(path)
	if err != nil {
		f.t.Fatal(err)
	}
	parsed := asciidoc.Process(source)
	if parsed.HasErrors() {
		f.t.Fatalf("invalid document: %s", source)
	}
	attrs := parsed.Analysis.Header.Attributes
	docID, _ := attrs.Lookup("overmind-id")
	docKind, _ := attrs.Lookup("overmind-type")
	if docID != id.String() || docKind != kind {
		f.t.Fatalf("document identity/type=%s %s", docID, docKind)
	}
	createdText, _ := attrs.Lookup("overmind-created-at")
	updatedText, _ := attrs.Lookup("overmind-updated-at")
	docCreated, err := time.Parse(time.RFC3339Nano, createdText)
	if err != nil {
		f.t.Fatal(err)
	}
	docUpdated, err := time.Parse(time.RFC3339Nano, updatedText)
	if err != nil {
		f.t.Fatal(err)
	}
	var storedKind, storedPath, created, updated string
	if err := f.db().QueryRow("SELECT type,path,created_at,updated_at FROM notes WHERE id=?", id.String()).Scan(&storedKind, &storedPath, &created, &updated); err != nil {
		f.t.Fatal(err)
	}
	at, err := time.Parse(time.RFC3339Nano, created)
	if err != nil || storedKind != kind || storedPath != path || created != updated || !at.Equal(docCreated) || !at.Equal(docUpdated) {
		f.t.Fatalf("document/index metadata mismatch: %s %s %s %s", storedKind, storedPath, created, updated)
	}
	tags, err := sqlitedb.New(f.db()).TagsByNoteID(context.Background(), id.String())
	if err != nil || !reflect.DeepEqual(tags, wantTags) {
		f.t.Fatalf("indexed tags=%v err=%v", tags, err)
	}
	docTags, _ := attrs.Lookup("overmind-tags")
	if docTags != strings.Join(wantTags, ", ") {
		f.t.Fatalf("document tags=%q", docTags)
	}
	return createdNote{ID: id, Path: path, Source: source}
}
func (f *commandFixture) requireSingleNote() {
	f.t.Helper()
	var count int
	if err := f.db().QueryRow("SELECT count(*) FROM notes").Scan(&count); err != nil || count != 1 {
		f.t.Fatalf("unexpected note count=%d err=%v", count, err)
	}
	files, err := os.ReadDir(filepath.Join(f.vault, "notes"))
	if err != nil || len(files) != 1 {
		f.t.Fatalf("unexpected documents=%v err=%v", files, err)
	}
}
