/**
 * `tokens` subcommand — outputs raw DTCG token data as formatted JSON.
 *
 * Supports filtering by --category and --family.
 * Always outputs JSON (this is the "raw data" command).
 */

import { parseArgs } from "@std/cli";
import { type TokenCategory, VALID_CATEGORIES } from "../types.ts";
import { ValidationError } from "../errors.ts";
import { TokenStore } from "../store/token-store.ts";
import { fetchTokens, TOKENS_URL } from "../store/fetcher.ts";
import { createColorContext } from "../output/color.ts";
import { createWriter } from "../output/writer.ts";
import { promptSelect } from "../prompt/select.ts";

interface GlobalOptions {
  json: boolean;
  quiet: boolean;
  noColor: boolean;
  debug: boolean;
  help: boolean;
  version: boolean;
}

export async function tokensCommand(
  args: string[],
  globals: GlobalOptions,
): Promise<void> {
  const color = createColorContext({ noColor: globals.noColor });
  const writer = createWriter(
    { quiet: globals.quiet, json: globals.json },
    color,
  );

  const parsed = parseArgs(args, {
    string: ["category", "family"],
    alias: { c: "category", f: "family" },
  });

  let category = parsed.category;
  let family = parsed.family;

  // Validate --category if present
  if (category !== undefined) {
    if (!VALID_CATEGORIES.includes(category as TokenCategory)) {
      throw new ValidationError(
        `Unknown category "${category}". Valid categories: ${
          VALID_CATEGORIES.join(", ")
        }`,
      );
    }
  }

  // Interactive prompting for --category when missing
  if (category === undefined) {
    const isInteractive = Deno.stdin.isTerminal() &&
      Deno.stdout.isTerminal();

    if (isInteractive) {
      const selected = await promptSelect({
        message: "Select a category:",
        choices: [...VALID_CATEGORIES],
      });
      if (selected === null) {
        Deno.exit(1);
      }
      category = selected;
    } else if (!Deno.stdout.isTerminal()) {
      // Output is piped/redirected — dump all categories
      category = undefined;
    } else {
      writer.stderr(
        "Error: --category is required when stdin is not a TTY",
      );
      Deno.exit(1);
    }
  }

  // --family requires --category (already guaranteed at this point)
  // Validate --family if provided explicitly
  // (no validation needed since family names come from the store)

  // Fetch tokens and load into store
  const json = await fetchTokens(TOKENS_URL);
  const store = new TokenStore();
  store.loadFromJSON(json);

  // Interactive prompting for --family when category is known but family is missing
  if (family === undefined && category !== undefined) {
    const families = store.getFamilies(category as TokenCategory);
    const isInteractive = Deno.stdin.isTerminal() &&
      Deno.stdout.isTerminal();
    if (families.length > 0 && isInteractive) {
      const selected = await promptSelect({
        message: `Select a family in "${category}":`,
        choices: families,
      });
      if (selected === null) {
        Deno.exit(1);
      }
      family = selected;
    }
    // If not interactive and family is missing, we don't error — family is optional
  }

  // Query store based on provided filters
  let result: Record<string, unknown>;

  if (category && family) {
    result = store.getByFamily(category as TokenCategory, family);
  } else if (category) {
    result = store.getByCategory(category as TokenCategory);
  } else {
    result = store.getAllTokens();
  }

  // Output as formatted DTCG JSON
  writer.stdout(JSON.stringify(result, null, 2));
}
