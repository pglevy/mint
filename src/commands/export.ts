import fs from 'node:fs';
import type { TokenStore } from '../store/token-store.js';
import type { TokenCategory } from '../types.js';
import { VALID_CATEGORIES } from '../types.js';

export interface ExportCommandOptions {
  out?: string;
  category?: string;
  json?: boolean;
}

export function exportCommand(store: TokenStore, options: ExportCommandOptions): void {
  // 1. Validate category if provided
  if (options.category) {
    if (!VALID_CATEGORIES.includes(options.category as TokenCategory)) {
      const msg = `Error: Unknown category "${options.category}". Valid categories: ${VALID_CATEGORIES.join(', ')}`;
      process.stderr.write(msg + '\n');
      if (options.json) {
        process.stdout.write(JSON.stringify({ error: msg }) + '\n');
      }
      process.exit(1);
    }
  }

  const category = options.category as TokenCategory | undefined;

  // 2. Get JSON output
  const jsonOutput = store.toJSON(category);

  // 3. Write to file or stdout
  if (options.out) {
    try {
      fs.writeFileSync(options.out, jsonOutput + '\n', 'utf-8');
      process.stderr.write(`Exported to ${options.out}\n`);
    } catch {
      const msg = `Error: Cannot write to ${options.out}`;
      process.stderr.write(msg + '\n');
      if (options.json) {
        process.stdout.write(JSON.stringify({ error: msg }) + '\n');
      }
      process.exit(1);
    }
  } else {
    process.stdout.write(jsonOutput + '\n');
  }
}
