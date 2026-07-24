import type { DTCGGroup } from "../types.ts";
import { FetchError } from "../errors.ts";

/** Default URL for the Sailwind design token source. */
export const TOKENS_URL =
  "https://raw.githubusercontent.com/pglevy/sailwind/main/public/tokens.json";

/**
 * Fetch and parse design tokens from a remote URL.
 * Throws `FetchError` on network failure or non-OK response.
 */
export async function fetchTokens(
  url: string,
): Promise<Record<string, DTCGGroup>> {
  let response: Response;
  try {
    response = await fetch(url);
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : String(err);
    throw new FetchError(
      `Failed to fetch tokens from ${url}: ${message}`,
      "Check your network connection and try again.",
    );
  }

  if (!response.ok) {
    throw new FetchError(
      `Failed to fetch tokens from ${url}: HTTP ${response.status} ${response.statusText}`,
      "Check your network connection and try again.",
    );
  }

  return await response.json() as Record<string, DTCGGroup>;
}
