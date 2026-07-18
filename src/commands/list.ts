import { parseArgs } from "@std/cli";
import { type TokenCategory, VALID_CATEGORIES } from "../types.ts";
import { ValidationError } from "../errors.ts";
import { TokenStore } from "../store/token-store.ts";
import { fetchTokens, TOKENS_URL } from "../store/fetcher.ts";
import { formatValue } from "../output/format-value.ts";
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

/**
 * The `list` command: displays all tokens in human-readable or JSON format,
 * optionally filtered by category.
 */
export async function listCommand(
  args: string[],
  globals: GlobalOptions,
): Promise<void> {
  const color = createColorContext({ noColor: globals.noColor });
  const writer = createWriter(
    { quiet: globals.quiet, json: globals.json },
    color,
  );

  // Parse subcommand-specific flags
  const parsed = parseArgs(args, {
    string: ["category"],
    alias: { c: "category" },
  });

  let category = parsed.category;

  // Validate --category if provided
  if (category !== undefined) {
    if (!VALID_CATEGORIES.includes(category as TokenCategory)) {
      throw new ValidationError(
        `Unknown category "${category}".`,
        `Valid categories: ${VALID_CATEGORIES.join(", ")}`,
      );
    }
  }

  // Interactive prompting: prompt for category when missing
  if (category === undefined) {
    if (Deno.stdin.isTerminal()) {
      const selected = await promptSelect({
        message: "Select a category:",
        choices: [...VALID_CATEGORIES],
      });
      if (selected === null) {
        Deno.exit(1);
      }
      category = selected;
    } else {
      writer.stderr(
        "Error: --category is required when stdin is not a TTY",
      );
      Deno.exit(1);
    }
  }

  const validatedCategory = category as TokenCategory | undefined;

  // Fetch tokens and load into store
  const json = await fetchTokens(TOKENS_URL);
  const store = new TokenStore();
  store.loadFromJSON(json);

  // Get token entries
  const tokens = store.listTokens(validatedCategory);

  // JSON mode: output token array
  if (globals.json) {
    writer.json(tokens);
    return;
  }

  // Human-readable mode: group by category with colored headers and aligned columns
  const grouped = new Map<TokenCategory, typeof tokens>();
  for (const token of tokens) {
    const list = grouped.get(token.category) ?? [];
    list.push(token);
    grouped.set(token.category, list);
  }

  for (const [cat, categoryTokens] of grouped) {
    // Show category header unless --quiet
    if (!globals.quiet) {
      const header = color.category(
        `${cat} (${categoryTokens.length} tokens)`,
        cat,
      );
      writer.stdout("");
      writer.stdout(`  ${header}`);
    }

    // Calculate column widths for alignment
    const pathWidths = categoryTokens.map((t) => t.path.length);
    const valueWidths = categoryTokens.map((t) => formatValue(t.value).length);
    const maxPath = Math.max(...pathWidths);
    const maxValue = Math.max(...valueWidths);

    for (const token of categoryTokens) {
      const formattedValue = formatValue(token.value);
      const paddedPath = token.path.padEnd(maxPath);
      const paddedValue = formattedValue.padEnd(maxValue);
      writer.stdout(`    ${paddedPath}  ${paddedValue}  ${token.type}`);
    }
  }
}
