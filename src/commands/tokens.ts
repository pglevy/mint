import type { TokenStore } from '../store/token-store.js';
import type { TokenCategory } from '../types.js';
import { VALID_CATEGORIES } from '../types.js';

export interface TokensCommandOptions {
  category?: string;
  family?: string;
  json?: boolean;
}

export function tokensCommand(store: TokenStore, options: TokensCommandOptions): void {
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
  let result: Record<string, unknown>;

  if (category && options.family) {
    // 2. Both --category and --family provided
    const familyTokens = store.getByFamily(category, options.family);
    result = { [category]: { [options.family]: familyTokens } };
  } else if (category) {
    // 3. Only --category provided
    const categoryTokens = store.getByCategory(category);
    result = { [category]: categoryTokens };
  } else {
    // 4. No filters — all tokens
    result = store.getAllTokens();
  }

  // 5. Output as formatted JSON
  process.stdout.write(JSON.stringify(result, null, 2) + '\n');
}
