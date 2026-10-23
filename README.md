# Mint CLI

A command-line tool for querying and exporting
[Sailwind](https://github.com/pglevy/sailwind) design tokens. Tokens are fetched
directly from the published token source — no local Sailwind install needed.

## Prerequisites

- [Go](https://go.dev/) 1.24 or later

## Quick Start

```bash
git clone https://github.com/pglevy/mint.git
cd mint
go run . list --category color
```

## Installing as a Global Command

To use `mint` directly from anywhere:

```bash
go install github.com/pglevy/mint@latest
```

This places a `mint` executable in `$(go env GOPATH)/bin` (usually
`~/go/bin`). Make sure that's on your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"  # add to ~/.zshrc if not already there
```

Or install from a local clone, which is also how you pick up code changes:

```bash
go install .
```

Then you can run:

```bash
mint list --category color
mint export -o tokens.json
mint --help
```

## Agents

Add a note to your global agent file (e.g., CLAUDE.md) to use `mint` for checking design token preferences.

For example:

```
## Design Preferences

Unless specified otherwise, use `mint` cli tool to look up preferred values for color (including gradient), spacing, and typography. Start with `-h` flag to see what's available.
```

## Usage

```bash
# List all tokens
mint list

# Filter by category
mint list --category color
mint list --category typography
mint list --category spacing

# Query tokens as DTCG JSON
mint tokens
mint tokens --category color --family red

# Export full token set to stdout
mint export

# Export to file, filtered by category
mint export -o tokens.json --category spacing

# JSON output for piping
mint list --category color --json | jq '.[] | .path'
```

## Global Options

| Flag            | Description                     |
| --------------- | ------------------------------- |
| `-h, --help`    | Show help text                  |
| `-v, --version` | Show version                    |
| `--json`        | Machine-readable JSON output    |
| `-q, --quiet`   | Suppress informational messages |
| `--no-color`    | Disable color output            |
| `--debug`       | Show detailed error information |

## Commands

| Command  | Description                                                             |
| -------- | ----------------------------------------------------------------------- |
| `list`   | List tokens grouped by category with name, value, and type              |
| `tokens` | Query tokens with `--category` and `--family` filters, output DTCG JSON |
| `export` | Export DTCG JSON to stdout or `-o <file>`                               |

## Development

```bash
# Run tests
go test ./...

# Vet
go vet ./...

# Format (lists files that need formatting)
gofmt -l .

# Build a binary in the current directory
go build -o mint .

# Stamp a release version
go build -ldflags "-X github.com/pglevy/mint/internal/cli.Version=1.0.0" -o mint .
```

The tests need no network access: the commands run against a fixture token
file, and `internal/cli/testdata/golden/` holds expected output captured from
the original Deno implementation.

## Project Structure

```
mint/
├── go.mod
├── main.go                    # Entry point
├── main_test.go               # Builds and runs the real binary
└── internal/
    ├── cli/                   # Routing, global flags, error reporting
    │   ├── cli.go
    │   ├── list.go
    │   ├── tokens.go
    │   ├── export.go
    │   └── testdata/          # Fixture tokens + golden command output
    ├── args/                  # Argument parsing (same rules as Deno's parseArgs)
    ├── tokens/                # Token store, DTCG flattening, fetcher
    ├── jsonv/                 # JSON with JavaScript key order and formatting
    ├── jsnum/                 # JavaScript number formatting and parsing
    ├── output/                # Color, writer, value formatting
    ├── prompt/                # Interactive arrow-key select
    ├── help/                  # Help text
    └── errs/                  # User-facing errors, "did you mean" suggestions
```

## Inspo

<https://clig.dev/>
