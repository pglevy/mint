#!/usr/bin/env node
import { Command } from 'commander';
import path from 'node:path';
import { parseCSSFile } from './parsers/css-parser.js';
import { parseSAILFile } from './parsers/sail-parser.js';
import { TokenStore } from './store/token-store.js';
import { listCommand } from './commands/list.js';
import { tokensCommand } from './commands/tokens.js';
import { exportCommand } from './commands/export.js';
import { mcpSchemaCommand } from './commands/mcp-schema.js';

/**
 * Default Sailwind source directory — relative to the mint/ project root.
 * Can be overridden with the --source flag.
 */
const DEFAULT_SOURCE = path.resolve(import.meta.dirname, '../../sailwind/src');

/**
 * Build a TokenStore from the Sailwind source files at the given path.
 * Handles errors with descriptive stderr messages and non-zero exit codes.
 * When jsonMode is true, also outputs `{ "error": "<message>" }` to stdout.
 */
function loadTokenStore(sourcePath: string, jsonMode: boolean): TokenStore {
  const cssPath = path.join(sourcePath, 'index.css');
  const sailPath = path.join(sourcePath, 'types/sail.ts');

  let cssResult;
  try {
    cssResult = parseCSSFile({ filePath: cssPath });
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    process.stderr.write(`${message}\n`);
    if (jsonMode) {
      process.stdout.write(JSON.stringify({ error: message }) + '\n');
    }
    process.exit(1);
  }

  const sailResult = parseSAILFile({ filePath: sailPath });

  const store = new TokenStore();
  store.load(cssResult, sailResult);
  return store;
}

const program = new Command();

program
  .name('mint')
  .description('Query Sailwind design tokens')
  .version('0.1.0')
  .option('--source <path>', 'Path to Sailwind source directory', DEFAULT_SOURCE)
  .option('--json', 'Output machine-readable JSON', false);

// ── list command ─────────────────────────────────────────────────────
program
  .command('list')
  .description('List all tokens, optionally filtered by category')
  .option('--category <category>', 'Filter by token category (color, typography, spacing)')
  .action((cmdOpts) => {
    const globalOpts = program.opts();
    const store = loadTokenStore(globalOpts.source, globalOpts.json);
    listCommand(store, {
      category: cmdOpts.category,
      json: globalOpts.json,
    });
  });

// ── tokens command ──────────────────────────────────────────────────
program
  .command('tokens')
  .description('Query tokens with filters, output as DTCG JSON')
  .option('--category <category>', 'Filter by token category (color, typography, spacing)')
  .option('--family <family>', 'Filter by token family (e.g., red, gray)')
  .action((cmdOpts) => {
    const globalOpts = program.opts();
    const store = loadTokenStore(globalOpts.source, globalOpts.json);
    tokensCommand(store, {
      category: cmdOpts.category,
      family: cmdOpts.family,
      json: globalOpts.json,
    });
  });

// ── export command ──────────────────────────────────────────────────
program
  .command('export')
  .description('Export tokens as DTCG JSON to stdout or file')
  .option('--out <filepath>', 'Write output to file instead of stdout')
  .option('--category <category>', 'Filter by token category (color, typography, spacing)')
  .action((cmdOpts) => {
    const globalOpts = program.opts();
    const store = loadTokenStore(globalOpts.source, globalOpts.json);
    exportCommand(store, {
      out: cmdOpts.out,
      category: cmdOpts.category,
      json: globalOpts.json,
    });
  });

// ── mcp-schema command ──────────────────────────────────────────────
program
  .command('mcp-schema')
  .description('Output MCP tool schema definition as JSON')
  .action(() => {
    mcpSchemaCommand();
  });

program.parse();
