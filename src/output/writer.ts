import type { ColorContext } from "./color.ts";

/**
 * Centralized output writer that enforces output routing rules:
 * - Primary data goes to stdout
 * - Informational/error messages go to stderr
 * - --quiet suppresses stderr
 * - --json mode outputs structured JSON
 */
export interface Writer {
  stdout(text: string): void;
  stderr(text: string): void;
  json(data: unknown): void;
  error(message: string, detail?: string): void;
}

/**
 * Create a Writer instance configured with quiet/json options and color context.
 */
export function createWriter(
  options: { quiet: boolean; json: boolean },
  color: ColorContext,
): Writer {
  const encoder = new TextEncoder();

  return {
    stdout(text: string): void {
      Deno.stdout.writeSync(encoder.encode(text + "\n"));
    },

    stderr(text: string): void {
      if (options.quiet) return;
      Deno.stderr.writeSync(encoder.encode(text + "\n"));
    },

    json(data: unknown): void {
      const output = JSON.stringify(data, null, 2);
      Deno.stdout.writeSync(encoder.encode(output + "\n"));
    },

    error(message: string, detail?: string): void {
      const formatted = color.error(message);
      Deno.stderr.writeSync(encoder.encode(formatted + "\n"));

      if (detail) {
        Deno.stderr.writeSync(encoder.encode(detail + "\n"));
      }

      // In JSON mode, also write structured error to stdout
      if (options.json) {
        const errorJson = JSON.stringify({ error: message }, null, 2);
        Deno.stdout.writeSync(encoder.encode(errorJson + "\n"));
      }
    },
  };
}
