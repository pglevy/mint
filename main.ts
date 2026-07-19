/**
 * Mint CLI entry point.
 *
 * Parses global arguments, routes to subcommand handlers,
 * and provides consistent error formatting.
 */

import { parseArgs } from "@std/cli";
import {
  showConciseHelp,
  showFullHelp,
  showSubcommandHelp,
} from "./src/help/help.ts";
import { MintError, suggestCommand } from "./src/errors.ts";
import { createColorContext } from "./src/output/color.ts";
import { createWriter } from "./src/output/writer.ts";
import { exportCommand } from "./src/commands/export.ts";
import { tokensCommand } from "./src/commands/tokens.ts";
import { listCommand } from "./src/commands/list.ts";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface GlobalOptions {
  json: boolean;
  quiet: boolean;
  noColor: boolean;
  debug: boolean;
  help: boolean;
  version: boolean;
}

type SubcommandHandler = (
  args: string[],
  globals: GlobalOptions,
) => Promise<void>;

// ---------------------------------------------------------------------------
// Subcommand registry
// ---------------------------------------------------------------------------

const VALID_SUBCOMMANDS = ["list", "tokens", "export"];

const handlers: Record<string, SubcommandHandler> = {
  list: listCommand,
  tokens: tokensCommand,
  export: exportCommand,
};

// ---------------------------------------------------------------------------
// Version
// ---------------------------------------------------------------------------

const VERSION = "0.2.0";

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

async function main(): Promise<void> {
  const parsed = parseArgs(Deno.args, {
    boolean: ["help", "version", "json", "quiet", "no-color", "debug"],
    alias: {
      h: "help",
      v: "version",
      q: "quiet",
    },
    stopEarly: false,
  });

  const globals: GlobalOptions = {
    help: Boolean(parsed.help),
    version: Boolean(parsed.version),
    json: Boolean(parsed.json),
    quiet: Boolean(parsed.quiet),
    noColor: Boolean(parsed["no-color"]),
    debug: Boolean(parsed.debug),
  };

  const color = createColorContext({ noColor: globals.noColor });
  const writer = createWriter(
    { quiet: globals.quiet, json: globals.json },
    color,
  );

  // Handle --version
  if (globals.version) {
    writer.stdout(VERSION);
    Deno.exit(0);
  }

  const positionals = parsed._ as string[];

  // Handle no-args → concise help
  if (positionals.length === 0 && !globals.help) {
    showConciseHelp();
    Deno.exit(0);
  }

  // Handle --help at top level (no subcommand)
  if (positionals.length === 0 && globals.help) {
    showFullHelp();
    Deno.exit(0);
  }

  // Route to subcommand
  const subcommand = String(positionals[0]);

  // Extract raw args after the subcommand name for the subcommand handler
  const subcommandIndex = Deno.args.indexOf(subcommand);
  const subArgs = subcommandIndex >= 0
    ? Deno.args.slice(subcommandIndex + 1).filter((a) =>
      // Filter out global flags that were already consumed
      ![
        "--json",
        "--quiet",
        "-q",
        "--no-color",
        "--debug",
        "--help",
        "-h",
        "--version",
        "-v",
      ].includes(a)
    )
    : [];

  // Check if subcommand is valid
  if (!VALID_SUBCOMMANDS.includes(subcommand)) {
    const suggestion = suggestCommand(subcommand, VALID_SUBCOMMANDS);
    let message = `Error: Unknown subcommand "${subcommand}".`;
    message += `\nValid subcommands: ${VALID_SUBCOMMANDS.join(", ")}`;
    if (suggestion) {
      message += `\n\nDid you mean "${suggestion}"?`;
    }
    writer.error(message);
    Deno.exit(1);
  }

  // Handle --help for a subcommand
  if (globals.help) {
    showSubcommandHelp(subcommand);
    Deno.exit(0);
  }

  // Execute the subcommand handler
  const handler = handlers[subcommand];
  await handler(subArgs, globals);
}

// ---------------------------------------------------------------------------
// Top-level error handling
// ---------------------------------------------------------------------------

try {
  await main();
} catch (err: unknown) {
  // Determine global options for error formatting (best-effort re-parse)
  const parsed = parseArgs(Deno.args, {
    boolean: ["json", "no-color", "debug", "quiet"],
    alias: { q: "quiet" },
    stopEarly: false,
  });

  const debug = Boolean(parsed.debug);
  const json = Boolean(parsed.json);
  const noColor = Boolean(parsed["no-color"]);
  const quiet = Boolean(parsed.quiet);

  const color = createColorContext({ noColor });
  const writer = createWriter({ quiet, json }, color);

  if (err instanceof MintError) {
    let message = err.message;
    if (err.suggestion) {
      message += `\n${color.dim(err.suggestion)}`;
    }
    writer.error(message);
    if (debug && err.stack) {
      writer.stderr(err.stack);
    }
  } else if (err instanceof Error) {
    writer.error(err.message);
    writer.stderr("Run with --debug for more details.");
    if (debug && err.stack) {
      writer.stderr(err.stack);
    }
  } else {
    writer.error("An unexpected error occurred.");
    writer.stderr("Run with --debug for more details.");
  }

  Deno.exit(1);
}
