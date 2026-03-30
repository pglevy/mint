#!/usr/bin/env node
import { Command } from 'commander';
import { TokenStore, TOKENS_URL } from './store/token-store.js';
import { listCommand } from './commands/list.js';
import { tokensCommand } from './commands/tokens.js';
import { exportCommand } from './commands/export.js';
import { mcpSchemaCommand } from './commands/mcp-schema.js';
import type { DTCGGroup } from './types.js';

async function loadStore(jsonMode: boolean): Promise<TokenStore> {
  let res: Response;
  try {
    res = await fetch(TOKENS_URL);
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    process.stderr.write(`Error fetching tokens: ${message}\n`);
    if (jsonMode) process.stdout.write(JSON.stringify({ error: message }) + '\n');
    process.exit(1);
  }

  if (!res.ok) {
    const message = `HTTP ${res.status} fetching ${TOKENS_URL}`;
    process.stderr.write(`${message}\n`);
    if (jsonMode) process.stdout.write(JSON.stringify({ error: message }) + '\n');
    process.exit(1);
  }

  const json = (await res.json()) as Record<string, DTCGGroup>;
  const store = new TokenStore();
  store.loadFromJSON(json);
  return store;
}

const program = new Command();

program
  .name('mint')
  .description('Query Sailwind design tokens')
  .version('0.1.0')
  .option('--json', 'Output machine-readable JSON', false);

program
  .command('list')
  .description('List all tokens, optionally filtered by category')
  .option('--category <category>', 'Filter by token category (color, typography, spacing)')
  .action(async (cmdOpts) => {
    const store = await loadStore(program.opts().json);
    listCommand(store, { category: cmdOpts.category, json: program.opts().json });
  });

program
  .command('tokens')
  .description('Query tokens with filters, output as DTCG JSON')
  .option('--category <category>', 'Filter by token category (color, typography, spacing)')
  .option('--family <family>', 'Filter by token family (e.g., red, gray)')
  .action(async (cmdOpts) => {
    const store = await loadStore(program.opts().json);
    tokensCommand(store, { category: cmdOpts.category, family: cmdOpts.family, json: program.opts().json });
  });

program
  .command('export')
  .description('Export tokens as DTCG JSON to stdout or file')
  .option('--out <filepath>', 'Write output to file instead of stdout')
  .option('--category <category>', 'Filter by token category (color, typography, spacing)')
  .action(async (cmdOpts) => {
    const store = await loadStore(program.opts().json);
    exportCommand(store, { out: cmdOpts.out, category: cmdOpts.category, json: program.opts().json });
  });

program
  .command('mcp-schema')
  .description('Output MCP tool schema definition as JSON')
  .action(() => { mcpSchemaCommand(); });

program.parseAsync();
