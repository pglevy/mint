# Mint CLI

A CLI tool for querying [Sailwind](https://github.com/pglevy/sailwind) design tokens. Tokens are fetched directly from the published token source — no local Sailwind install needed.

## Installation

```bash
git clone https://github.com/pglevy/mint.git
cd mint
npm install
npm run build
npm link
```

This makes the `mint` command available globally in your terminal.

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
mint export --out tokens.json --category spacing

# Output MCP tool schema
mint mcp-schema
```

### Global Options

| Flag | Description |
|------|-------------|
| `--json` | Machine-readable JSON output |

## Commands

| Command | Description |
|---------|-------------|
| `list` | List tokens grouped by category with name, value, and type |
| `tokens` | Query tokens with `--category` and `--family` filters, output as DTCG JSON |
| `export` | Export DTCG JSON to stdout or `--out <file>` |
| `mcp-schema` | Output MCP tool definitions for agent integration |

## Project Structure

```
src/
  cli.ts          # Entry point
  types.ts        # Shared types
  store/
    token-store.ts  # Fetches remote tokens and provides query methods
  commands/
    list.ts
    tokens.ts
    export.ts
    mcp-schema.ts
```

## Unlinking

If you need to unlink later, run either or both of these commands, depending on which node you're using:

```
npm unlink -g mint-cli          # removes from whichever npm is active (nvm)
/opt/homebrew/bin/npm unlink -g mint-cli  # removes from homebrew's npm
```