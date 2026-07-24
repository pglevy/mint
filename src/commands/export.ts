import { parseArgs } from "@std/cli";
import { type TokenCategory, VALID_CATEGORIES } from "../types.ts";
import { FileWriteError, ValidationError } from "../errors.ts";
import { TokenStore } from "../store/token-store.ts";
import { fetchTokens, TOKENS_URL } from "../store/fetcher.ts";
import { createColorContext } from "../output/color.ts";
import { createWriter } from "../output/writer.ts";

interface GlobalOptions {
  json: boolean;
  quiet: boolean;
  noColor: boolean;
  debug: boolean;
  help: boolean;
  version: boolean;
}

/**
 * The `export` command: generates DTCG JSON and writes to stdout or a file.
 * Validates --category, fetches tokens, and outputs the result.
 */
export async function exportCommand(
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
    string: ["category", "output"],
    alias: { c: "category", o: "output" },
  });

  const category = parsed.category;
  const outputPath = parsed.output;

  // Validate --category if provided
  if (category !== undefined) {
    if (!VALID_CATEGORIES.includes(category as TokenCategory)) {
      throw new ValidationError(
        `Unknown category "${category}".`,
        `Valid categories: ${VALID_CATEGORIES.join(", ")}`,
      );
    }
  }

  const validatedCategory = category as TokenCategory | undefined;

  // Fetch tokens and load into store
  const json = await fetchTokens(TOKENS_URL);
  const store = new TokenStore();
  store.loadFromJSON(json);

  // Generate DTCG JSON
  const dtcgJson = store.toJSON(validatedCategory);

  // Output to file or stdout
  if (outputPath) {
    try {
      await Deno.writeTextFile(outputPath, dtcgJson + "\n");
    } catch (_err: unknown) {
      throw new FileWriteError(
        `Cannot write to "${outputPath}".`,
        "Check that the directory exists and you have write permissions.",
      );
    }
    // Confirm on stderr unless --quiet
    writer.stderr(`Wrote tokens to ${outputPath}`);
  } else {
    writer.stdout(dtcgJson);
  }
}
