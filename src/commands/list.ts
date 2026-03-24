import type { TokenStore } from '../store/token-store.js';
import type { TokenCategory } from '../types.js';
import { VALID_CATEGORIES } from '../types.js';

export interface ListCommandOptions {
  category?: string;
  json?: boolean;
}

export function listCommand(store: TokenStore, options: ListCommandOptions): void {
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

  const validatedCategory = options.category as TokenCategory | undefined;

  // 2. Get flat token list
  const tokens = store.listTokens(validatedCategory);

  // 3. JSON mode — output as JSON array
  if (options.json) {
    process.stdout.write(JSON.stringify(tokens, null, 2) + '\n');
    return;
  }

  // 4. Human-readable mode — group by category
  const grouped = new Map<TokenCategory, typeof tokens>();
  for (const token of tokens) {
    const list = grouped.get(token.category) ?? [];
    list.push(token);
    grouped.set(token.category, list);
  }

  for (const [category, categoryTokens] of grouped) {
    process.stdout.write(`\n  ${category} (${categoryTokens.length} tokens):\n`);
    for (const token of categoryTokens) {
      process.stdout.write(`    ${token.path}  ${token.value}  (${token.type})\n`);
    }
  }
}
