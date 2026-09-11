build_dir := "bin"
cli_binary_name := "cspeek"
cli_main_package := "./cmd/cspeek"
cli_config := justfile_directory() + "/config.yml"

# List available recipes
[group('help')]
default:
    @just --list

# Run all unit and integration tests
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

# Check that go.mod and go.sum are tidy without changing them
[group('quality')]
mod-tidy-check:
    go mod tidy -diff

# Run all repository quality checks
[group('quality')]
check: fmt-check mod-tidy-check vet test

# Format Go code
[group('quality')]
fmt:
    go fmt ./...
    go tool templ fmt .

# Check that Go code is formatted
[group('quality')]
fmt-check:
    @files="$(gofmt -l $(rg --files -g '*.go'))"; if [ -n "$files" ]; then printf '%s\n' "$files"; exit 1; fi
    go tool templ fmt -fail .

# Run Go's static analysis
[group('quality')]
vet:
    go vet ./...

# Remove build artifacts and reset the local SQLite database
[group('artifacts')]
clean:
    rm -rf "{{ build_dir }}"

# Generate code, run checks, and compile the API and CLI binaries
[group('artifacts')]
build: check
    mkdir -p {{ build_dir }}
    go build -o {{ build_dir }}/{{ cli_binary_name }} {{ cli_main_package }}

# Run the CSpeek CLI and forward its arguments
[group('development')]
run *args:
    CSPEEK_CONFIG_FILE="{{ cli_config }}" go run {{ cli_main_package }} {{ args }}
