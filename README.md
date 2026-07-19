# Mint CLI

A CLI tool for querying and exporting
[Sailwind](https://github.com/pglevy/sailwind) design tokens. Tokens are fetched
directly from the published token source — no local Sailwind install needed.

## Prerequisites

- [Deno](https://deno.com/) 2.x or later

## Quick Start

```bash
git clone https://github.com/pglevy/mint.git
cd mint
deno task start list
```

No install step required — Deno runs TypeScript directly.

## Installing as a Global Command

To use `mint` directly from anywhere (instead of `deno task start`):

```bash
deno install -f --allow-net=raw.githubusercontent.com --allow-env --name mint main.ts
```

This places a `mint` executable in `~/.deno/bin/`. Make sure that's on your
PATH:

```bash
export PATH="$HOME/.deno/bin:$PATH"  # add to ~/.zshrc if not already there
```

Then you can run:

```bash
mint list --category color
mint export -o tokens.json
mint --help
```

To update after code changes, re-run the install command with `-f` (force).

## Usage

```bash
# List all tokens
deno task start list

# Filter by category
deno task start list --category color
deno task start list --category typography
deno task start list --category spacing

# Query tokens as DTCG JSON
deno task start tokens
deno task start tokens --category color --family red

# Export full token set to stdout
deno task start export

# Export to file, filtered by category
deno task start export -o tokens.json --category spacing

# JSON output for piping
deno task start list --category color --json | jq '.[] | .path'
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

| Command      | Description                                                             |
| ------------ | ----------------------------------------------------------------------- |
| `list`       | List tokens grouped by category with name, value, and type              |
| `tokens`     | Query tokens with `--category` and `--family` filters, output DTCG JSON |
| `export`     | Export DTCG JSON to stdout or `-o <file>`                               |

## Development

```bash
# Run tests
deno task test

# Lint
deno task lint

# Format
deno task fmt

# Type check
deno task check
```

## Project Structure

```
mint/
├── deno.json              # Config, imports, tasks
├── deno.lock
├── main.ts                # Entry point
├── src/
│   ├── types.ts           # Shared types (DTCGToken, DTCGGroup, TokenCategory)
│   ├── errors.ts          # Error classes, suggestCommand()
│   ├── store/
│   │   ├── token-store.ts # TokenStore class
│   │   └── fetcher.ts     # Network fetch isolation
│   ├── commands/
│   │   ├── list.ts
│   │   ├── tokens.ts
│   │   └── export.ts
│   ├── output/
│   │   ├── format-value.ts
│   │   ├── color.ts
│   │   └── writer.ts
│   ├── help/
│   │   └── help.ts
│   └── prompt/
│       └── select.ts
├── tests/
│   ├── store/
│   │   └── token-store.test.ts
│   ├── output/
│   │   └── format-value.test.ts
│   ├── commands/
│   │   └── integration.test.ts
│   └── property/
│       └── round-trip.test.ts
└── gitignore/             # Local workspace files (not in version control)
```

## Repository

<https://github.com/pglevy/mint>
