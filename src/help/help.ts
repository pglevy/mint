/**
 * Help system for the Mint CLI.
 *
 * Provides concise (no-args), full (--help), and subcommand-specific help text.
 * Examples are ordered before flags per REQ-4.4.
 */

interface SubcommandInfo {
  name: string;
  description: string;
  examples: string[];
  flags: string[];
}

const SUBCOMMANDS: SubcommandInfo[] = [
  {
    name: "list",
    description: "List design tokens",
    examples: ["mint list --category color"],
    flags: ["--category <color|typography|spacing>  Filter by token category"],
  },
  {
    name: "tokens",
    description: "Output raw DTCG token data",
    examples: ["mint tokens --category typography --family heading"],
    flags: [
      "--category <color|typography|spacing>  Filter by token category",
      "--family <name>                        Filter by token family",
    ],
  },
  {
    name: "export",
    description: "Export tokens as DTCG JSON",
    examples: ["mint export --category spacing -o tokens.json"],
    flags: [
      "--category <color|typography|spacing>  Filter by token category",
      "-o, --output <path>                    Write output to a file",
    ],
  },
  {
    name: "mcp-schema",
    description: "Output MCP tool schema",
    examples: ["mint mcp-schema"],
    flags: [],
  },
];

const GLOBAL_FLAGS = [
  "-h, --help       Show help",
  "-v, --version    Show version",
  "    --json       Output as JSON",
  "-q, --quiet      Suppress informational messages",
  "    --no-color   Disable color output",
  "    --debug      Show detailed error information",
];

/**
 * Display concise help when the CLI is invoked with no arguments.
 * Must be ≤15 lines of output.
 */
export function showConciseHelp(): void {
  const lines = [
    "Mint — query and export Sailwind design tokens.",
    "",
    "Examples:",
    "  mint list --category color",
    "  mint export -o tokens.json",
    "",
    "Subcommands: list, tokens, export, mcp-schema",
    "",
    "Run mint --help for full usage information.",
  ];
  console.log(lines.join("\n"));
}

/**
 * Display full help when --help is passed at the top level.
 * Includes description, 3+ examples, all flags, all subcommands, and repo link.
 */
export function showFullHelp(): void {
  const lines = [
    "Mint — query and export Sailwind design tokens.",
    "",
    "Examples:",
    "  mint list                          List all tokens",
    "  mint list --category color         List tokens in a category",
    "  mint tokens --category typography  Output raw DTCG data",
    "  mint export -o tokens.json         Export tokens to a file",
    "  mint mcp-schema --json             Output MCP schema as JSON",
    "",
    "Flags:",
    ...GLOBAL_FLAGS.map((f) => `  ${f}`),
    "",
    "Subcommands:",
    ...SUBCOMMANDS.map((cmd) => `  ${cmd.name.padEnd(12)} ${cmd.description}`),
    "",
    "Repository: https://github.com/pglevy/mint",
  ];
  console.log(lines.join("\n"));
}

/**
 * Display help for a specific subcommand.
 * Includes description, example(s), and flags for that subcommand.
 */
export function showSubcommandHelp(name: string): void {
  const cmd = SUBCOMMANDS.find((c) => c.name === name);
  if (!cmd) {
    console.log(`Unknown subcommand: ${name}`);
    return;
  }

  const lines = [
    `mint ${cmd.name} — ${cmd.description}`,
    "",
    "Examples:",
    ...cmd.examples.map((ex) => `  ${ex}`),
  ];

  if (cmd.flags.length > 0) {
    lines.push("", "Flags:", ...cmd.flags.map((f) => `  ${f}`));
  }

  // Always show global flags that apply
  lines.push(
    "",
    "Global Flags:",
    ...GLOBAL_FLAGS.map((f) => `  ${f}`),
  );

  console.log(lines.join("\n"));
}
