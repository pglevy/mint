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

    // Strip the category prefix from paths (e.g., "spacing.radius.sm" → "radius.sm")
    const stripPrefix = (path: string) => {
      const dotIdx = path.indexOf(".");
      return dotIdx >= 0 ? path.slice(dotIdx + 1) : path;
    };

    // Extract the family (first segment) from a stripped path
    const getFamily = (path: string) => {
      const dotIdx = path.indexOf(".");
      return dotIdx >= 0 ? path.slice(0, dotIdx) : path;
    };

    // Extract a sortable numeric value from a token
    const numericValue = (token: typeof categoryTokens[0]): number => {
      if (typeof token.value === "number") return token.value;
      if (typeof token.value === "string") {
        const n = parseFloat(token.value);
        if (!isNaN(n)) return n;
      }
      if (typeof token.value === "object" && token.value !== null) {
        // For composite tokens, try the "value" key if present
        const v = (token.value as Record<string, unknown>)["value"];
        if (typeof v === "number") return v;
        if (typeof v === "string") {
          const n = parseFloat(v);
          if (!isNaN(n)) return n;
        }
      }
      return Infinity;
    };

    // Sort: primary by family (alpha), secondary by numeric value (low to high)
    const sorted = [...categoryTokens].sort((a, b) => {
      const famA = getFamily(stripPrefix(a.path));
      const famB = getFamily(stripPrefix(b.path));
      const famCmp = famA.localeCompare(famB);
      if (famCmp !== 0) return famCmp;

      return numericValue(a) - numericValue(b);
    });

    // Calculate column width for alignment
    const strippedPaths = sorted.map((t) => stripPrefix(t.path));
    const maxPath = Math.max(...strippedPaths.map((p) => p.length));

    for (let i = 0; i < sorted.length; i++) {
      // Add blank line between families
      if (i > 0) {
        const prevFamily = getFamily(strippedPaths[i - 1]);
        const currFamily = getFamily(strippedPaths[i]);
        if (prevFamily !== currFamily) {
          writer.stdout("");
        }
      }
      const formattedValue = formatValue(sorted[i].value);
      const paddedPath = strippedPaths[i].padEnd(maxPath);
      writer.stdout(`    ${paddedPath}  ${formattedValue}`);
    }
  }
}
