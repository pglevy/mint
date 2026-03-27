---
name: mint-cli
description: Guide for querying Sailwind design tokens using the Mint CLI. Use when looking up colors, typography, spacing, or semantic tokens for UI components in this project.
---

# Mint Design Token CLI

Mint is a CLI tool that queries Sailwind design tokens. Use it to look up colors, typography, spacing, and semantic tokens when building UI components.

## How to invoke

All commands run from the `mint/` workspace folder:

```bash
npx tsx src/cli.ts <command> [options]
```

## Commands

### List tokens

Show all tokens grouped by category, or filter by category:

```bash
npx tsx src/cli.ts list
npx tsx src/cli.ts list --category color
npx tsx src/cli.ts list --category typography
npx tsx src/cli.ts list --category spacing
```

### Query tokens as JSON

Get structured DTCG JSON output, with optional category and family filters:

```bash
npx tsx src/cli.ts tokens
npx tsx src/cli.ts tokens --category color --family red
npx tsx src/cli.ts tokens --category spacing
```

### Export tokens

Export the full token set or a filtered subset as DTCG JSON:

```bash
npx tsx src/cli.ts export
npx tsx src/cli.ts export --category color --out color-tokens.json
```

### MCP schema

Output the MCP tool definition schema:

```bash
npx tsx src/cli.ts mcp-schema
```

## Token categories

- **color** — Aurora color palette (red, orange, yellow, green, teal, sky, blue, purple, pink, gray) plus semantic colors (destructive, positive, accent, secondary, standard) and aliases (black)
- **typography** — Font family, font weight, and custom text sizes
- **spacing** — Custom spacing values, border radius, and SAIL margin size mappings

## When to use

- When you need the exact hex value for a Sailwind color (e.g., "what's the hex for red-500?")
- When building a component and need to know available spacing or typography tokens
- When you need to understand semantic color mappings (e.g., which color is "destructive"?)
- When exporting tokens for consumption by other tools

## JSON output

Add `--json` to any command for machine-readable output. The `tokens` and `export` commands always output JSON.
