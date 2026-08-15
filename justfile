binary_name := "overmind"
build_dir := "bin"
main_package := "./cmd/overmind"
development_config := "./config.yml"

# List available recipes
[group('help')]
default:
    @just --list

# Run all unit tests, examples, and saved fuzz regression cases
[group('quality')]
test:
    go test ./...

# Run all tests with Go's race detector
[group('quality')]
test-race:
    go test -race ./...

# Generate and report statement coverage for the AsciiDoc lexer
[group('quality')]
coverage:
    go test -coverprofile=coverage.out ./pkg/asciidoc/lexer
    go tool cover -func=coverage.out

# Actively fuzz exact source reconstruction
[group('quality')]
fuzz-scanner duration="10s":
    go test ./pkg/asciidoc/lexer -run '^$' -fuzz '^FuzzScannerReconstructsSource$' -fuzztime "{{ duration }}"

# Actively fuzz line classification
[group('quality')]
fuzz-match duration="10s":
    go test ./pkg/asciidoc/lexer -run '^$' -fuzz '^FuzzMatchLinePreservesRaw$' -fuzztime "{{ duration }}"

# Actively run every fuzz target sequentially
[group('quality')]
fuzz duration="10s":
    just fuzz-scanner "{{ duration }}"
    just fuzz-match "{{ duration }}"

# Run formatting checks, vet, and tests
[group('quality')]
check: fmt-check vet test

# Format Go code
[group('quality')]
fmt:
    gofmt -w cmd internal pkg

# Check that Go code is formatted
[group('quality')]
fmt-check:
    @files="$(gofmt -l cmd internal pkg)"; if [ -n "$files" ]; then printf '%s\n' "$files"; exit 1; fi

# Run go vet
[group('quality')]
vet:
    go vet ./...

# Regenerate the type-safe SQLite query layer
[group('generation')]
generate:
    go tool sqlc generate

# Create a new sequential SQLite migration
[group('database')]
migration name:
    go tool goose -dir internal/infrastructure/sqliteindex/migrations -s create {{ name }} sql

# Build the CLI binary into ./bin
[group('artifacts')]
build: check
    mkdir -p {{ build_dir }}
    go build -o {{ build_dir }}/{{ binary_name }} {{ main_package }}

# Install the CLI binary into GOPATH/bin or GOBIN
[group('artifacts')]
install: check
    go install {{ main_package }}

# Run the CLI; pass arguments after `--`
[group('development')]
run *args:
    go run {{ main_package }} --config {{ development_config }} {{ replace(args, "-- ", "") }}

# Remove build artifacts
[group('artifacts')]
clean:
    rm -rf {{ build_dir }}
