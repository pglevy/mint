/**
 * Base error class for all Mint CLI errors.
 * Carries an optional suggestion string for "did you mean?" hints.
 */
export class MintError extends Error {
  constructor(message: string, public readonly suggestion?: string) {
    super(message);
    this.name = "MintError";
  }
}

/** Thrown when a network fetch fails. */
export class FetchError extends MintError {
  constructor(message: string, suggestion?: string) {
    super(message, suggestion);
    this.name = "FetchError";
  }
}

/** Thrown when user-supplied input fails validation. */
export class ValidationError extends MintError {
  constructor(message: string, suggestion?: string) {
    super(message, suggestion);
    this.name = "ValidationError";
  }
}

/** Thrown when writing to the filesystem fails. */
export class FileWriteError extends MintError {
  constructor(message: string, suggestion?: string) {
    super(message, suggestion);
    this.name = "FileWriteError";
  }
}

/**
 * Compute Levenshtein edit distance between two strings.
 */
function levenshtein(a: string, b: string): number {
  const m = a.length;
  const n = b.length;
  const dp: number[][] = Array.from(
    { length: m + 1 },
    () => Array(n + 1).fill(0),
  );

  for (let i = 0; i <= m; i++) dp[i][0] = i;
  for (let j = 0; j <= n; j++) dp[0][j] = j;

  for (let i = 1; i <= m; i++) {
    for (let j = 1; j <= n; j++) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1;
      dp[i][j] = Math.min(
        dp[i - 1][j] + 1,
        dp[i][j - 1] + 1,
        dp[i - 1][j - 1] + cost,
      );
    }
  }

  return dp[m][n];
}

/**
 * Suggest the closest valid command for a mistyped input.
 * Returns the closest match if its edit distance is ≤ 2, otherwise null.
 */
export function suggestCommand(
  input: string,
  valid: string[],
): string | null {
  let best: string | null = null;
  let bestDistance = Infinity;

  for (const candidate of valid) {
    const distance = levenshtein(input, candidate);
    if (distance < bestDistance) {
      bestDistance = distance;
      best = candidate;
    }
  }

  return bestDistance <= 2 ? best : null;
}
