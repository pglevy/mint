# Mint CLI

A proof-of-concept CLI tool that exposes Sailwind design tokens as a queryable, agent-friendly API. Mint parses Sailwind's CSS theme block and SAIL type definitions to extract color, typography, and spacing tokens, structures them in W3C DTCG-compatible JSON, and provides commands for listing, filtering, and exporting tokens.

## Setup

```bash
npm install
```

## Usage

```bash
# List all tokens grouped by category
npx tsx src/cli.ts list

# List only color tokens
npx tsx src/cli.ts list --category color

# Query tokens as DTCG JSON
npx tsx src/cli.ts tokens
npx tsx src/cli.ts tokens --category color --family red

# Export full token set to stdout
npx tsx src/cli.ts export

# Export to file, filtered by category
npx tsx src/cli.ts export --out tokens.json --category spacing

# Output MCP tool schema
npx tsx src/cli.ts mcp-schema
```

### Global Options

| Flag | Description |
|------|-------------|
| `--source <path>` | Override the Sailwind source directory |
| `--json` | Machine-readable JSON output (all commands) |

## Commands

| Command | Description |
|---------|-------------|
| `list` | List tokens grouped by category with name, value, and type |
| `tokens` | Query tokens with `--category` and `--family` filters, output as DTCG JSON |
| `export` | Export DTCG JSON to stdout or `--out <file>` |
| `mcp-schema` | Output MCP tool definitions for LLM agent integration |

## Testing

```bash
npm test
```

## Project Structure

```
src/
  cli.ts                  # CLI entry point (Commander.js)
  types.ts                # Shared interfaces and constants
  parsers/
    css-parser.ts         # Extracts tokens from Sailwind index.css
    sail-parser.ts        # Extracts SAIL enum types from types/sail.ts
  store/
    token-store.ts        # In-memory DTCG token tree with query methods
    semantic-mapper.ts    # Maps SAIL types to semantic token aliases
  commands/
    list.ts               # mint list
    tokens.ts             # mint tokens
    export.ts             # mint export
    mcp-schema.ts         # mint mcp-schema
```
