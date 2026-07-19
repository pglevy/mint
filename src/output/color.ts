import type { TokenCategory } from "../types.ts";

/**
 * ANSI escape codes for terminal colors.
 */
const ANSI = {
  reset: "\x1b[0m",
  bold: "\x1b[1m",
  dim: "\x1b[2m",
  red: "\x1b[31m",
  green: "\x1b[32m",
  cyan: "\x1b[36m",
  magenta: "\x1b[35m",
  yellow: "\x1b[33m",
} as const;

/**
 * Maps token categories to their display colors.
 */
const CATEGORY_COLORS: Record<TokenCategory, string> = {
  color: ANSI.cyan,
  gradient: ANSI.green,
  typography: ANSI.magenta,
  spacing: ANSI.yellow,
};

/**
 * Provides color formatting methods that respect the enabled/disabled state.
 * When disabled, all methods return text unchanged (no ANSI codes).
 */
export interface ColorContext {
  enabled: boolean;
  category(text: string, cat: TokenCategory): string;
  error(text: string): string;
  dim(text: string): string;
  bold(text: string): string;
}

/**
 * Create a ColorContext that determines color support from:
 * - TTY status (Deno.stdout.isTerminal())
 * - NO_COLOR environment variable
 * - --no-color flag
 */
export function createColorContext(
  options: { noColor: boolean },
): ColorContext {
  const enabled = Deno.stdout.isTerminal() &&
    !Deno.env.get("NO_COLOR") &&
    !options.noColor;

  if (!enabled) {
    return {
      enabled: false,
      category: (text: string, _cat: TokenCategory) => text,
      error: (text: string) => text,
      dim: (text: string) => text,
      bold: (text: string) => text,
    };
  }

  return {
    enabled: true,
    category: (text: string, cat: TokenCategory) =>
      `${CATEGORY_COLORS[cat]}${text}${ANSI.reset}`,
    error: (text: string) => `${ANSI.red}${text}${ANSI.reset}`,
    dim: (text: string) => `${ANSI.dim}${text}${ANSI.reset}`,
    bold: (text: string) => `${ANSI.bold}${text}${ANSI.reset}`,
  };
}
