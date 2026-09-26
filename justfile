set default-list

build_dir := "bin"
cli_binary_name := "overmind"
cli_main_package := "./cmd/overmind"
cli_config := justfile_directory() + "/config.yml"

# Run all unit tests, examples, and saved fuzz regression cases
[group('tests')]
test:
    go test ./...

# Run all tests with Go's race detector
[group('tests')]
test-race:
    go test -race ./...

# Report statement coverage for all project packages
[group('tests')]
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Run the AsciiDoc package tests
[group('asciidoc')]
[group('tests')]
test-asciidoc:
    go test ./pkg/asciidoc/...

# Actively fuzz exact source reconstruction
[group('asciidoc')]
[group('tests')]
fuzz-scanner duration="10s":
    go test ./pkg/asciidoc/lexer -run '^$' -fuzz '^FuzzScannerReconstructsSource$' -fuzztime "{{ duration }}"

# Actively fuzz line classification
[group('asciidoc')]
[group('tests')]
fuzz-match duration="10s":
    go test ./pkg/asciidoc/lexer -run '^$' -fuzz '^FuzzMatchLinePreservesRaw$' -fuzztime "{{ duration }}"

# Actively fuzz parser source bounds
[group('asciidoc')]
[group('tests')]
fuzz-parser duration="10s":
    go test ./pkg/asciidoc/parser -run '^$' -fuzz '^FuzzParserAlwaysReturnsCompleteSourceBounds$' -fuzztime "{{ duration }}"

# Actively run every fuzz target sequentially
[group('asciidoc')]
[group('tests')]
fuzz duration="10s":
    just fuzz-scanner "{{ duration }}"
    just fuzz-match "{{ duration }}"
    just fuzz-parser "{{ duration }}"

# Check that go.mod and go.sum are tidy without changing them
[group('quality')]
mod-tidy-check:
    go mod tidy -diff

# Run all repository quality checks
[group('quality')]
check: sqlc fmt-check mod-tidy-check vet test

# Format Go code
[group('quality')]
fmt:
    go fmt ./...

# Check that Go code is formatted
[group('quality')]
fmt-check:
    @files="$(gofmt -l $(rg --files -g '*.go'))"; if [ -n "$files" ]; then printf '%s\n' "$files"; exit 1; fi

# Run Go's static analysis
[group('quality')]
vet:
    go vet ./...

# Generate the type-safe SQLite access layer
[group('generation')]
sqlc:
    sqlc generate

# Create a sequential SQLite migration: just migration add_something
[group('database')]
migration name:
    goose -dir internal/infra/sqliteindex/migrations -s create "{{ name }}" sql

# Run checks and compile the CLI binary
[group('artifacts')]
build: check
    mkdir -p {{ build_dir }}
    go build -o {{ build_dir }}/{{ cli_binary_name }} {{ cli_main_package }}

# Install the CLI binary into GOPATH/bin or GOBIN
[group('artifacts')]
install: check
    go install {{ cli_main_package }}

# Run the Overmind CLI and forward its arguments
[group('development')]
[positional-arguments]
run *args:
    OVERMIND_CONFIG_FILE="{{ cli_config }}" go run {{ cli_main_package }} "$@"

# Remove build artifacts
[group('artifacts')]
clean:
    rm -rf "{{ build_dir }}"
